/*
Copyright 2021 The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package opcua

import (
	"context"
	"errors"
	"io/ioutil"
	"strconv"

	"github.com/kubeedge/mappers-go/pkg/common"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"k8s.io/klog/v2"
)

// OPCUAConfig configurations for OPCUA.
type OPCUAConfig struct {
	URL            string
	User           string
	Passwordfile   string
	SecurityPolicy string
	SecurityMode   string
	Certfile       string
	RemoteCertfile string
	Keyfile        string
}

// OPCUAClient is the client structure.
type OPCUAClient struct {
	Client *opcua.Client
}

var clients map[string]*OPCUAClient

func readPassword(filename string) (string, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", errors.New("Failed to load certificate")
	}
	// Remove the last character '\n'
	return string(b[:len(b)-1]), nil
}

// NewClient new the OPCUA client.
func NewClient(config OPCUAConfig) (client *OPCUAClient, err error) {
	ctx := context.Background()
	var opts []opcua.Option
	if config.Certfile != "" {
		opts = []opcua.Option{
			opcua.SecurityPolicy(config.SecurityPolicy),
			opcua.SecurityModeString(config.SecurityMode),
			opcua.CertificateFile(config.Certfile),
			opcua.PrivateKeyFile(config.Keyfile),
			opcua.RemoteCertificateFile(config.RemoteCertfile),
		}
	} else if config.User != "" {
		password, err := readPassword(config.Passwordfile)
		if err != nil {
			return nil, err
		}

		opts = []opcua.Option{
			opcua.AuthUsername(config.User, password),
			opcua.AuthPolicyID("username"),
		}
	} else {
		opts = []opcua.Option{}
	}

	c := opcua.NewClient(config.URL, opts...)
	if err = c.Connect(ctx); err != nil {
		return &OPCUAClient{}, err
	}
	return &OPCUAClient{Client: c}, nil
}

// GetStatus get device status.
// Please rewrite this function depending on your devices.
func (c *OPCUAClient) GetStatus() string {
	return common.DEVSTOK
}

func valueToString(v *ua.Variant) string {
	switch v.Type() {
	case ua.TypeIDBoolean:
		return strconv.FormatBool(v.Bool())
	case ua.TypeIDString, ua.TypeIDXMLElement, ua.TypeIDLocalizedText, ua.TypeIDQualifiedName:
		return v.String()
	case ua.TypeIDSByte, ua.TypeIDInt16, ua.TypeIDInt32, ua.TypeIDInt64:
		return strconv.FormatInt(v.Int(), 10)
	case ua.TypeIDByte, ua.TypeIDUint16, ua.TypeIDUint32, ua.TypeIDUint64:
		return strconv.FormatUint(v.Uint(), 10)
	case ua.TypeIDByteString:
		return string(v.ByteString())
	case ua.TypeIDFloat, ua.TypeIDDouble:
		return strconv.FormatFloat(v.Float(), 'E', -1, 32)
	default:
		return ""
	}
}

// Get get register.
func (c *OPCUAClient) Get(nodeID string) (results string, err error) {
	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		klog.Errorf("invalid node id: %v", err)
		return "", errors.New("Invalid node ID")
	}

	req := &ua.ReadRequest{
		MaxAge: 2000,
		NodesToRead: []*ua.ReadValueID{
			{
				NodeID: id,
			},
		},
		TimestampsToReturn: ua.TimestampsToReturnBoth,
	}

	resp, err := c.Client.Read(req)
	if err != nil {
		klog.Errorf("Read failed: %v", err)
		return "", err
	}
	if resp.Results[0].Status != ua.StatusOK {
		klog.Errorf("Read status failed: %v", resp.Results[0].Status)
		return "", errors.New(resp.Results[0].Status.Error())
	}
	return valueToString(resp.Results[0].Value), nil
}

// Set set register.
func (c *OPCUAClient) Set(nodeID string, value string) (results string, err error) {
	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		klog.Errorf("invalid node id: %v", err)
		return "", errors.New("Invalid node ID")
	}

	v, err := ua.NewVariant(value)
	if err != nil {
		klog.Errorf("invalid value: %v", err)
		return "", errors.New("Invalid value")
	}

	req := &ua.WriteRequest{
		NodesToWrite: []*ua.WriteValue{
			{
				NodeID:      id,
				AttributeID: ua.AttributeIDValue,
				Value: &ua.DataValue{
					EncodingMask: ua.DataValueValue,
					Value:        v,
				},
			},
		},
	}

	resp, err := c.Client.Write(req)
	if err != nil {
		klog.Errorf("Write failed: %v", err)
		return "", err
	}
	klog.V(4).Info("Set node ", nodeID, value)
	return resp.Results[0].Error(), nil
}
