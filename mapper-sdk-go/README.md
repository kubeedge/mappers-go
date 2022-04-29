![Gitee Latest Dev Tag](https://img.shields.io/badge/latest--dev-v0.0.1-orange) ![Gitee go.mod Go version](https://img.shields.io/badge/Go-v1.17-brightgreen) [![Gitee License](https://camo.githubusercontent.com/3e671e69d5fad7978893d028dcdeb3af16edb20b61f23cd276f738a76f33f3cf/68747470733a2f2f696d672e736869656c64732e696f2f6769746875622f6c6963656e73652f6b756265656467652f6b756265656467652e7376673f7374796c653d666c61742d737175617265)](https://gitee.com/ascend/mappers-go-sample/blob/mapper-go-sdk/LICENSE)
# MapperSDK
Before you start, read the mapper-sdk-go design to familiar with the mapper-sdk-go structure:[Mapper SDK Design](../docs/MapperDesign.md)
## OverView
This repository is a set of Go packages that can be used to build Go-based mapper for use within the KubeEdge framework.

## QuickStart
1. Developers need to provide [CRD](../build/crd-samples)  to generate configmap.If you need a secure connection, configure the cert's path in the [config.yaml](../_template/mapper-sdk/res/config.yaml).
2. Use the following instructions in [Makefile](../Makefile) to generate mapper-sdk-go model
```shell
   make sdkmodel
```
You can find mapper-sdk-go model in [mappers](../mappers)
3. Developers can make their own mapper by implementing the [ProtocolDriver](pkg/models/protocoldriver.go) interface for their desired IoT protocol.


## Command Line Options

      --config-file string          Config file name (default "..\\res\\config.yaml")
      --mqtt-address string         MQTT broker address
      --mqtt-certification string   certification file path
      --mqtt-password string        password
      --mqtt-privatekey string      private key file path
      --mqtt-username string        username
      --v string                    log level (default "1")


## Supported MQTT
### MQTT topics
	TopicTwinUpdateDelta = "$hw/events/device/%s/twin/update/delta"  
	TopicStateUpdate     = "$hw/events/device/%s/state/update"
	TopicTwinUpdate      = "$hw/events/device/%s/twin/update"
	TopicDataUpdate      = "$ke/events/device/%s/data/update"
	TopicDeviceUpdate    = "$hw/events/node/%s/membership/updated"
### 
1. `$hw/events/device/+/twin/update/delta`:This topic is used to synchronize cloud data. + symbol can be replaced with ID of the device whose state is to be updated.
2. `$hw/events/device/+/state/update`: This topic is used to update the state of the device. + symbol can be replaced with ID of the device whose state is to be updated.
3. `$hw/events/device/+/twin/+`: The two + symbols can be replaced by the deviceID on whose twin the operation is to be performed and any one of(update,cloud_updated,get) respectively.
4. `$ke/events/device/+/data/update`: This topic is add in KubeEdge v1.4, and used for delivering time-serial data. This topic is not processed by edgecore, instead, they should be processed by third-party component on edge node such as EMQ Kuiper.
5. `$hw/events/node/%s/membership/updated`: This topic is used to remove/add device. + symbol can be replaced with ID of the device whose state is to be updated.
### In addition
If you want to accept large packets over HTTPS instead of mqtt, you can set ```CollectCycle``` to ```-1``` in configmap.  
Then the twin that ```CollectCycle``` be sett to ```-1``` will not be actively reported to mqtt broker
## Enable MQTT Security Features
### Generate the self-signed CA certificate
First, we need a self signed CA certificate. If you want to generate this certificate, you need to sign it with a private key. You can generate this private key by executing the following command:
```shell
openssl genrsa -out ca.key 2048
```
This command will generate a key with a length of 2048 and store it in ca.key. If you have this key, you can use it to generate a root certificate:
```shell
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.crt
```
The root certificate is the starting point of the whole trust chain. If the issuer of each level of certificate and the issuer of the root certificate are trusted, the certificate is trusted. We can use it to issue certificates for the mqtt used by the edge part of kubeedge.
### Generate server certificate
Next, you need to generate the server private key to ensure the control of its certificate. The process of generating the private key is similar to the above:
```shell
openssl genrsa -out server.key 2048
```
Create file ``openssl.cnf``
```
[req]
default_bits  = 2048
distinguished_name = req_distinguished_name
req_extensions = req_ext
x509_extensions = v3_req
prompt = no
[req_distinguished_name]
countryName = CN
stateOrProvinceName = Hubei
localityName = Wuhan
organizationName = JS
commonName = kubeedge
[req_ext]
subjectAltName = @alt_names
[v3_req]
subjectAltName = @alt_names
[alt_names]
IP.1 = BROKER_ADDRESS
DNS.1 = BROKER_ADDRESS
```
`req_distinguished_name` ：according to the situation to modify

`alt_names`： modify BROKER_ADDRESS to the real IP or DNS address of mqtt broker such as IP.1 = 127.0.0.1 or DNS.1 = kubeedge.mapper.com


Then, use this key and configuration to issue a request to generate a certificate
```shell
openssl req -new -key ./server.key -config openssl.cnf -out server.csr
```
Then use the root certificate to issue the certificate of server:
```shell
openssl x509 -req -in ./server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 3650 -sha256 -extensions v3_req -extfile openssl.cnf
```
### Generate client certificate
If you need to use two-way connection authentication, you also need to create a certificate for the client. First, we need to create the client key:
```shell
openssl genrsa -out client.key 2048
```
Create a request to generate a client certificate using the client key:
```shell
openssl req -new -key client.key -out client.csr -subj "/C=CN/ST=Hubei/L=Wuhan/O=JS/CN=client"
```
Finally, you should use the generated CA certificate to sign the client and generate the client certificate:
```shell
openssl x509 -req -days 3650 -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt
```
### Next steps to enable MQTT TLS/SSL
After completing the appeal operation, we can enable TLS / SSL two-way authentication in edgecore. Edgecore uses mosquitto as mqtt broker by default.

First, we need to move the generated `ca.crt` certificate file to the `/etc/kubeedge/ca` directory, and put the `client.crt` and `client.key` to `/etc/kubeedge/certs` directory.

Then, we need to edit `edgecore.yaml`, which is located in the `/etc/kubedge/config` directory.
We need to change the following field properties to the certificate path we just generated, such as the following:
```yaml
  eventBus:
    enable: true
    eventBusTLS:
      enable: true
      tlsMqttCAFile: /etc/kubeedge/ca/ca.crt
      tlsMqttCertFile: /etc/kubeedge/certs/client.crt
      tlsMqttPrivateKeyFile: /etc/kubeedge/certs/client.key
    mqttMode: 2
    mqttPassword: ""
    mqttPubClientID: ""
    mqttQOS: 0
    mqttRetain: false
    mqttServerExternal: tcps://192.168.137.51:8883
    mqttServerInternal: tcp://127.0.0.1:1884
    mqttSessionQueueSize: 100
    mqttSubClientID: ""
    mqttUsername: ""
```
Finally, you need to move the `ca.crt` file to the `/etc/mosquitto/ca_certificates` directory and move the `server.crt` file and `server.key` file to the `/etc/mosquitto/certs` directory.After that, you should modify `/etc/mosquitto/mosquitto.conf`, add the following content at the bottom of the file.The file path needs to be modified according to the specific situation.
```
listener 8883
cafile /etc/mosquitto/ca_certificates/ca.crt
certfile /etc/mosquitto/certs/server.crt
keyfile /etc/mosquitto/certs/server.key
allow_anonymous true
require_certificate true
use_identity_as_username true
```
After finished configuring,You need to execute the following command:
1. **Restart mosquitto**
```shell
# find process id
ps -aux | grep mosquitto
# kill the process
kill -9 xxx
# start mosquitto
mosquitto -c /etc/mosquitto/mosquitto.conf -d
```
2. **Restart edgecore**
```shell
systemctl restart edgecore.service
```
So far, MQTT TLS/SSL two-way authentication in edgecore is completed.
## Supported Restful API
The URLs listed below are given in the form of local IP. You can use these services from any network accessible to mapper   

Port ```1215``` is enabled by default.      

```deviceInstances-ID```  
according to your own CRD definition  

```propertyName```  
according to your own CRD definition  

If you have any questions,you can see examples in the [example](../mappers/virtualdevice-sdk/README.md)  

The functions and urls are as follows. 
1. Detect whether the RESTful service starts normally  
   Method=<font color=green>**GET**</font>   
   https://127.0.0.1:1215/api/v1/ping

2. Get device's property  
Method=<font color=green>**GET**</font>  
https://127.0.0.1:1215/api/v1/device/id/deviceInstances-ID/propertyName

3. Set device's config(If you want to use this method,cloudCore's `Twin.Desired` should be null)  
Method=<font color=#60D6F4>**PUT**</font> 
https://127.0.0.1:1215/api/v1/device/id/deviceInstances-ID?propertyName=Value
4. Add a deviceInstance  
Method=<font color=orange>**POST**</font>  
https://127.0.0.1:1215/api/v1/callback/device  
You must provide a JSON body that conforms to the CRD definition
5. Delete a deviceInstance  
   Method=<font color=#FF5555>**DEL**</font>
   https://127.0.0.1:1215/api/v1/callback/device/id/deviceInstances-ID
## Enable Restful Security Features
The steps for generating certificates are similar to those for MQTT certificates. You can refer to the MQTT certificate generation steps.

Make the following convention, use a three-digit binary number, from left to right, represent the file path of the CA certificate, the server certificate, and the server key, whether the file path is provided, the number zero means the file is not provided, and the number one means the file is provided.

|     | Status                 |
|-----|------------------------|
| 000 | No certification       |
| 001 | Illegal                |
| 010 | Illegal                |
| 011 | One-way authentication |
| 100 | Illegal                |
| 101 | Illegal                |
| 110 | Illegal                |
| 111 | Two-way authentication |
The `config.yaml` provided by the user must comply with the above agreement.
## More details

You can get more details in [UserGuideofMapperSDK](../docs/UserGuideofMapperSDK.md)