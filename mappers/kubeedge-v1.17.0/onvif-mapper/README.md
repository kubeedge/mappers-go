This mapper is for the ONVIF IP camera. For resource-limited, it's only be tested with the HIKVISION camera.

Supported functions:
- Save frame. You can define it in device-instance.yaml and save the rtsp stream as video frame files.
- Save video. You can define it in device-instance.yaml and save the rtsp stream as video files.

steps:

1. Run onvif mapper 

   There are two ways to run onvif mapper:

a). Start locally
- Install the dependences:
    ```
    sudo apt-get update && 
    sudo apt-get install -y upx-ucl gcc-aarch64-linux-gnu libc6-dev-arm64-cross gcc-arm-linux-gnueabi libc6-dev-armel-cross libva-dev libva-drm2 libx11-dev libvdpau-dev libxext-dev libsdl1.2-dev libxcb1-dev libxau-dev libxdmcp-dev yasm
    ```
  and install ffmpeg with commond:
    ```
    sudo curl -sLO https://ffmpeg.org/releases/ffmpeg-4.1.6.tar.bz2 && 
    tar -jx --strip-components=1 -f ffmpeg-4.1.6.tar.bz2 &&  
    ./configure &&  make && 
    sudo make install
    ```
  This may take about 5 minutes to download and build all dependencies.
- Locally compile

  You can compile and run the mapper code directly:
    ```
    go run cmd/main.go --v <log level,like 3> --config-file <path to config yaml>
    ```
b). Start using a container image
- Build onvif mapper image:
    ```
    docker build -f Dockerfile_stream -t [YOUR MAPPER IMAGE NAME] .
    ```
  It may take about 8 minutes to build the docker image

- Deploy onvif mapper container:

    After successfully building the onvif mapper image, you can deploy the mapper in the cluster through deployment or other methods.
    A sample configuration file for mapper deployment is provided in the **resource** directory.

2. Build and submit the device yaml file:

  After successfully deploying onvif mapper, users can build the device-instance and device-model configuration files according to the 
  characteristics of the user edge onvif device, and execute the following commands to submit to the kubeedge cluster:
  ```
  kubectl apply -f <path to device model or device instance yaml>
  ```

  An example device-model and device-instance configuration file for onvif device is provided in the resource directory.
3. View log:

   Users can view the logs of the mapper container to determine whether the edge device is managed correctly.
