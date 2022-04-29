# MapperSDK Example
Use a sample device to guide you to use MapperSDK

## Function introduction
Generate random numbers and report them to the cloud, and control 
the range of generating numbers according to the value returned by the cloud

## How to run this code
```shell
cd cmd
```
```shell
go run main.go --v 4 // Debug level log
```

## MQTT Information
### MQTT Security Configuration
Developers need to configure their own authentication keys. Mapper-sdk enables server-side security authentication by default, but does not enable client-side authentication. If necessary, we will add it in the next version.
Developers need to provide the root certificate path, client key path and client certificate path in the `config.yaml` file.
### Send
The program can collect information from device.Then send to mqtt broker according to the ```CollectCycle``` in configmap  
If you set the log level to 4, you can see the information in the terminal.
### Set
The program can subscribe to messages sent from the cloud and perform corresponding tasks. You can use these topics  
1. ```$hw/events/device/``` + random-instance-01+```/twin/update/delta```  to set VirtualDevice's limit
2. ```"$hw/events/device/```+random-instance-01+```/state/update"``` to get VirtualDevice's state
3. ```"$hw/events/node/```+ random-instance-01+```/membership/updated"``` to check the JSON file and add or delete the device according to the payload content
## RestFul API Usage Example
### HTTPS Security Configuration
Developers need to configure the authentication key by themselves. By default, mapper-sdk enables https two-way security authentication for restful interfaces. The authentication of the client to the server needs to be implemented by the developer.
Developers need to provide the root certificate path, server key path and server certificate path in the config.yaml file.  
This example provides the security key of the client side.You can find them in [res/https-client-key](./res/https-client-key).Configure the key to your HTTPS client to access the following restful API
### <font color=green>**GET**</font>   GetFloatData
```https://127.0.0.1:1215/api/v1/device/id/random-instance-01/random-float```

### <font color=orange>**POST**</font> AddDevice

```https://127.0.0.1:1215/api/v1/callback/device```
#### Body
configmap for new devices,like this
```json
{
    "deviceInstances": [
        {
            "id": "random-instance-05",
            "name": "random-instance-05",
            "protocol": "customized-protocol-random-instance-02",
            "model": "random-01",
            "twins": [
                {
                    "propertyName": "random-int",
                    "desired": {
                        "value": "100",
                        "metadata": {
                            "timestamp": "1550049403598",
                            "type": "integer"
                        }
                    },
                    "reported": {
                        "value": "100",
                        "metadata": {
                            "timestamp": "1550049403598",
                            "type": "integer"
                        }
                    }
                },
                {
                    "propertyName": "random-float",
                    "desired": {
                        "value": "30",
                        "metadata": {
                            "timestamp": "1550049403598",
                            "type": "float"
                        }
                    },
                    "reported": {
                        "value": "30",
                        "metadata": {
                            "timestamp": "1550049403598",
                            "type": "float"
                        }
                    }
                }
            ],
            "propertyVisitors": [
                {
                    "name": "random-int",
                    "propertyName": "random-int",
                    "modelName": "random-01",
                    "protocol": "customized-protocol",
                    "visitorConfig": {
                        "protocolName": "virtualProtocol",
                        "configData": {
                            "dataType": "int"
                        }
                    }
                },
                {
                    "name": "random-float",
                    "propertyName": "random-float",
                    "modelName": "random-01",
                    "protocol": "customized-protocol",
                    "visitorConfig": {
                        "protocolName": "virtualProtocol",
                        "configData": {
                            "dataType": "float"
                        }
                    }
                }
            ]
        }
    ],
    "deviceModels": [
        {
            "name": "random-01",
            "properties": [
                {
                    "name": "random-int",
                    "dataType": "int",
                    "description": "random-int",
                    "accessMode": "ReadWrite",
                    "defaultValue": 100,
                    "minimum": 0,
                    "maximum": 0
                },
                {
                    "name": "random-float",
                    "dataType": "float",
                    "description": "random-float",
                    "accessMode": "ReadOnly",
                    "defaultValue": 30,
                    "minimum": 0,
                    "maximum": 0
                }
            ]
        }
    ],
    "protocols": [
        {
            "name": "customized-protocol-random-instance-05",
            "protocol": "customized-protocol",
            "protocolConfig": {
                "protocolName": "virtualProtocol",
                "configData": {
                    "deviceID": 1
                }
            },
            "protocolCommonConfig": {
                "customizedValues": {
                    "protocolID": 1
                }
            }
        }
    ]
}
```

### <font color=#60D6F4>**PUT**</font> WriteData
```https://127.0.0.1:1215/api/v1/device/id/random-instance-01?random-int=1```
#### QueryParams:random-int

### <font color=#FF5555>**DEL**</font>  RemoveDevice
```https://127.0.0.1:1215/api/v1/callback/device/id/random-instance-01```

