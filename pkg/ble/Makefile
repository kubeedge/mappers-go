SHELL = /bin/bash

build:
	@GOOS=linux CGO_ENABLED=0 GOARCH=arm go build -o ble-go # please give your own arch, the arm here is a example
	@docker build -t gopherlv/ble-go .  # please give your own image repo here
	@rm ble-go

push: build
	@$(eval REV := $(shell git rev-parse HEAD|cut -c 1-8))
	@docker tag gopherlv/ble-go gopherlv/ke-ble-mapper:$(REV) # please tag your own name of image
	@docker push gopherlv/ke-ble-mapper:$(REV)

