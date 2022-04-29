package config

import (
	"errors"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/klog/v2"
	"testing"
)

func TestConfig(t *testing.T) {
	config := Config{}
	if err := config.testParse(); err != nil {
		t.Log(err)
		t.FailNow()
	}
	// You can add assert.Equal to prove that your config is correct
	assert.Equal(t, "mqtts://127.0.0.1:8883", config.Mqtt.ServerAddress)
	assert.Equal(t, "../configmap_test.json", config.Configmap)
}

// testParse used to test Parse ,but it uses config_test.yaml
func (c *Config) testParse() error {
	var level klog.Level
	var loglevel string
	var configFile string
	// -config-file /home/xxx
	defaultConfigFile = "config_test.yaml"
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
