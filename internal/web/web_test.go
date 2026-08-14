package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func chdirRepoRoot(t *testing.T) {
	t.Helper()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
}

func TestHandlerServesIndex(t *testing.T) {
	chdirRepoRoot(t)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	contentType := recorder.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", contentType)
	}
	if !strings.Contains(recorder.Body.String(), "TODO") {
		t.Fatalf("body does not contain TODO: %s", recorder.Body.String())
	}
}

func TestHandlerServesStaticAsset(t *testing.T) {
	chdirRepoRoot(t)

	request := httptest.NewRequest(http.MethodGet, "/styles.css", nil)
	recorder := httptest.NewRecorder()

	Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "body") {
		t.Fatalf("styles.css body looks unexpected: %s", recorder.Body.String())
	}
}
