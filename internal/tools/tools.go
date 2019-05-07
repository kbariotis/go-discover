// +build tools

package tools

import (
	// See README.md in this package
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/mattn/goveralls"
	_ "github.com/maxbrunsfeld/counterfeiter/v6"
)
