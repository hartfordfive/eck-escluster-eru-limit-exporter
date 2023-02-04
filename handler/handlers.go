package handler

import (
	"fmt"
	"net/http"

	"github.com/hartfordfive/eck-escluster-eru-limit-exporter/config"
	"github.com/hartfordfive/eck-escluster-eru-limit-exporter/logger"
)

var HealthHandler = func(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, "OK")
}

var ShowConfigHandler = func(w http.ResponseWriter, req *http.Request) {
	logger.Logger.Info("Debug config requested")
	w.Header().Set("Content-Type", "text/yaml")
	printCnf, err := config.GlobalConfig.Serialize()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s\n", printCnf)
}
