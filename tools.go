//go:build tools
// +build tools

package tools

// This file ensures build tool dependencies are tracked in go.mod
// See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint" // v1.55.2
)

