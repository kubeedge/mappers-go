# MapperDMI Example
Use a sample device to guide you to use mapper by dmi

## Function introduction
Generate random numbers and report them to the cloud, and control
the range of generating numbers according to the value returned by the cloud

## How to run this code

### Start Mapper
Ensure that the running node of the code is in nodeSelector's list
```shell
cd cmd
```
```shell
go run main.go --v 4 --config-file=../config.yaml
```

### Add Device
1. copy ```resource/random-device-model.yaml``` and ```resource/random-device-instance.yaml``` to  cloud node
2. ```kubectl apply -f random-device-model.yaml```
3. Modify the nodeSelector in ```random-device-instance.yaml```, make the value of the field the name of your edge node. (You can get the edge node name by ```kubectl get nodes```)
4. ```kubectl apply -f random-device-instance.yaml```
5. Ensure that the device has been added correctly: ```kubectl get device```

### Warning
This demo is for testing purposes only. If you want to deploy in a production environment, please use container deployment. And you can get more details by ```make```.