
FROM ubuntu:16.04

RUN mkdir -p kubeedge

COPY ./bin/Template kubeedge/
COPY ./config.yaml kubeedge/

WORKDIR kubeedge

ENTRYPOINT ["/kubeedge/Template", "--v", "5"]
