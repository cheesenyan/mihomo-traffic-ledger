//go:build !windows

package main

import (
	"fmt"
	"net/http"
)

func defaultMihomoEndpoint() string { return "" }

func newNamedPipeHTTPClient(pipePath string) (*http.Client, error) {
	return nil, fmt.Errorf("named pipe endpoint %q is only supported on Windows", pipePath)
}
