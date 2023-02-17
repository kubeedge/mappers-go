.PHONY: build test clean docker
help:
	#
	# Usage:
	#   make build  :  compile main.go
	#   make clean  :  clean binaries
	#   make docker :  pack binaries to image
	@echo
build:
	go build -o ./bin/idmvs ./cmd/main.go
	@echo  "[INFO] go build successful"
	@echo  "[INFO] you can cd ./bin to execute it"
clean:
	rm -rf ./bin/*
docker:
	docker build -f ./Dockerfile -t mapper-idmvs:v0.0.1 .
