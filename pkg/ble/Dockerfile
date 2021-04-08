FROM alpine
RUN mkdir -p /opt/kubeedge/
ADD config.yaml /bin
ADD modbus-go /bin
CMD ["/bin/ble-go","--v","5"]
