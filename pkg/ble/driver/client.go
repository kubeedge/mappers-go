package driver

import (
	"github.com/currantlabs/ble"
	"github.com/currantlabs/ble/linux"
	"golang.org/x/net/context"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

type BluetoothConfig struct {
	Addr string
}

// OPCUAClient is the client structure.
type BluetoothClient struct {
	Client ble.Client
	mu     sync.Mutex
}

func NewClient(config BluetoothConfig) (bc BluetoothClient, err error) {
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), 1*time.Minute))
	host, err := linux.NewDevice()
	if err != nil {
		klog.Error(err)
		return
	}
	bc.Client, err = host.Dial(ctx, ble.NewAddr(config.Addr))
	return
}

// Disconnect from a Device
func (bc *BluetoothClient) Disconnect() error {
	return bc.Client.CancelConnection()
}

func (bc *BluetoothClient) Set(c ble.UUID, b []byte) error {
	if p, err := bc.Client.DiscoverProfile(true); err == nil {
		if u := p.Find(ble.NewCharacteristic(c)); u != nil {
			c := u.(*ble.Characteristic)
			if err := bc.Client.WriteCharacteristic(c, b, false); err != nil {
				klog.Error(err)
				return err
			}
		}
	}
	return nil
}

// GetStatus get device status.
// Now we could only get the connection status.
func (bc *BluetoothClient) GetStatus() string {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	rssi := bc.Client.ReadRSSI()
	if rssi < 0 {
		return DEVSTOK
	}
	return DEVSTDISCONN
}
