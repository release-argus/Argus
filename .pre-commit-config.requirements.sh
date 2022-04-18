#!/bin/bash

# go-revive
go install github.com/mgechev/revive@latest
# go-sec
go install github.com/securego/gosec/v2/cmd/gosec@latest
# go-structslop
go install github.com/orijtech/structslop/cmd/structslop@latest
# golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.45.2
