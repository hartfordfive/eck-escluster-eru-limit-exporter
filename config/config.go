package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"

	units "github.com/docker/go-units"
	"gopkg.in/yaml.v2"
)

var GlobalConfig *Config

// Config is the top level configuration used by the service discovery module
type Config struct {
	ListenIP    string            `yaml:"listen_ip" json:"listen_ip"`
	ListenPort  int               `yaml:"listen_port" json:"listen_port"`
	MetricsPath string            `yaml:"metrics_path" json:"metrics_path"`
	EruSize     string            `yaml:"eru_size" json:"eru_size"`
	Clusters    map[string]string `yaml:"clusters" json:"clusters"`
}

// NewConfig constructs a new Config instance
func NewConfig(configPath string) (*Config, error) {
	c := &Config{}
	b, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("Could not read config: %s", err)
	}
	err = yaml.Unmarshal([]byte(b), &c)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal config: %v", err)
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

// Serialize serializes the configuration so that it can be printed and viewed
func (c *Config) Serialize() (string, error) {
	if b, err := yaml.Marshal(c); err != nil {
		return "", err
	} else {
		return string(b), nil
	}
}

// Validate runs a validation on the configuration file
func (c *Config) Validate() error {
	if c.ListenIP == "" {
		c.ListenIP = "127.0.0.1" // default value
	} else if net.ParseIP(c.ListenIP) == nil {
		return errors.New("Invalid IP for listen_addr")
	}
	if c.ListenPort == 0 {
		c.ListenPort = 8889 // default value
	}
	if c.ListenPort < 1 || c.ListenPort > 65535 {
		return errors.New("Value 'listen_port' must be between 1 and 65535")
	}
	if c.MetricsPath == "" {
		c.MetricsPath = "/metrics"
	} else if _, err := url.ParseRequestURI(c.MetricsPath); err != nil {
		return errors.New("Invalid value specified for metrics_path")
	}
	if c.EruSize == "" {
		c.EruSize = "64Gb"
	} else if _, err := units.FromHumanSize(c.EruSize); err != nil {
		return errors.New("Invalid memory unit specified for eru_size")
	}
	if c.Clusters == nil || len(c.Clusters) < 1 {
		return errors.New("Must have at least one cluster defined")
	}
	return nil
}
