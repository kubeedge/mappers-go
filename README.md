# mappers-go
KubeEdge Device Mappers written in go, it works for KubeEdge 1.4 version and later.

## Construction method
If you want to connect your edge device to kubeedge, you can create and write mapper to complete it. Now, two methods 
are provided, the basic ```mappers``` and ```mapper-sdk-go```.
### mappers
Mappers implement modbus, bluetooth, onvif and opcua. If you want to connect devices to kubeedge of these protocols, you can find
them directly in [mappers](./mappers).
### mapper-go-sdk
Mapper-sdk-go is a basic framework written in go. Based on this framework, developers can more easily implement a new mapper. Mapper-sdk has realized the connection to kubeedge, provides data conversion, and manages the basic properties and status of devices, etc. Basic capabilities and abstract definition of the driver interface. Developers only need to implement the customized protocol driver interface of the corresponding device to realize the function of mapper.
You can get more information and details in [mapper-sdk-go](./mapper-sdk-go/)
