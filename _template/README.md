# How to create your own mappers
Before you start, read the mapper design to familiar with the mapper structure and communications with other part in Kubeedge: [mapper design v2](https://github.com/kubeedge/kubeedge/blob/master/docs/proposals/mapper-design-v2.md). 

## 1. Design the device model and device instance CRDs

## 2. Generate the code template
The mapper template is to generate a framework for the customized mapper. Run the command and input your mapper's name:
```shell
github.com/kubeedge/mappers-go$ make template
Please input the mapper name (like 'Bluetooth', 'BLE'): test
```
A folder named as you input will be generated in the directory "mappers". The file tree is as below:
mapper
├── cmd ----------------------------- Main process.
│   └── main.go --------------------- Almost need not change.
├── config -------------------------- Configuration parse.
│   ├── config.go ------------------- Almost need not change.
│   ├── config_test.go
│   └── config.yaml ----------------- Configuration file including Mqtt access information.
├── configmap ----------------------- Configmap parse and generate  related structure.
│   ├── configmap_negtest.json
│   ├── configmap_test.json
│   ├── parse.go -------------------- Refine the parse process as your CRD.
│   ├── parse_test.go
│   └── type.go --------------------- Add the mapper-specific data structure here.
├── config.yaml
├── deployment.yaml
├── device -------------------------- Device management. Such as reading the device configuration and create the data structure accordingly, connecting the devices, and reading/writing device registers.
│   ├── device.go ------------------- Refine the initialization of driver client and customized visitor parsing.
│   ├── devstatus.go ---------------- Almost need not change. It's for device status process.
│   └── twindata.go ----------------- Almost need not change. It's for Twin or data message process.
├── Dockerfile
├── driver -------------------------- This is for reading/writing devices.
│   └── client.go ------------------- Fill in the functions like getting register/setting register.
├── globals
│   └── globals.go ------------------ Almost need not change.
├── hack
│   └── make-rules
│       └── mapper.sh
└── Makefile
