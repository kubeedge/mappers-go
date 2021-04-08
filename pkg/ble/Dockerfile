FROM alpine
RUN mkdir -p /opt/kubeedge/
ADD config.yaml /bin
ADD ble-go /bin
CMD ["/bin/ble-go","--v","5"]
