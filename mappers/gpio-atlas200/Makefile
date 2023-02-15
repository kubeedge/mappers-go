.PHONY: build test clean docker
help:
	#
	# Usage:
	#   make build  :  compile main.go
	#   make clean  :  clean binaries
	#   make docker :  pack binaries to image
	@echo
build:
	go build -o ./bin/gpio  ./cmd/main.go
	@echo  "[INFO] go build successful"
	@echo  "[INFO] you can cd ./bin to execute it"
clean:
	rm -rf ./bin/*
docker:
	docker build -f ./Dockerfile docker build -f ./Dockerfile -t mapper-gpio-device:v0.0.1  .
