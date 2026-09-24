//go:build windows

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"
)

func defaultMihomoEndpoint() string { return defaultWindowsMihomoEndpoint }

type namedPipeDialFunc func(context.Context, string) (net.Conn, error)
type namedPipeListFunc func() ([]string, error)

var resolvedMihomoPipeLog struct {
	sync.Mutex
	path string
}

func newNamedPipeHTTPClient(pipePath string, autoDiscover bool) (*http.Client, error) {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dialMihomoNamedPipe(ctx, pipePath, autoDiscover, discoverMihomoNamedPipes, winio.DialPipeContext)
		},
	}
	return &http.Client{Transport: transport, Timeout: 10 * time.Second}, nil
}

func dialMihomoNamedPipe(ctx context.Context, requested string, autoDiscover bool, list namedPipeListFunc, dial namedPipeDialFunc) (net.Conn, error) {
	conn, requestedErr := dial(ctx, requested)
	if requestedErr == nil {
		logResolvedMihomoPipe(requested, requested)
		return conn, nil
	}
	if !autoDiscover {
		return nil, requestedErr
	}

	discovered, listErr := list()
	if listErr != nil {
		return nil, fmt.Errorf("open Mihomo named pipe %s: %w", requested, errors.Join(requestedErr, listErr))
	}
	candidates := orderMihomoNamedPipeCandidates(requested, discovered)
	var dialErrors []error
	dialErrors = append(dialErrors, fmt.Errorf("%s: %w", requested, requestedErr))
	for _, candidate := range candidates {
		conn, err := dial(ctx, candidate)
		if err == nil {
			logResolvedMihomoPipe(requested, candidate)
			return conn, nil
		}
		dialErrors = append(dialErrors, fmt.Errorf("%s: %w", candidate, err))
	}
	return nil, fmt.Errorf("no usable Mihomo named pipe found: %w", errors.Join(dialErrors...))
}

func orderMihomoNamedPipeCandidates(requested string, discovered []string) []string {
	seen := map[string]bool{strings.ToLower(requested): true}
	ordered := make([]string, 0, len(discovered))
	for _, candidate := range discovered {
		candidate = strings.TrimSpace(candidate)
		lower := strings.ToLower(candidate)
		if candidate == "" || seen[lower] || !strings.HasPrefix(lower, `\\.\pipe\verge-mihomo-`) {
			continue
		}
		seen[lower] = true
		ordered = append(ordered, candidate)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		leftRank := mihomoNamedPipeRank(ordered[i])
		rightRank := mihomoNamedPipeRank(ordered[j])
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		return strings.ToLower(ordered[i]) < strings.ToLower(ordered[j])
	})
	return ordered
}

func mihomoNamedPipeRank(path string) int {
	lower := strings.ToLower(path)
	switch {
	case strings.Contains(lower, `\verge-mihomo-production-`):
		return 0
	case strings.Contains(lower, `\verge-mihomo-sidecar-release-`):
		return 1
	case strings.Contains(lower, `\verge-mihomo-sidecar-dev-`):
		return 2
	default:
		return 3
	}
}

func discoverMihomoNamedPipes() ([]string, error) {
	pattern, err := windows.UTF16PtrFromString(`\\.\pipe\verge-mihomo-*`)
	if err != nil {
		return nil, err
	}
	var data windows.Win32finddata
	handle, err := windows.FindFirstFile(pattern, &data)
	if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer windows.FindClose(handle)

	pipes := make([]string, 0, 4)
	for {
		name := windows.UTF16ToString(data.FileName[:])
		if name != "" && name != "." && name != ".." {
			pipes = append(pipes, `\\.\pipe\`+name)
		}
		err = windows.FindNextFile(handle, &data)
		if errors.Is(err, windows.ERROR_NO_MORE_FILES) {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return pipes, nil
}

func logResolvedMihomoPipe(requested, resolved string) {
	resolvedMihomoPipeLog.Lock()
	defer resolvedMihomoPipeLog.Unlock()
	if resolvedMihomoPipeLog.path == resolved {
		return
	}
	resolvedMihomoPipeLog.path = resolved
	if !strings.EqualFold(requested, resolved) {
		log.Printf("Mihomo named pipe auto-discovered: %s", resolved)
	}
}
