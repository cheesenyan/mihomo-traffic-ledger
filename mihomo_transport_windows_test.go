//go:build windows

package main

import (
	"context"
	"errors"
	"net"
	"reflect"
	"testing"
)

func TestOrderMihomoNamedPipeCandidates(t *testing.T) {
	requested := `\\.\pipe\verge-mihomo`
	discovered := []string{
		`\\.\pipe\verge-mihomo-sidecar-dev-bbb`,
		`\\.\pipe\unrelated`,
		`\\.\pipe\verge-mihomo-sidecar-release-ccc`,
		`\\.\pipe\verge-mihomo-production-aaa`,
		requested,
	}
	want := []string{
		`\\.\pipe\verge-mihomo-production-aaa`,
		`\\.\pipe\verge-mihomo-sidecar-release-ccc`,
		`\\.\pipe\verge-mihomo-sidecar-dev-bbb`,
	}
	if got := orderMihomoNamedPipeCandidates(requested, discovered); !reflect.DeepEqual(got, want) {
		t.Fatalf("candidates = %#v, want %#v", got, want)
	}
}

func TestDialMihomoNamedPipeAutoDiscoversProductionPipe(t *testing.T) {
	requested := `\\.\pipe\verge-mihomo`
	production := `\\.\pipe\verge-mihomo-production-owner`
	var attempts []string
	dial := func(_ context.Context, path string) (net.Conn, error) {
		attempts = append(attempts, path)
		if path != production {
			return nil, errors.New("not found")
		}
		client, server := net.Pipe()
		server.Close()
		return client, nil
	}
	conn, err := dialMihomoNamedPipe(context.Background(), requested, true, func() ([]string, error) {
		return []string{production}, nil
	}, dial)
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	want := []string{requested, production}
	if !reflect.DeepEqual(attempts, want) {
		t.Fatalf("attempts = %#v, want %#v", attempts, want)
	}
}

func TestDialMihomoNamedPipeExplicitPathDoesNotDiscover(t *testing.T) {
	requested := `\\.\pipe\custom-mihomo`
	listed := false
	_, err := dialMihomoNamedPipe(context.Background(), requested, false, func() ([]string, error) {
		listed = true
		return []string{`\\.\pipe\verge-mihomo-production-owner`}, nil
	}, func(context.Context, string) (net.Conn, error) {
		return nil, errors.New("not found")
	})
	if err == nil {
		t.Fatal("expected dial error")
	}
	if listed {
		t.Fatal("explicit named pipe path should not trigger discovery")
	}
}
