
FROM ubuntu:16.04

RUN mkdir -p kubeedge

COPY ./bin/banyan kubeedge/
COPY ./config.yaml kubeedge/

WORKDIR kubeedge

ENTRYPOINT ["/kubeedge/banyan", "--v", "5"]
