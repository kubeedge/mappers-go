# User Guide of Customized Mapper in KubeEdge
## 1 Introduction
### 1.1 Purpose
This document is to guide the user to generate a customized mapper and run it. Now we have Bluetooth, Modbus, and Opcua mapper. For other mappers, we could follow this guide to genenrate your own mapper to control the edge devices.
### 1.2 Scope
This document is only suit for KubeEdge cloud/edge system.
 
## 2 Describing the System
### 2.1 CRD
We use CRD to define a new device. There're 2 device related CRDs. One is device model, it defines some common definitions such as device properties, the property is like temperature, humidity, and enable-counter; another CRD is device instance, it defines the detailed definition related to the specific device. More details could refer to the [device crd design](https://github.com/kubeedge/kubeedge/blob/master/docs/proposals/device-crd.md#device-model-crd).
### 2.2 Device controller, Device Twin and Mapper
For this part, you could refer to the [mapper design](https://github.com/kubeedge/kubeedge/blob/master/docs/proposals/mapper-design-v2.md).
## 3 Step by Step Instructions
### 3.1 Define the device model and device instance
As an example, the [ONVIF mapper](ONVIF mapper device definition) is one customized mapper. Here's some points to be noticed in device instance configuration.
```language
  protocol:
    customizedProtocol:
      protocolName: onvif
      configData:             ---> The connection configuration
        url: 192.168.1.64:80
```
```
  propertyVisitors:
    - propertyName: saveVideo ---> name same as it in model definition
      customizedProtocol:
        protocolName: onvif
        configData:           ---> visit method or requried values. This part will be in VisitorConfig part of the configmap structure which is aligned with BLE, Modbus mapper.
          method: SaveVideo
          frameCount: 50
          format: mp4
      customizedValues:       ---> this part will be in ProtocolCommonConfig part of the configmap structure.
        reportNumber: 1
```
The upper customized parts need to be parser by yourself in the code. The code location and method could be refered at source code [customized config](https://github.com/sailorvii/mappers-go/blob/onvif/mappers/onvif/device/device.go).

### 3.2 Generate the original code
We could use the mapper template to generate the code framework. More detailed pelease refer to [How to use mapper template](https://github.com/sailorvii/mappers-go/blob/onvif/_template/README.md).
### 3.3 Add your own code
The main parts you need to add:
1. configmap/type.go. Here defines the customized fields of configmap structure.
2. device/device.go. Here parsers and use the customized fields.
3. driver/client.go. Here add the read/write funtions.
4. deployment.yaml. Refine as your environment. Notice here.
```
volumes:
      - name: config-volume
        configMap:
          name: device-profile-config-test ---> the configmap is created after you "kubectl create" the device model and device instance. 
```
### 3.4 Make it
Run "make help" in "mappers-go" directory to get all availabe make methods. Try them and make your mapper.
### 3.5 Run
Normally, the mapper is run as a container with the deployment definition. If you want to debug as a daemon, copy the configmap to the "/opt/kubeedge/deviceProfile.json" on the host, then run the executed binary.
