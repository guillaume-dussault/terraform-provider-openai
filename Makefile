HOSTNAME=registry.terraform.io
NAMESPACE=guillaume-dussault
NAME=openai
BINARY=terraform-provider-${NAME}
VERSION=1.0
OS_ARCH=darwin_arm64

default: install

build:
	go build -gcflags="all=-N -l" -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

generate:
	go mod tidy
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

.PHONY: generate
