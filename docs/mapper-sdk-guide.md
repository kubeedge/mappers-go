# User Guide of Customized Mapper SDK in KubeEdge
## 1 Introduction
### 1.1 Purpose
The purpose of this document is to guide users to generate a custom 
mapper by mapper-sdk-go and run it. 
Now we have **_ModBus_** , **_Gige_** and **_device that communicates 
over a network_**  mappers. For other protocol, you can generate your own mapper according to this guide to control the
edge devices.
### 1.2 Scope
This document is suit for KubeEdge cloud/edge system.
If you have a MQTT Broker and provide a configmap.json, you can use
this mapper to connect your device easily.
## 2 Describing the System
### 2.1 CRD
We use CRD to define a new device. There're 2 device related CRDs. One is device model, it defines some common definitions 
such as device properties, the property is like temperature, humidity, and enable-counter; another CRD is device instance, 
it defines the detailed definition related to the specific device. More details could refer to the [device crd design](https://github.com/kubeedge/kubeedge/blob/master/docs/proposals/device-crd.md#device-model-crd).  
In order for mapper-sdk to recognize your protocol normally,  `protocol customizedProtocol protocolName` in **instance.yaml** must be filled in
```yaml
spec:
  deviceModelRef:
    name: modelName
  protocol:
    customizedProtocol:
      protocolName: protocolName # must be filled in
```
### 2.2 Mapper-SDK-GO Design
For this part, you could refer to the [mapper-sdk-go design](MapperDesign.md).

## 3 Step by Step Instructions
Use example to help build your first mapper.
### 3.1 Define the device model and device instance
As an example, the VirtualDevice mapper is a customized mapper. Its function is to generate random int and float numbers, 
and can send commands through the cloud  or RESTful API to control the maximum value of int nums.
You can find the [model](../build/crd-samples/devices/random-device-model.yaml) and [instance](../build/crd-samples/devices/random-device-instance.yaml) definition of virtual device in [build](../build/crd-samples/devices).

### 3.2 Define structure
Mapper-sdk-go will parse the configmap you defined. The most important three parameters are
`protocolCommonConfig` `visitorConfig` `ProtocolConfig` . These parameters will provide to you in ```[]byte```.
So define your structure according to your own configmap. You can get an example of virtual Instance:[virtual Instance Driver](../mappers/virtualdevice-sdk/driver/sampledriver.go)

### 3.3 Implements interface
Mapper-sdk-go will provide the following interfaces,and you can find them in [protocoldriver.go](../mapper-sdk-go/pkg/models/protocoldriver.go)
```go
type ProtocolDriver interface {
// InitDevice  do the job of initialization
// input:
//         protocolCommon:   protocolCommonConfig in configmap file
// return:
//         err :   error information
InitDevice(protocolCommon []byte) (err error)

// ReadDeviceData is an interface that reads data from a specific device, 
// data is a type of string
// input:
//         protocolCommon:   protocolCommonConfig in configmap file
//         visitor:          visitorConfig in configmap file
//         protocol:         protocolConfig in configmap file
// return:
//         err :   error information
ReadDeviceData (protocolCommon, visitor, protocol []byte) (data interface{}, err error)

// WriteDeviceData is an interface that write data to a specific device,
// data's DataType is Consistent with configmap
// input: 
//         data:   Data written to the device,type depends on the configuration file
//         protocolCommon:   protocolCommonConfig in configmap file
//         visitor:          visitorConfig in configmap file
//         protocol:         protocolConfig in configmap file
// return:
//         err :   error information
WriteDeviceData (data interface{}, protocolCommon, visitor, protocol []byte) (err error)

// GetDeviceStatus is an interface to get the device status true is OK , 
// false is DISCONNECTED
// input:
//         protocolCommon:   protocolCommonConfig in configmap file
//         visitor:          visitorConfig in configmap file
//         protocol:         protocolConfig in configmap file
// return:
//         status :   get devices's status  true means normal
GetDeviceStatus(protocolCommon, visitor, protocol []byte) (status bool)

// StopDevice is an interface to stop all devices
// input:
//         nil
// return:
//         err :   error information
StopDevice() (err error)
}
```
### 3.4 Make it
Run "make help" in mappers directory to get all availabe make methods. Try them and make your mapper.
### 3.5 More information about examples
Refer to [README.md](../mappers/virtualdevice-sdk/README.md), Start your mapper service