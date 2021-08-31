The module is for the ONVIF IP camera. For resource-limited, it's only be tested with the HIKVISION camera.

Supported functions:
- Save frame. You could define the frame format, to be saved frame count(If not defined, it will be saved till the save flag to be set as false).
- Save video. You could define the video format, frame interval, and frame count.
- Reboot the camera. This is a standard ONVIF function. Users could expand those functions such as "get the date/time". More details, please read the code.

Notes:
- The password or certification files should be passed to the mapper. You could mount them in the docker file by "COPY" or mount as kubernetes secret.
