FROM scratch

WORKDIR /
# -- for modbus
EXPOSE 5020
COPY ./bin/modbusDevice /simulator
ENTRYPOINT ["/simulator"]

