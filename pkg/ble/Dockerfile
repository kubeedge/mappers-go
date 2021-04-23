FROM alpine

RUN mkdir -p kubeedge

COPY ./config.yaml kubeedge/
COPY ./ble kubeedge/

WORKDIR kubeedge

CMD ["/kubeedge/ble","--v","5"]
