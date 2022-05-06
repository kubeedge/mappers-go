// Package config used to parse config.yaml and cmd param
package config

import (
	"errors"
	"io/ioutil"
	"runtime"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

// Config is the modbus mapper configuration.
type Config struct {
	Mqtt      Mqtt   `yaml:"mqtt,omitempty"`
	HTTP     HTTP   `yaml:"http,omitempty"`
	Configmap string `yaml:"configmap"`
}

// Mqtt is the Mqtt configuration.
type Mqtt struct {
	ServerAddress string `yaml:"server,omitempty"`
	ServerName    string `yaml:"servername"`
	Username      string `yaml:"username,omitempty"`
	Password      string `yaml:"password,omitempty"`
	ClientID	  string  `yaml:"clientId"`
	Cert          string `yaml:"certification,omitempty"`
	PrivateKey    string `yaml:"privatekey,omitempty"`
	CaCert        string `yaml:"caCert,omitempty"`
}

// HTTP is the HTTP configuration
type HTTP struct {
	CaCert     string `yaml:"caCert,omitempty"`
	Cert       string `yaml:"certification,omitempty"`
	PrivateKey string `yaml:"privatekey,omitempty"`
}

// ErrConfigCert error of certification configuration.
var ErrConfigCert = errors.New("both certification and private key must be provided")

// Command line parameters default config.yaml path
var defaultConfigFile string

// Parse to parse the configuration file. If failed, return error.
func (c *Config) Parse() error {
	var level klog.Level
	var loglevel string
	var configFile string
	// -config-file /home/xxx
	sysType := runtime.GOOS
	if sysType == "linux" {
		defaultConfigFile = "../res/config.yaml"
	} else if sysType == "windows" {
		defaultConfigFile = "..\\res\\config.yaml"
	}
	pflag.StringVar(&loglevel, "v", "1", "log level")
	pflag.StringVar(&configFile, "config-file", defaultConfigFile, "Config file name")
	pflag.StringVar(&c.Mqtt.ServerAddress, "mqtt-address", c.Mqtt.ServerAddress, "MQTT broker address")
	pflag.StringVar(&c.Mqtt.Username, "mqtt-username", c.Mqtt.Username, "username")
	pflag.StringVar(&c.Mqtt.Password, "mqtt-password", c.Mqtt.Password, "password")
	pflag.StringVar(&c.Mqtt.Cert, "mqtt-certification", c.Mqtt.Cert, "certification file path")
	pflag.StringVar(&c.Mqtt.PrivateKey, "mqtt-privatekey", c.Mqtt.PrivateKey, "private key file path")
	pflag.Parse()
	cf, err := ioutil.ReadFile(configFile)
	if err != nil {
		return errors.New("config.yaml not found," + err.Error())
	}
	if err = yaml.Unmarshal(cf, c); err != nil {
		return errors.New("yaml.Unmarshal error:," + err.Error())
	}
	if err = level.Set(loglevel); err != nil {
		return errors.New("set loglevel error:," + err.Error())
	}
	if c.Mqtt.Cert != "" && c.Mqtt.PrivateKey == "" {
		klog.V(1).Info("The PrivateKey path is empty,", ErrConfigCert.Error())
	} else if c.Mqtt.Cert == "" && c.Mqtt.PrivateKey != "" {
		klog.V(1).Info("The CertPath is empty,", ErrConfigCert.Error())
	} else if c.Mqtt.Cert == "" && c.Mqtt.PrivateKey == "" {
		klog.V(1).Info("The connection is not secure,if you want to be secure,", ErrConfigCert.Error())
	}
	return nil
}
