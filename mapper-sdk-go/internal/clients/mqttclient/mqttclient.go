// Package mqttclient implements MQTT client initialization and message processing
package mqttclient

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog/v2"
)

// MqttClient is parameters for Mqtt client.
type MqttClient struct {
	Qos        byte
	Retained   bool
	IP         string
	User       string
	Passwd     string
	ClientID   string
	Cert       string
	PrivateKey string
	Client     mqtt.Client
	ServerName string
	CaCert     string
}

// newTLSConfig new TLS configuration.
// Only one side check. Mqtt broker check the cert from client.
func newTLSConfig(caCert string, certFile string, privateKey string, serverName string) (*tls.Config, error) {
	// Import client certificate/key pair
	cert, err := tls.LoadX509KeyPair(certFile, privateKey)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	crt, err := ioutil.ReadFile(caCert)
	if err != nil {
		klog.Errorf("Failed to read cert %s:%v", caCert, err)
		os.Exit(1)
	}
	pool.AppendCertsFromPEM(crt)
	// Create tls.Config with desired tls properties
	return &tls.Config{
		// ClientAuth = whether to requests cert from server.
		// Since the server is set up for SSL, this happens
		// anyways.
		ServerName: serverName,
		// ClientCAs = certs used to validate client cert.
		ClientCAs: nil,
		RootCAs:   pool,
		// InsecureSkipVerify = verify that cert contents
		// match server. IP matches what is in cert etc.
		InsecureSkipVerify: false,
		// Certificates = list of certs client sends to server.
		Certificates: []tls.Certificate{cert},
	}, nil
}

// Connect used to the Mqtt server.
func (mc *MqttClient) Connect() error {
	opts := mqtt.NewClientOptions().AddBroker(mc.IP).SetClientID(mc.ClientID).SetCleanSession(true)
	if mc.Cert != "" {
		tlsConfig, err := newTLSConfig(mc.CaCert, mc.Cert, mc.PrivateKey, mc.ServerName)
		if err != nil {
			return err
		}
		opts.SetTLSConfig(tlsConfig)
	}
	opts.SetUsername(mc.User)
	opts.SetPassword(mc.Passwd)
	mc.Client = mqtt.NewClient(opts)
	// The token is used to indicate when actions have completed.
	if tc := mc.Client.Connect(); tc.Wait() && tc.Error() != nil {
		return tc.Error()
	}
	mc.Qos = 0          // At most 1 time
	mc.Retained = false // Not retained
	return nil
}

// Publish a Mqtt message.
func (mc *MqttClient) Publish(topic string, payload interface{}) error {
	if tc := mc.Client.Publish(topic, mc.Qos, mc.Retained, payload); tc.Wait() && tc.Error() != nil {
		return tc.Error()
	}
	return nil
}

// Subscribe a Mqtt topic.
func (mc *MqttClient) Subscribe(topic string, onMessage mqtt.MessageHandler) error {
	if tc := mc.Client.Subscribe(topic, mc.Qos, onMessage); tc.Wait() && tc.Error() != nil {
		return tc.Error()
	}
	return nil
}
