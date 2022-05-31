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

package ble

import (
	"sync"
	"time"

	"github.com/kubeedge/mappers-go/pkg/common"

	"github.com/currantlabs/ble"
	"github.com/currantlabs/ble/linux"
	"golang.org/x/net/context"
	"k8s.io/klog/v2"
)

type BleConfig struct {
	Addr string
}

// BleClient is the client structure.
type BleClient struct {
	Client ble.Client
	mu     sync.Mutex
}

func NewClient(config BleConfig) (bc BleClient, err error) {
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), 1*time.Minute))
	host, err := linux.NewDevice()
	if err != nil {
		klog.Errorf("New device error: %v", err)
		return
	}
	bc.Client, err = host.Dial(ctx, ble.NewAddr(config.Addr))
	return
}

// Disconnect from a Device
func (bc *BleClient) Disconnect() error {
	return bc.Client.CancelConnection()
}

func (bc *BleClient) Set(c ble.UUID, b []byte) error {
	if p, err := bc.Client.DiscoverProfile(true); err == nil {
		if u := p.Find(ble.NewCharacteristic(c)); u != nil {
			c := u.(*ble.Characteristic)
			if err := bc.Client.WriteCharacteristic(c, b, false); err != nil {
				klog.Errorf("Write characteristic error %v", err)
				return err
			}
		}
	}
	return nil
}

func (bc *BleClient) Read(c *ble.Characteristic) ([]byte, error) {
	return bc.Client.ReadCharacteristic(c)
}

// GetStatus get device status.
// Now we could only get the connection status.
func (bc *BleClient) GetStatus() string {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	rssi := bc.Client.ReadRSSI()
	if rssi < 0 {
		return common.DEVSTOK
	}
	return common.DEVSTDISCONN
}
