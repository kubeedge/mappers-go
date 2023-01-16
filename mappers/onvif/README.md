The module is for the ONVIF IP camera. For resource-limited, it's only be tested with the HIKVISION camera.

Supported functions:
- Save frame. You could define the frame format, to be saved frame count(If not defined, it will be saved till the save flag to be set as false).
- Save video. You could define the video format, frame interval, and frame count.
- Reboot the camera. This is a standard ONVIF function. Users could expand those functions such as "get the date/time". More details, please read the code.

Notes:
- The password or certification files should be passed to the mapper. You could mount them in the docker file by "COPY" or mount as kubernetes secret.

- Build a mapper of `onvif` with command:
    ```
    make mapper onvif build
    ```
    and find the binary in `mappers-go/mappers/onvif/bin/`

- If you fail to build the mapper of `onvif`, please install the dependences with command:
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
- Build an image of onvif mapper with command:
    ```
    make mapper onvif package
    ```
    and get the infomation of the image with command:
    ```
    docker images onvif-mapper
    ```