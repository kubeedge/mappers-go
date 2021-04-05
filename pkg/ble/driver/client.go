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

//import (
//	"github.com/paypal/gatt"
//	"github.com/sirupsen/logrus"
//	"k8s.io/klog/v2"
//	"os"
//	"strings"
//)
//
//var (
//	done                 = make(chan struct{})
//	deviceName           string
//	DeviceConnected      = make(chan bool)
//	CharacteristicsLists = make([]*gatt.Characteristic, 0)
//	GattPeripheral       gatt.Peripheral
//)
//
////Initiate initiates the watcher module
//func Initiate(device gatt.Device, nameOfDevice string) {
//	deviceName = nameOfDevice
//	// Register optional handlers.
//	device.Handle(
//		gatt.PeripheralConnected(onPeripheralConnected),
//		gatt.PeripheralDisconnected(onPeripheralDisconnected),
//		gatt.PeripheralDiscovered(onPeripheralDiscovered),
//	)
//	if err := device.Init(onStateChanged); err != nil {
//		klog.Errorf("Init device failed with error: %v", err)
//	}
//	<-done
//	klog.Infof("Watcher Done")
//}
//
////onPeripheralConnected contains the operations to be performed as soon as the peripheral device is connected
//func onPeripheralConnected(p gatt.Peripheral, err error) {
//	GattPeripheral = p
//	ss, err := p.DiscoverServices(nil)
//	if err != nil {
//		klog.Errorf("Failed to discover services, err: %s\n", err)
//		os.Exit(1)
//	}
//	for _, s := range ss {
//		// Discovery characteristics
//		cs, err := p.DiscoverCharacteristics(nil, s)
//		if err != nil {
//			klog.Errorf("Failed to discover characteristics for service %s, err: %v\n", s.Name(), err)
//			continue
//		}
//		CharacteristicsLists = append(CharacteristicsLists, cs...)
//	}
//	for _, c := range CharacteristicsLists {
//		klog.Infof("Characteristic: %v,name: %v,properties: %v", c.UUID().String(), c.Name(), c.Properties().String())
//
//		// Read the characteristic, if possible.
//		if (c.Properties() & gatt.CharRead) != 0 {
//			b, err := p.ReadCharacteristic(c)
//			if err != nil {
//				klog.Infof("Failed to read characteristic, err: %s\n", err)
//				continue
//			}
//			logrus.Infof("value %x | %q\n", b, b)
//		}
//
//		// Discovery descriptors
//		ds, err := p.DiscoverDescriptors(nil, c)
//		if err != nil {
//			logrus.Infof("Failed to discover descriptors, err: %s\n", err)
//			continue
//		}
//
//		for _, d := range ds {
//			logrus.Infof("Descriptor: %v", d.UUID().String())
//
//			// Read descriptor (could fail, if it's not readable)
//			b, err := p.ReadDescriptor(d)
//			if err != nil {
//				logrus.Infof("Failed to read descriptor, err: %s\n", err)
//				continue
//			}
//			logrus.Infof("descriptor %q\n", b)
//		}
//		// Subscribe the characteristic, if possible.
//		if (c.Properties() & (gatt.CharNotify | gatt.CharIndicate)) != 0 {
//			f := func(c *gatt.Characteristic, b []byte, err error) {
//				logrus.Infof("notified: % X | %q\n", b, b)
//			}
//			if err := p.SetNotifyValue(c, f); err != nil {
//				logrus.Printf("Failed to subscribe characteristic, err: %s\n", err)
//				continue
//			}
//		}
//	}
//	DeviceConnected <- true
//}
//
////onPeripheralDisconnected contains the operations to be performed as soon as the peripheral device is disconnected
//func onPeripheralDisconnected(p gatt.Peripheral, err error) {
//	logrus.Infof("Disconnecting  from bluetooth device....")
//	DeviceConnected <- false
//	close(done)
//	p.Device().CancelConnection(p)
//}
//
////onPeripheralDiscovered contains the operations to be performed as soon as the peripheral device is discovered
//func onPeripheralDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
//	if strings.EqualFold(a.LocalName, strings.Replace(deviceName, "-", " ", -1)) {
//		logrus.Infof("Device: %s found !!!! Stop Scanning for devices", deviceName)
//		// Stop scanning once we've got the peripheral we're looking for.
//		p.Device().StopScanning()
//		logrus.Infof("Connecting to %s", deviceName)
//		p.Device().Connect(p)
//	}
//}
//
////onStateChanged contains the operations to be performed when the state of the peripheral device changes
//func onStateChanged(device gatt.Device, s gatt.State) {
//	switch s {
//	case gatt.StatePoweredOn:
//		logrus.Infof("Scanning for BLE device Broadcasts...")
//		device.Scan([]gatt.UUID{}, true)
//		return
//	default:
//		device.StopScanning()
//	}
//}
//
////ReadCharacteristic is used to read the value of the characteristic
//func ReadCharacteristic(p gatt.Peripheral, c *gatt.Characteristic) ([]byte, error) {
//	value, err := p.ReadCharacteristic(c)
//	if err != nil {
//		logrus.Errorf("Error in reading characteristic, err: %s\n", err)
//		return nil, err
//	}
//	return value, nil
//}
