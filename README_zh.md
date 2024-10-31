简体中文 | [English](./README.md)

# mappers-go
KubeEdge Device Mappers是用go编写的，适用于KubeEdge 1.4或更高版本。

## 快速上手
如果想把边缘设备连接到kubeedge，可以创建和编写mapper来实现。现提供了两种方式，"mappers"和"mapper-sdk-go"。

### mappers
Mappers实现了modbus、蓝牙、onvif和opcua。如果想将设备通过这些协议连接到kubeedge，可以在[mappers](./mappers)中找到示例。
### mapper-go-sdk
Mapper-sdk-go是一个用go编写的基础框架。基于这个框架，开发人员可以更容易的实现一个新的mapper。Mapper-sdk-go已实现与KubeEdge的连接，数据转换，管理设备的基本属性和状态等，还提供了驱动程序接口的基本功能和抽象定义。开发人员只需要实现相应设备的自定义协议驱动程序接口，即可实现mapper的功能，
可以在[mapper-sdk-go](./mapper-sdk-go/)中获取更多详细信息。
