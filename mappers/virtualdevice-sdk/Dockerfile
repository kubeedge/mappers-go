FROM alpine:3.14
RUN mkdir -p kubeedge
COPY ./bin/virtualdevice-sdk kubeedge/bin/
COPY ./res kubeedge/res/
WORKDIR kubeedge/bin
ENTRYPOINT ["./virtualdevice-sdk", "--v", "4"]
