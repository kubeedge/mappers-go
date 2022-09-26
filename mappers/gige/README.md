# GigE Mapper

GigE Mapper supports GigE Vision protocol cameras access. It uses the third-party open source camera SDK library [rc_genicam_api](https://github.com/roboception/rc_genicam_api) to access cameras from different manufacturers through the GeniCam protocol. Due to resource constraints, only two cameras, Basler acA640 and HIKROBOT MV-CA050-12GC, are currently tested.

## Supported functions

- Support the connection of GigE cameras. Actively connecting with the camera through the device SN. One Mapper can support multiple cameras to be connected at the same time, which can be distinguished by the device SN. You can use the tool [gc_config](#gc_config) provided by rc_genicam_api to query the camera SN.
- Support the parameter configuration of GigE camera. Configure various parameters of the camera through the ymal file, and use the tool [gc_info](#gc_info) provided by rc_genicam_api to query the camera parameter name, data type and configurable range.
- Support camera capture function, support for exporting PNG and PNM image formats, and support the following pixel formats processing:

  - Monochrome pixel formats 
    - Mono8
    - Confidence8
    - Error8
  - Color pixel formats 
    - RGB8
    - BayerRG8
    - BayerBG8
    - BayerGR8
    - BayerGB8
    - YCbCr411_8
    - YCbCr422_8
    - YUV422_8


## How to use GigE Mapper

```shell
make mapper gige
source ~/.bashrc
sudo ldconfig
cd bin
./gige
```

## How to use rc_genicam_api tools

Refer to [rc_genicam_api](https://github.com/roboception/rc_genicam_api#readme) for details.Here are some common commands.

### gc_config

***gc_config -l***

List all available GigE Vision devices.

Output format: interface-id:serialNumber (displayName, device-id).

### gc_info

***gc_info serialNumber***

Display the configuration  file of the camera, including all camera parameter names, parameter values, data type, configurable range, etc. The camera [deviceinstance.yaml](./crd_example/deviceinstance.yaml) and [devicemodel.yaml](./crd_example/devicemodel.yaml) file can be modified according to this file.

***gc_info serialNumber key=value***

Modify the value of the attribute key to value, and display the configuration  file of the camera.

### gc_stream

***gc_stream -f png serialNumber n=1***

Grab the image and store it locally

-f set the picture format, can be connected to pnm or png, and the default is pnm.

n Set the number of snapshots, and the default is 1
