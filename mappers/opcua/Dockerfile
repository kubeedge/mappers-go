
FROM ubuntu:16.04

RUN mkdir -p kubeedge

COPY ./bin/opcua kubeedge/
COPY ./config.yaml kubeedge/

WORKDIR kubeedge

ENTRYPOINT ["/kubeedge/opcua", "--v", "5"]
