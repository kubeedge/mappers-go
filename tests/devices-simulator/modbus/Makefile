SHELL := /bin/bash

curr_dir := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
rest_args := $(wordlist 2, $(words $(MAKECMDGOALS)), $(MAKECMDGOALS))
$(eval $(rest_args):;@:)

help:
	#
	# Usage:
	#   make device {protocol-name} <action> <parameter>:  execute device building process.
	#
	# Actions:
	#   -           mod, m  :  download code dependencies.
	#   -          lint, l  :  verify code via go fmt and `golangci-lint`.
	#   -         build, b  :  compile code.
	#
	# Parameters:
	#   ARM   : true or undefined
	#   ARM64 : true or undefined
	#
	# Example:
	#   -  make device modbus ARM64=true :  execute `build` "modbus" device for ARM64.
	#   -        make device modbus build :  execute `build` "modbus" device.
	@echo

make_rules := $(shell ls $(curr_dir)/hack/make-rules | sed 's/.sh//g')
$(make_rules):
	@$(curr_dir)/hack/make-rules/$@.sh $(rest_args)

.DEFAULT_GOAL := help
.PHONY: $(make_rules) build package