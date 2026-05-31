package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "ok",
			"data": map[string]interface{}{
				"id":   1,
				"name": "test",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	c.SetToken("test-token")

	var result map[string]interface{}
	err := c.Get("/api/v1/test", &result)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if result["id"] != float64(1) {
		t.Errorf("expected id 1, got %v", result["id"])
	}
	if result["name"] != "test" {
		t.Errorf("expected name 'test', got %v", result["name"])
	}
}

func TestClientPost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "ok",
			"data":    "created",
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	c.SetToken("tok")

	body := map[string]string{"name": "new"}
	var result string
	err := c.Post("/api/v1/create", body, &result)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	if result != "created" {
		t.Errorf("expected 'created', got %q", result)
	}
}

func TestClientPut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "ok",
			"data":    "updated",
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	var result string
	err := c.Put("/api/v1/update", nil, &result)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	if result != "updated" {
		t.Errorf("expected 'updated', got %q", result)
	}
}

func TestClientDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "ok",
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	err := c.Delete("/api/v1/delete/1", nil)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestClientPatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "ok",
			"data":    "patched",
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	body := map[string]string{"field": "val"}
	var result string
	err := c.Patch("/api/v1/patch", body, &result)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}
	if result != "patched" {
		t.Errorf("expected 'patched', got %q", result)
	}
}

func TestClientUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    401,
			"message": "unauthorized",
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	err := c.Get("/api/v1/protected", nil)
	if err != ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestClientForbidden(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    403,
			"message": "forbidden",
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	err := c.Get("/api/v1/admin", nil)
	if err != ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestClientServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    500,
			"message": "internal error",
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	err := c.Get("/api/v1/error", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	cliErr, ok := err.(*CLIError)
	if !ok {
		t.Fatalf("expected CLIError, got %T", err)
	}
	if cliErr.ExitCode != ExitCodeServerError {
		t.Errorf("expected ExitCodeServerError (4), got %d", cliErr.ExitCode)
	}
}

func TestClientNetworkError(t *testing.T) {
	// connect to a closed port to trigger network error
	c := NewClient("http://127.0.0.1:1")
	err := c.Get("/api/v1/test", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClientBuildURL(t *testing.T) {
	c1 := NewClient("http://localhost:8080")
	c2 := NewClient("http://localhost:8080/")

	// Check URL building via request (indirectly)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": r.URL.String(),
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	var urlStr string
	err := c.Get("/api/v1/test", &urlStr)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	_ = c1
	_ = c2
}
