FROM ubuntu:16.04

RUN mkdir -p kubeedge

COPY bluetooth kubeedge/

COPY configuration/config.yaml kubeedge/configuration/config.yaml 

WORKDIR kubeedge

ENTRYPOINT ["/kubeedge/bluetooth","--logtostderr=true"]
