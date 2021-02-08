FROM ubuntu:16.04

RUN mkdir -p kubeedge ca

COPY ./opcua kubeedge/
COPY ./config.yaml kubeedge/
COPY ./ca /ca

WORKDIR kubeedge

ENTRYPOINT ["/kubeedge/opcua", "--v", "5"]
