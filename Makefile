.PHONY: all
all: modbusmapper opcuamapper

.PHONY: modbusmapper
modbusmapper:
	go build -o ./pkg/modbus/modbus ./pkg/modbus

.PHONY: modbusmapper_image
modbusmapper_image:modbusmapper
	sudo docker build -t modbusmapper:v1.0 ./pkg/modbus

.PHONY: opcuamapper
opcuamapper:
	go build -o ./pkg/opcua/opcua ./pkg/opcua

.PHONY: opcuamapper_image
opcuamapper_image:opcuamapper
	sudo docker build -t opcuamapper:v1.0 ./pkg/opcua

clean:
	rm -f ./pkg/modbus/modbus ./pkg/opcua/opcua

define VERIFY_HELP_INFO
# verify golang,vendor and codegen
#
# Example:
# make verify
endef
.PHONY: verify
ifeq ($(HELP),y)
verify:
	@echo "$$VERIFY_HELP_INFO"
else
verify:verify-golang
endif

.PHONY: verify-golang
verify-golang:
	hack/verify-golang.sh


define LINT_HELP_INFO
# run golang lint check.
#
# Example:
#   make lint
#   make lint HELP=y
endef
.PHONY: lint
ifeq ($(HELP),y)
lint:
	@echo "$$LINT_HELP_INFO"
else
lint:
	hack/make-rules/lint.sh
endif

