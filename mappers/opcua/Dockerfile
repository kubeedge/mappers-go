
FROM ubuntu:16.04

RUN mkdir -p kubeedge

RUN mkdir -p ca
COPY ./ca/*  /ca/


COPY ./bin/opcua kubeedge/
COPY ./config.yaml kubeedge/

WORKDIR kubeedge

ENTRYPOINT ["/kubeedge/opcua", "--v", "5"]
