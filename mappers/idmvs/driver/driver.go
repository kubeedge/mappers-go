package driver

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

// IDMVSProtocolConfig is the protocol config structure.
type IDMVSProtocolConfig struct {
	ProtocolName       string `json:"protocolName"`
	ProtocolConfigData `json:"configData"`
}
// ProtocolConfigData is the protocol config data structure.
type ProtocolConfigData struct {
}

// IDMVSProtocolCommonConfig is the protocol common config structure.
type IDMVSProtocolCommonConfig struct {
	CommonCustomizedValues `json:"customizedValues"`
}
// CommonCustomizedValues is the customized values structure.
type CommonCustomizedValues struct {
	Port int `json:"TCPport"`
}
// IDMVSVisitorConfig  is the visitor config structure.
type IDMVSVisitorConfig struct {
	ProtocolName      string `json:"protocolName"`
	VisitorConfigData `json:"configData"`
}

// VisitorConfigData  is the visitor config data  structure.
type VisitorConfigData struct {
}

// IDMVS Realize the structure
type IDMVS struct {
	mutex                sync.Mutex
	protocolConfig       IDMVSProtocolConfig
	protocolCommonConfig IDMVSProtocolCommonConfig
	visitorConfig        IDMVSVisitorConfig
	listeners            map[int]*IDMVSInstance
}

// IDMVSInstance is the instance structure
type IDMVSInstance struct {
	server      net.Listener
	codeValue   string
	status      bool
	reportTimes int
}

// InitDevice Sth that need to do in the first
// If you need mount a persistent connection, you should provide parameters in configmap's protocolCommon.
// and handle these parameters in the following function
func (d *IDMVS) InitDevice(protocolCommon []byte) (err error) {
	if protocolCommon != nil {
		if err = json.Unmarshal(protocolCommon, &d.protocolCommonConfig); err != nil {
			fmt.Printf("Unmarshal ProtocolCommonConfig error: %v\n", err)
			return err
		}
	}
	d.connect()
	return nil
}

// SetConfig Parse the configmap's raw json message
// In the case of high concurrency, d.mutex helps you get the correct value
func (d *IDMVS) SetConfig(protocolCommon, visitor, protocol []byte) (port int, err error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if protocolCommon != nil {
		if err = json.Unmarshal(protocolCommon, &d.protocolCommonConfig); err != nil {
			fmt.Printf("Unmarshal protocolCommonConfig error: %v\n", err)
			return 0, err
		}
	}
	if visitor != nil {
		if err = json.Unmarshal(visitor, &d.visitorConfig); err != nil {
			fmt.Printf("Unmarshal visitorConfig error: %v\n", err)
			return 0, err
		}
	}
	if protocol != nil {
		if err = json.Unmarshal(protocol, &d.protocolConfig); err != nil {
			fmt.Printf("Unmarshal protocolConfig error: %v\n", err)
			return 0, err
		}
	}
	return d.protocolCommonConfig.Port, nil
}

// ReadDeviceData  is an interface that reads data from a specific device, data is a type of string
func (d *IDMVS) ReadDeviceData(protocolCommon, visitor, protocol []byte) (data interface{}, err error) {
	// Parse raw json message to get a IDMVS instance
	port, err := d.SetConfig(protocolCommon, visitor, protocol)
	if err != nil {
		return nil, err
	}
	if d.listeners[port].reportTimes == 0 {
		return "NoRead", nil
	}

	d.listeners[port].reportTimes--
	return d.listeners[port].codeValue, nil
}

// WriteDeviceData is an interface that write data to a specific device, data's DataType is Consistent with configmap
func (d *IDMVS) WriteDeviceData(data interface{}, protocolCommon, visitor, protocol []byte) (err error) {
	return nil
}

// StopDevice is an interface to disconnect a specific device
// This function is called when mapper stops serving
func (d *IDMVS) StopDevice() (err error) {
	return nil
}

// GetDeviceStatus is an interface to get the device status true is OK , false is DISCONNECTED
func (d *IDMVS) GetDeviceStatus(protocolCommon, visitor, protocol []byte) (status bool) {
	port, err := d.SetConfig(protocolCommon, visitor, protocol)
	if err != nil {
		return false
	}
	return d.listeners[port].status
}

func (d *IDMVS) connect() {
	if d.listeners == nil {
		d.listeners = make(map[int]*IDMVSInstance)
	}
	clientPort := strconv.Itoa(d.protocolCommonConfig.Port)
	address := fmt.Sprintf("0.0.0.0:%s", clientPort)
	d.listeners[d.protocolCommonConfig.Port] = new(IDMVSInstance)
	var err error
	d.listeners[d.protocolCommonConfig.Port].server, err = net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("Listen %s server failed,err:%s\n", address, err)
		return
	}
	go func() {
		conn, err := d.listeners[d.protocolCommonConfig.Port].server.Accept()
		if err != nil {
			fmt.Println("Listen.Accept failed,err:", err)
		}
		defer conn.Close()
		for {
			buf := make([]byte, 256)
			n, err := conn.Read(buf[:])
			if err != nil {
				fmt.Println("Read from tcp server failed,err:", err)
				d.listeners[d.protocolCommonConfig.Port].status = false
			} else {
				data := buf[:n]
				barCode := *(*string)(unsafe.Pointer(&data))
				barCode = strings.TrimRight(strings.TrimLeft(barCode, "<p>"), "</p>")
				d.listeners[d.protocolCommonConfig.Port].codeValue = barCode
				d.listeners[d.protocolCommonConfig.Port].status = true
				d.listeners[d.protocolCommonConfig.Port].reportTimes = 2 //one for devicetwin,and another for third-part application
			}
		}
	}()
}
