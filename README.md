# mappers-go
KubeEdge Device Mappers written in go, it works for KubeEdge 1.4 version and later.

## Construction method
If you want to connect your edge device to kubeedge, you can create and write mapper to complete it. Now, two methods 
are provided, the basic ```mappers``` and ```mapper-sdk-go```.
### mappers
Mappers implement modbus, bluetooth, onvif and opcua. If you want to connect devices to kubeedge of these protocols, you can find
them directly in [mappers](./mappers).
### mapper-go-sdk
Mapper-sdk-go is written based on the basic mapper, which provides a convenient RESTful interface. Users can quickly 
get device data and device status through these interfaces.
Mapper-sdk-go supports connect customized devices to kubeedge fastly. If you want to access a device with a new protocol, 
mapper-sdk-go is a good choice.
You can get more information and details in [mapper-sdk-go](./mapper-sdk-go/)
