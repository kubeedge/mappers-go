# It's a stub docker file for onvif. We'll leverage it later soon.
FROM ubuntu:18.04

RUN mkdir -p /kubeedge
RUN mkdir -p /ca

COPY ./bin/onvif /kubeedge/
COPY ./config.yaml /kubeedge/
WORKDIR /kubeedge

ENTRYPOINT ["/kubeedge/onvif", "--v", "5"]
