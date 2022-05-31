FROM golang:1.16-alpine as builder
ENV GOPROXY=https://goproxy.io,direct
ENV GO111MODULE=on
ENV ENVGOCACHE=/go/pkg/.cache/go-build

ENV MQTT_SERVER=tcp://127.0.0.1:1883 SOCKET_ID="" SOCKET_USER="" DEVICE_ID="" MAX_HUMIDITY="" MIN_HUMIDITY=""
ENV METASERVER_ENABLE=false METASERVER_ADDR="http://127.0.0.1:10550"

WORKDIR /work
ADD . .
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix cgo -o app mappers/modbus/cmd/main.go

FROM alpine
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone
COPY --from=builder /work/app /work/app
COPY --from=builder /work/mappers/modbus/config/config.yaml /work/config.yaml

WORKDIR /work

CMD ./app --v 5 --mqtt-address $MQTT_SERVER --metaserver-enable $METASERVER_ENABLE --metaserver-addr $METASERVER_ADDR
