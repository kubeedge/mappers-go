# Bluetooth With KubeEdge Using SmartRF06 Evaluation Board - CC2650EM-7ID 

This document ported [bluetooth-CC2650-demo] to a project for [CC2650EM-7ID] mounted on [SmartRF06 Evaluation Board].

## Description

We will cover how to connect to SimpleBLEPeripheral - a basic sample of TI [BLE-STACK] - to KubeEdge. We do not use any peripherals such as LEDs and light sensors except a LCD. And we do not require firmware modifications.

More details on bluetooth mapper can be found [here]

## Prerequisites

### Hardware Prerequisites

1. Texas instruments CC2650 bluetooth device
2. Linux based edge node with bluetooth support  (An Ubuntu 18.04 laptop has been used in this demo)

### Software Prerequisites

1. Golang (1.14+)
2. KubeEdge (v1.5+)

## Steps to reproduce

1. Clone and run KubeEdge.
    Please ensure that the kubeedge setup is up and running before execution of step 4 (mentioned below).

2. Clone the kubeedge/mapper-go repository.

```console
git clone https://github.com/kubeedge/mappers-go.git $GOPATH/src/github.com/kubeedge/mappers-go
```

3. Create the CC2650EM-7ID device model and device instance.

```console
cd $GOPATH/src/github.com/kubeedge/mappers-go/build/crd-samples/devices/
kubectl apply -f CC2650-SmartRF06EB-device-model.yaml
sed -i "s#edge-node#<your-edge-node-name>#g" CC2650-SmartRF06EB-device-instance.yaml
kubectl apply -f CC2650-SmartRF06EB-device-instance.yaml
```

4. Please ensure that bluetooth service of your device is ON

5. Set 'bluetooth=true' label for the node (This label is a prerequisite for the scheduler to schedule bluetooth_mapper pod on the node [which meets the hardware / software prerequisites] )

```console
kubectl label nodes <your-edge-node-name> bluetooth=true
```

6. Include the configuration file `configuration/config-smartrf06-cc2650.yaml` for [CC2650EM-7ID] in the dockerfile.

    Original: 
    ```
    FROM ubuntu:16.04

    RUN mkdir -p kubeedge

    COPY bluetooth kubeedge/

    COPY configuration/config.yaml kubeedge/configuration/config.yaml 

    WORKDIR kubeedge

    ENTRYPOINT ["/kubeedge/bluetooth","--logtostderr=true"]

    ```

    Modified: 
    ```
    FROM ubuntu:16.04

    RUN mkdir -p kubeedge

    COPY bluetooth kubeedge/

    COPY configuration/config-smartrf06-cc2650.yaml kubeedge/configuration/config.yaml 

    WORKDIR kubeedge

    ENTRYPOINT ["/kubeedge/bluetooth","--logtostderr=true"]
    ```

7. Build the mapper by following the steps given below.

```console
cd $GOPATH/src/github.com/kubeedge/mappers-go
make bluetoothmapper_image
docker tag bluetoothmapper:v1.0 <your_dockerhub_username>/bluetoothmapper:v1.0
docker push <your_dockerhub_username>/bluetoothmapper:v1.0

# Note: Before trying to push the docker image to the remote repository please ensure that you have signed into docker from your node, if not please type the followig command to sign in
docker login
# Please enter your username and password when prompted
```

8. Deploy the mapper by following the steps given below.

```console
cd $GOPATH/src/github.com/kubeedge/mappers-go/pkg/bluetooth

# Please enter the following details in the deployment.yaml :-
#    1. Replace <edge_node_name> with the name of your edge node at spec.template.spec.voluems.configMap.name
#    2. Replace <your_dockerhub_username> with your dockerhub username at spec.template.spec.containers.image

kubectl create -f deployment.yaml
```

9. Turn ON the SmartRF06 Evaluation Board device. 

10. The bluetooth mapper is now running, You can monitor the logs of the mapper by using docker logs. You can also play around with the device twin state by altering the desired property in the device instance
and see the result reflect on the SensorTag device. The configurations of the bluetooth mapper can be altered at runtime Please click [Runtime Configurations] for more details.

## How to find the CharacteristicUUID of your device. 

> All operations below were done with root privileges.

Before creating a device instance, you need to know the information about the Bluetooth device. KubeEdge's bluetooth_mapper uses the [paypal/gatt] go module for Bluetooth control.


1. Make sure your device is in 'advertising' state. If you are on an existing connection, disconnect it.

2. Check the prerequisites for [paypal/gatt].

```console 
hciconfig hci0 down
systemctl stop bluetooth
```

See the [paypal/gatt] readme.md for details.

3. Clone the paypal/gatt repository.

```console 
go get github.com/paypal/gatt
cd $GOPATH/src/github.com/paypal/gatt/
```

4. Find the device's peripheral ID


```console
# go run examples/discoverer.go 
2021/04/03 15:11:10 dev: hci0 up
2021/04/03 15:11:10 dev: hci0 down
2021/04/03 15:11:10 dev: hci0 opened
State: PoweredOn
scanning...
2021/04/03 15:11:11 DATA: [ 50 00 20 03 ]

Peripheral ID:AA:AA:AA:AA:AA:AA, NAME:(SimpleBLEPeripheral)
  Local Name        = SimpleBLEPeripheral
  TX Power Level    = 0
  Manufacturer Data = []
  Service Data      = []

```

For example, a device 'SimpleBLEPeripheral' has a Peripheral ID `AA:AA:AA:AA:AA:AA'.


5. Use the peripheral ID found in 4. to find the device's profile information.

```console
# go run examples/explorer.go AA:AA:AA:AA:AA:AA
2021/04/03 15:21:23 dev: hci0 up
2021/04/03 15:21:23 dev: hci0 down
2021/04/03 15:21:23 dev: hci0 opened
State: PoweredOn
Scanning...
2021/04/03 15:21:24 DATA: [ 50 00 20 03 ]

Peripheral ID:AA:AA:AA:AA:AA:AA, NAME:(SimpleBLEPeripheral)
  Local Name        = SimpleBLEPeripheral
  TX Power Level    = 0
  Manufacturer Data = []
  Service Data      = []

Connected
Service: 1800 (Generic Access)
  Characteristic  2a00 (Device Name)
    properties    read 
    value         53696d706c6520424c45205065726970686572616c | "Simple BLE Peripheral"
  Characteristic  2a01 (Appearance)
    properties    read 
    value         0000 | "\x00\x00"
  Characteristic  2a04 (Peripheral Preferred Connection Parameters)
    properties    read 
    value         5000a0000000e803 | "P\x00\xa0\x00\x00\x00\xe8\x03"
…
…

Service: fff0
  Characteristic  fff1
    properties    read write 
    value         00 | "\x00"
  Descriptor      2901 (Characteristic User Description)
    value         43686172616374657269737469632031 | "Characteristic 1"
  Characteristic  fff2
    properties    read 
    value         2002 | " \x02"
  Descriptor      2901 (Characteristic User Description)
    value         43686172616374657269737469632032 | "Characteristic 2"
  Characteristic  fff3
    properties    write 
  Descriptor      2901 (Characteristic User Description)
    value         43686172616374657269737469632033 | "Characteristic 3"
  Characteristic  fff4
    properties    notify 
  Descriptor      2902 (Client Characteristic Configuration)
    value         0000 | "\x00\x00"
  Descriptor      2901 (Characteristic User Description)
    value         43686172616374657269737469632034 | "Characteristic 4"
  Characteristic  fff5
    properties    read 
    value         0a2b0005 | "\n+\x00\x05"
  Descriptor      2901 (Characteristic User Description)
    value         43686172616374657269737469632035 | "Characteristic 5"

Waiting for 5 seconds to get some notifiations, if any.
Disconnected
Done
```

We can see that the CharacteristicUUID of service characteristic 1 is `fff1`, and the CharacteristicUUID of characteristic 2 is `fff2`. These are written to `spec.propertyVisitors[].bluetooth.characteristicUUID` in the device instance CRD.

```yaml 
# cat CC2650-SmartRF06EB-device-instance.yaml
apiVersion: devices.kubeedge.io/v1alpha2
kind: Device
metadata:
  name: smartrf06eb-instance-01
  labels:
    description: TISmartRF06EB
    manufacturer: TexasInstruments
    model: simplebleperipheral
spec:
  deviceModelRef:
    name: simplebleperipheral
  protocol:
    bluetooth: {} 
  propertyVisitors:
    - propertyName: simpleprofilechar1
      bluetooth:
        characteristicUUID: fff1
        dataWrite:
          "1": [1]
          "2": [2]
          "3": [3]
          "4": [4]
          "5": [5]
          "6": [6]
          "7": [7]
          "8": [8]
          "9": [9]
          "A": [10]
          "B": [11]
          "C": [12]
          "D": [13]
          "E": [14]
          "F": [15]
    - propertyName: simpleprofilechar2
      bluetooth:
        characteristicUUID: fff2
  nodeSelector:
    nodeSelectorTerms:
    - matchExpressions:
      - key: ''
        operator: In
        values:
          - edge-node #pls give your edge node name
status:
  twins:
    - propertyName: simpleprofilechar1
      desired:
        value: "0"

```

[SmartRF06 Evaluation Board]: https://www.ti.com/tool/SMARTRF06EBK
[CC2650EM-7ID]: https://www.ti.com/tool/CC2650EM-7ID-RD
[bluetooth-CC2650-demo]: https://github.com/kubeedge/examples/tree/master/bluetooth-CC2650-demo
[here]: https://github.com/kubeedge/kubeedge/blob/master/docs/components/mappers/bluetooth_mapper.md
[BLE-STACK]: https://www.ti.com/tool/BLE-STACK
[Runtime Configurations]: https://github.com/kubeedge/kubeedge/blob/master/docs/components/mappers/bluetooth_mapper.md#runtime-configuration-modifications
[paypal/gatt]: https://github.com/paypal/gatt
