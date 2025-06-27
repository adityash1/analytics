# Define the root directory
ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

# Export environment variables from .env file
include $(ROOT_DIR)/.env
export

.PHONY: server dashboard

server:
	@cd cmd/tracker && go build && ./tracker -ip 123.123.123.123

dashboard:
	@cd cmd/cli && \
	go build -o localdash && \
	env $$(cat ../../.env | xargs) ./localdash -site 1 -start 20240625 -end 20240627
