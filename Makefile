.PHONY: modbusmapper
modbusmapper:
	go build ./pkg/modbus
.PHONY: modbusmapper_image
modbusmapper_image:modbusmapper
	sudo docker build -t modbusmapper:v1.0 ./pkg/modbus

.PHONY: mappers
mappers:modbusmapper
