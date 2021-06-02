
FROM ubuntu:16.04

RUN mkdir -p kubeedge

COPY ./bin/onvif kubeedge/
COPY ./config.yaml kubeedge/

WORKDIR kubeedge

ENTRYPOINT ["/kubeedge/onvif", "--v", "5"]
