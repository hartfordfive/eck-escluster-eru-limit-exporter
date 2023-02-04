package main

import (
	"context"
	"flag"
	"fmt"
	//defaultLogger "log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	units "github.com/docker/go-units"
	"github.com/gorilla/mux"
	"github.com/hartfordfive/eck-escluster-eru-limit-exporter/config"
	"github.com/hartfordfive/eck-escluster-eru-limit-exporter/logger"
	"github.com/hartfordfive/eck-escluster-eru-limit-exporter/metrics"
	"github.com/hartfordfive/eck-escluster-eru-limit-exporter/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const binaryName = "eck-escluster-eru-limit-exporter"

var (
	flagConfPath       *string
	flagDebug          *bool
	flagVersion        *bool
	flagValidateConfig *bool
	log                *zap.Logger
	conf               *config.Config
)

func init() {

	flagConfPath = flag.String(
		"conf",
		fmt.Sprintf("/etc/%s/exporter.conf", binaryName),
		"Path to the configuration file",
	)
	flagVersion = flag.Bool("version", false, "Show version and exit")
	flagDebug = flag.Bool("debug", false, "Enable debug mode")
	flagValidateConfig = flag.Bool("validate", false, "Validate config and exit")
	flag.Parse()

	prometheus.Register(metrics.MetricBuildInfo)
	prometheus.Register(metrics.MetricClusterEruLimit)
	prometheus.Register(metrics.MetricEruSize)

	var log *zap.Logger
	var loggerErr error

	atom := zap.NewAtomicLevel()

	log = zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.Lock(os.Stdout),
		atom,
	))
	defer log.Sync()

	if *flagDebug {
		atom.SetLevel(zap.DebugLevel)
	}

	if loggerErr != nil {
		fmt.Printf("Could not initialize logger: %s\n", loggerErr)
		os.Exit(1)
	}
	logger.Logger = log
	defer logger.Logger.Sync()

	if *flagVersion {
		fmt.Printf("%s %s (Git hash: %s)\n", binaryName, version.Version, version.CommitHash)
		os.Exit(0)
	}

	cnf, err := config.NewConfig(*flagConfPath)

	if err != nil {
		logger.Logger.Error(fmt.Sprintf("%s", err))
		os.Exit(1)
	}
	conf = cnf

}

// prometheusMiddleware implements mux.MiddlewareFunc.
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := metrics.NewResponseWriter(w)
		next.ServeHTTP(rw, r)
	})
}

func validateConfig(cnf *config.Config) int {
	if cnf.Validate() != nil {
		return 1
	}
	return 0
}

func main() {

	if *flagValidateConfig {
		logger.Logger.Info("Validating configuration", zap.String("path", *flagConfPath))
		res := validateConfig(conf)
		if res == 0 {
			logger.Logger.Info("Configuration OK")
		}
		os.Exit(res)
	}

	logger.Logger.Info("Starting server")

	config.GlobalConfig = conf

	metrics.MetricBuildInfo.WithLabelValues(version.Version, version.CommitHash).Inc()
	eruSize, _ := units.FromHumanSize(conf.EruSize)
	metrics.MetricEruSize.Set(float64(eruSize))

	for cluster, maxMem := range conf.Clusters {
		maxMemBytes, _ := units.FromHumanSize(maxMem)
		metrics.MetricClusterEruLimit.WithLabelValues(cluster).Set(float64(maxMemBytes))
	}

	listenAddr := fmt.Sprintf("%s:%d", conf.ListenIP, conf.ListenPort)

	r := mux.NewRouter()
	srv := &http.Server{
		Handler:      r,
		Addr:         listenAddr,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)

	r.HandleFunc("/config", func(w http.ResponseWriter, req *http.Request) {
		logger.Logger.Info("Debug config requested")
		w.Header().Set("Content-Type", "text/yaml")
		printCnf, err := config.GlobalConfig.Serialize()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%s\n", printCnf)
	}).Methods("GET")

	r.HandleFunc("/cluster-limit", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cluster := req.URL.Query().Get("cluster")
		// In this method, we should return the ERU limit for the given
		if _, ok := conf.Clusters[cluster]; !ok {
			msg := fmt.Sprintf("Cluster %s not defined", cluster)
			logger.Logger.Error(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		val := conf.Clusters[cluster]
		res, err := units.FromHumanSize(val)
		if err != nil {
			msg := fmt.Sprintf("Invalid cluster memory limit found: %s", err.Error())
			logger.Logger.Error(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%f\n", res)

	}).Methods("GET")

	r.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, "OK")
	}).Methods("GET")

	r.Handle("/metrics", promhttp.Handler()).Methods("GET")

	go func() {
		logger.Logger.Info("Starting exporter",
			zap.String("address", listenAddr),
		)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Logger.Fatal(fmt.Sprintf("Error starting web server: %v", err))
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for {
			select {
			case killSig := <-interruptChan:
				if killSig == os.Interrupt || killSig == syscall.SIGTERM {
					logger.Logger.Info("Received shutdown notification")
					wg.Done()
					return
				}
			}
		}
	}()

	wg.Wait()

	logger.Logger.Info("Shutting down")

	ctx, cancelHTTPServer := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancelHTTPServer()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		panic(err)
	}

	logger.Logger.Info("Server shutdown complete")
}
