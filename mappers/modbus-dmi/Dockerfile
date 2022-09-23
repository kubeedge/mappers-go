FROM ubuntu:16.04

ENV MQTT_SERVER=tcp://127.0.0.1:1883
ENV METASERVER_ADDR="http://127.0.0.1:10550"

RUN mkdir -p kubeedge

COPY ./bin/modbus kubeedge/
COPY ./config.yaml kubeedge/

WORKDIR kubeedge

CMD ./modbus --v 5 --mqtt-address $MQTT_SERVER --metaserver-addr $METASERVER_ADDR
