// Package httpclient implements HTTP client initialization and message processing
package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/config"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/internal/httpadapter"
	"github.com/kubeedge/mappers-go/mapper-sdk-go/pkg/di"
)

// HTTPClient is structure used to init HttpClient
type HTTPClient struct {
	IP             string
	Port           string
	WriteTimeout   time.Duration
	ReadTimeout    time.Duration
	server         *http.Server
	restController *httpadapter.RestController
}

// NewHTTPClient initializes a new Http client instance
func NewHTTPClient(dic *di.Container) *HTTPClient {
	return &HTTPClient{
		IP:             "0.0.0.0",
		Port:           "1215",
		WriteTimeout:   10 * time.Second,
		ReadTimeout:    10 * time.Second,
		restController: httpadapter.NewRestController(mux.NewRouter(), dic),
	}
}

// Init is a method to construct HTTP server
func (hc *HTTPClient) Init(c config.Config) error {
	hc.restController.InitRestRoutes()
	if c.HTTP.CaCert == "" {
		hc.server = &http.Server{
			Addr:         hc.IP + ":" + hc.Port,
			WriteTimeout: hc.WriteTimeout,
			ReadTimeout:  hc.ReadTimeout,
			Handler:      hc.restController.Router,
		}
	} else {
		// Enable two-way authentication http tls
		caCrtPath := c.HTTP.CaCert
		pool := x509.NewCertPool()
		crt, err := ioutil.ReadFile(caCrtPath)
		if err != nil {
			klog.Errorf("Failed to read cert %s:%v", caCrtPath, err)
			return err
		}
		pool.AppendCertsFromPEM(crt)
		hc.server = &http.Server{
			Addr:         hc.IP + ":" + hc.Port,
			WriteTimeout: hc.WriteTimeout,
			ReadTimeout:  hc.ReadTimeout,
			Handler:      hc.restController.Router,
			TLSConfig: &tls.Config{
				ClientCAs: pool,
				// check client certificate file
				ClientAuth: tls.RequireAndVerifyClientCert,
			},
		}
	}
	klog.V(1).Info("HttpServer Start......")
	go func() {
		_, err := hc.Receive(c)
		if err != nil {
			klog.Errorf("Http Receive error:%v", err)
		}
	}()
	return nil
}

// UnInit is a method to close http server
func (hc *HTTPClient) UnInit() {
	err := hc.server.Close()
	if err != nil {
		klog.Error("Http server close err:", err.Error())
		return
	}
}

// Send no messages need to be sent
func (hc *HTTPClient) Send(message interface{}) error {
	return nil
}

// Receive http server start listen
func (hc *HTTPClient) Receive(c config.Config) (interface{}, error) {
	if c.HTTP.CaCert == "" && c.HTTP.Cert == "" && c.HTTP.PrivateKey == "" {
		err := hc.server.ListenAndServe()
		if err != nil {
			return nil, err
		}
	} else if c.HTTP.Cert != "" && c.HTTP.PrivateKey != "" {
		serverCrtPath := c.HTTP.Cert
		serverKeyPath := c.HTTP.PrivateKey
		err := hc.server.ListenAndServeTLS(serverCrtPath,
			serverKeyPath)
		if err != nil {
			klog.Error("HTTP Server exited...")
			return "", err
		}
	} else {
		err := errors.New("the certificate file provided is incomplete or does not match")
		return "", err
	}
	return "", nil
}
