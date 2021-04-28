
FROM ubuntu:16.04

RUN mkdir -p kubeedge

COPY ./bin/ble kubeedge/
COPY ./config.yaml kubeedge/

WORKDIR kubeedge

ENTRYPOINT ["/kubeedge/ble", "--v", "5"]
