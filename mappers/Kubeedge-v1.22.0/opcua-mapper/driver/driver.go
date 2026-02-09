package driver

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/kubeedge/mapper-framework/pkg/common"
	"k8s.io/klog/v2"
)

// OPCUAClient wraps the gopcua client
type OPCUAClient struct {
	client *opcua.Client
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
}

var (
	// Global client cache to reuse connections
	clientCache = make(map[string]*OPCUAClient)
	cacheMutex  sync.Mutex
)

func NewClient(config ProtocolConfig) (*CustomizedClient, error) {
	cacheKey := fmt.Sprintf("%s:%s", config.URL, config.UserName)

	// Here we have to check if client already exist in the cache or not.
	cacheMutex.Lock()
	if client, exists := clientCache[cacheKey]; exists {
		cacheMutex.Unlock()
		klog.V(3).Infof("Reusing existing OPC UA client for %s", config.URL)
		return &CustomizedClient{
			client:         client,
			ProtocolConfig: config,
		}, nil
	}
	cacheMutex.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	var opts []opcua.Option

	if config.SecurityPolicy == "" {
		config.SecurityPolicy = "None"
	}
	if config.SecurityMode == "" {
		config.SecurityMode = "None"
	}

	// Configure authentication and security
	if config.Certificate != "" && config.PrivateKey != "" {
		// Certificate-based authentication
		opts = []opcua.Option{
			opcua.SecurityPolicy(config.SecurityPolicy),
			opcua.SecurityModeString(config.SecurityMode),
			opcua.CertificateFile(config.Certificate),
			opcua.PrivateKeyFile(config.PrivateKey),
		}

		if config.RemoteCertificate != "" {
			opts = append(opts, opcua.RemoteCertificateFile(config.RemoteCertificate))
		}
	} else if config.UserName != "" {
		// Username/password authentication
		password := config.Password
		if password != "" && fileExists(password) {
			pass, err := readPasswordFile(password)
			if err != nil {
				cancel()
				return nil, fmt.Errorf("failed to read password file: %v", err)
			}
			password = pass
		}

		opts = []opcua.Option{
			opcua.AuthUsername(config.UserName, password),
			opcua.AuthPolicyID("username"),
		}
	}
	c, err := opcua.NewClient(config.URL, opts...)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create OPC UA client: %v", err)
	}
	if err := c.Connect(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to OPC UA server: %v", err)
	}

	klog.V(2).Infof("Successfully connected to OPC UA server: %s", config.URL)

	opcuaClient := &OPCUAClient{
		client: c,
		ctx:    ctx,
		cancel: cancel,
	}

	// Cache the client
	cacheMutex.Lock()
	clientCache[cacheKey] = opcuaClient
	cacheMutex.Unlock()

	return &CustomizedClient{
		client:         opcuaClient,
		ProtocolConfig: config,
	}, nil
}

func (c *CustomizedClient) InitDevice() error {
	klog.V(2).Infof("Initializing OPC UA device: %s", c.deviceName)

	if c.client == nil {
		return errors.New("OPC UA client not initialized")
	}

	klog.V(2).Infof("OPC UA device initialized successfully")
	return nil
}

func (c *CustomizedClient) GetDeviceData(visitor *VisitorConfig) (interface{}, error) {
	if c.client == nil {
		return nil, errors.New("OPC UA client not initialized")
	}

	c.client.mu.Lock()
	defer c.client.mu.Unlock()

	// Parse node ID
	nodeID, err := ua.ParseNodeID(visitor.NodeID)
	if err != nil {
		return nil, fmt.Errorf("invalid node ID %s: %v", visitor.NodeID, err)
	}

	req := &ua.ReadRequest{
		MaxAge: 2000,
		NodesToRead: []*ua.ReadValueID{
			{NodeID: nodeID},
		},
		TimestampsToReturn: ua.TimestampsToReturnBoth,
	}

	resp, err := c.client.client.Read(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("read failed: %v", err)
	}

	if resp.Results[0].Status != ua.StatusOK {
		return nil, fmt.Errorf("read status error: %s", resp.Results[0].Status)
	}

	// Convert variant to interface{}
	value := variantToInterface(resp.Results[0].Value)

	klog.V(4).Infof("Read from node %s: %v", visitor.NodeID, value)
	return value, nil
}

func (c *CustomizedClient) SetDeviceData(value interface{}, visitor *VisitorConfig) error {
	if c.client == nil {
		return errors.New("OPC UA client not initialized")
	}

	c.client.mu.Lock()
	defer c.client.mu.Unlock()

	// Parse node ID
	nodeID, err := ua.ParseNodeID(visitor.NodeID)
	if err != nil {
		return fmt.Errorf("invalid node ID %s: %v", visitor.NodeID, err)
	}

	// Convert value to string for ua.NewVariant
	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	case int, int8, int16, int32, int64:
		strValue = fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		strValue = fmt.Sprintf("%d", v)
	case float32, float64:
		strValue = fmt.Sprintf("%f", v)
	case bool:
		strValue = strconv.FormatBool(v)
	default:
		strValue = fmt.Sprintf("%v", v)
	}

	// Create variant from converted string value
	variant, err := ua.NewVariant(strValue)
	if err != nil {
		return fmt.Errorf("invalid value %v: %v", value, err)
	}

	// Create a write request
	req := &ua.WriteRequest{
		NodesToWrite: []*ua.WriteValue{
			{
				NodeID:      nodeID,
				AttributeID: ua.AttributeIDValue,
				Value: &ua.DataValue{
					EncodingMask: ua.DataValueValue,
					Value:        variant,
				},
			},
		},
	}

	// Executing the created write request
	resp, err := c.client.client.Write(context.Background(), req)
	if err != nil {
		return fmt.Errorf("write failed: %v", err)
	}

	if resp.Results[0] != ua.StatusOK {
		return fmt.Errorf("write status error: %s", resp.Results[0])
	}

	klog.V(3).Infof("Wrote to node %s: %v", visitor.NodeID, value)
	return nil
}

// DeviceDataWrite is called by WriteDevice in device.go and is used for devicemethod write operations.“
func (c *CustomizedClient) DeviceDataWrite(visitor *VisitorConfig, deviceMethodName, propertyName string, value interface{}) error {
	klog.V(3).Infof("DeviceDataWrite called: method=%s, property=%s, value=%v", deviceMethodName, propertyName, value)

	return c.SetDeviceData(value, visitor)
}

func (c *CustomizedClient) StopDevice() error {
	if c.client != nil {
		c.client.cancel()
		if err := c.client.client.Close(context.Background()); err != nil {
			klog.Errorf("Error closing OPC UA client: %v", err)
			return err
		}
	}
	return nil
}

func (c *CustomizedClient) GetDeviceStatus(twins []common.Twin) (status bool) {
	// Check if client is connected
	if c.client == nil {
		return false
	}

	// OPC UA client is connected, return true
	return true
}

// Helper functions

func variantToInterface(v *ua.Variant) interface{} {
	if v == nil {
		return nil
	}

	switch v.Type() {
	case ua.TypeIDBoolean:
		return v.Bool()
	case ua.TypeIDSByte, ua.TypeIDInt16, ua.TypeIDInt32, ua.TypeIDInt64:
		return v.Int()
	case ua.TypeIDByte, ua.TypeIDUint16, ua.TypeIDUint32, ua.TypeIDUint64:
		return v.Uint()
	case ua.TypeIDFloat, ua.TypeIDDouble:
		return v.Float()
	case ua.TypeIDString, ua.TypeIDXMLElement, ua.TypeIDLocalizedText, ua.TypeIDQualifiedName:
		return v.String()
	case ua.TypeIDByteString:
		return string(v.ByteString())
	default:
		return v.String()
	}
}

func readPasswordFile(filename string) (string, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	password := string(b)
	if len(password) > 0 && password[len(password)-1] == '\n' {
		password = password[:len(password)-1]
	}
	return password, nil
}

func fileExists(path string) bool {
	_, err := ioutil.ReadFile(path)
	return err == nil
}
