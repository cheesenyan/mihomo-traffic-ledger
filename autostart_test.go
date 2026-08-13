package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAutostartSettingsHandlerReadsAndWritesBackend(t *testing.T) {
	originalRead := readAutostartEnabled
	originalWrite := writeAutostartEnabled
	t.Cleanup(func() {
		readAutostartEnabled = originalRead
		writeAutostartEnabled = originalWrite
	})

	enabled := true
	readAutostartEnabled = func() (bool, error) { return enabled, nil }
	writeAutostartEnabled = func(next bool) error {
		enabled = next
		return nil
	}

	svc := newTestService(t)
	for _, test := range []struct {
		method string
		body   string
		want   string
	}{
		{method: http.MethodGet, want: `{"enabled":true}`},
		{method: http.MethodPut, body: `{"enabled":false}`, want: `{"enabled":false}`},
		{method: http.MethodPut, body: `{"enabled":true}`, want: `{"enabled":true}`},
	} {
		req := httptest.NewRequest(test.method, "/api/settings/autostart", strings.NewReader(test.body))
		rec := httptest.NewRecorder()
		svc.handleAutostartSettings(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status=%d body=%s", test.method, rec.Code, rec.Body.String())
		}
		if got := strings.TrimSpace(rec.Body.String()); got != test.want {
			t.Fatalf("%s response=%s want=%s", test.method, got, test.want)
		}
	}
}

func TestAutostartSettingsHandlerRejectsInvalidInputAndBackendErrors(t *testing.T) {
	originalRead := readAutostartEnabled
	originalWrite := writeAutostartEnabled
	t.Cleanup(func() {
		readAutostartEnabled = originalRead
		writeAutostartEnabled = originalWrite
	})

	svc := newTestService(t)
	readAutostartEnabled = func() (bool, error) { return false, errors.New("read failed") }
	rec := httptest.NewRecorder()
	svc.handleAutostartSettings(rec, httptest.NewRequest(http.MethodGet, "/api/settings/autostart", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("read error status=%d", rec.Code)
	}

	readAutostartEnabled = func() (bool, error) { return false, nil }
	writeAutostartEnabled = func(bool) error { return errors.New("write failed") }
	for _, body := range []string{`{`, `{"enabled":true}`} {
		rec = httptest.NewRecorder()
		svc.handleAutostartSettings(rec, httptest.NewRequest(http.MethodPut, "/api/settings/autostart", strings.NewReader(body)))
		if rec.Code < http.StatusBadRequest {
			t.Fatalf("body=%q status=%d", body, rec.Code)
		}
	}
}

func TestRoutesExposeAutostartSettings(t *testing.T) {
	svc := newTestService(t)
	req := httptest.NewRequest(http.MethodGet, "/api/settings/autostart", nil)
	rec := httptest.NewRecorder()
	svc.routes().ServeHTTP(rec, req)
	if rec.Code == http.StatusNotFound {
		t.Fatal("autostart settings route is missing")
	}
}
