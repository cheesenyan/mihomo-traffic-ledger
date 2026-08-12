package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultWindowsMihomoEndpoint = "npipe://./pipe/verge-mihomo"

func resolveMihomoTransport(endpoint string, fallback *http.Client) (string, *http.Client, error) {
	endpoint = strings.TrimSpace(endpoint)
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", nil, err
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		if fallback == nil {
			fallback = &http.Client{Timeout: 10 * time.Second}
		}
		return strings.TrimRight(endpoint, "/"), fallback, nil
	case "npipe":
		pipePath := `\\.\pipe\` + strings.TrimPrefix(strings.TrimPrefix(parsed.Path, "/pipe/"), "/")
		if parsed.Host != "." || pipePath == `\\.\pipe\` {
			return "", nil, fmt.Errorf("invalid named pipe endpoint %q", endpoint)
		}
		client, err := newNamedPipeHTTPClient(pipePath)
		if err != nil {
			return "", nil, err
		}
		return "http://mihomo", client, nil
	default:
		return "", nil, fmt.Errorf("unsupported Mihomo endpoint scheme %q", parsed.Scheme)
	}
}
