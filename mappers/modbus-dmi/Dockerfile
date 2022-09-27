FROM ubuntu:16.04

RUN mkdir -p kubeedge

COPY ./bin/modbus-dmi kubeedge/
COPY ./config.yaml kubeedge/

WORKDIR kubeedge

CMD ./modbus-dmi
