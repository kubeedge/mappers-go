FROM ubuntu:16.04

RUN mkdir -p kubeedge

COPY ./bin/virtualdevice-dmi kubeedge/
COPY ./config.yaml kubeedge/

WORKDIR kubeedge

CMD ./virtualdevice-dmi
