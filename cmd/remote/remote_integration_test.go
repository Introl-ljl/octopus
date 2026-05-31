package remote

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bestruirui/octopus/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type testContext struct {
	server  *httptest.Server
	rootCmd *cobra.Command
	homeDir string
}

func setupTest(t *testing.T) *testContext {
	t.Helper()
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	viper.Reset()
	os.MkdirAll(filepath.Join(homeDir, ".octopus"), 0755)

	rootCmd := &cobra.Command{Use: "octopus"}
	rootCmd.AddCommand(RemoteCmd)

	return &testContext{
		homeDir: homeDir,
		rootCmd: rootCmd,
	}
}

func (tc *testContext) setupServer(t *testing.T) *httptest.Server {
	t.Helper()
	tc.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	viper.Set("profiles", []map[string]interface{}{
		{"name": "test", "base_url": tc.server.URL, "default_output": "table", "active": true},
	})
	viper.Set("session", map[string]interface{}{
		"profile_name": "test",
		"username":     "admin",
		"token":        "test-auth-token",
		"expire_at":    "2030-01-01",
	})
	viper.WriteConfigAs(filepath.Join(tc.homeDir, ".octopus", "config.yaml"))
	return tc.server
}

func (tc *testContext) run(args ...string) error {
	tc.rootCmd.SetArgs(args)
	return tc.rootCmd.Execute()
}

func (tc *testContext) runCapture(args ...string) (*cobra.Command, error) {
	tc.rootCmd.SetArgs(args)
	return tc.rootCmd.ExecuteC()
}

// -- Auth / Session tests --

func TestAuthLogin_ValidatesFlags(t *testing.T) {
	tc := setupTest(t)
	err := tc.run("remote", "login")
	if err == nil {
		t.Fatal("expected error for missing --username")
	}
}

func TestStatusCommand(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	requestPath := ""
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": map[string]interface{}{
				"username": "admin", "role": "admin",
			},
		})
	})
	err := tc.run("remote", "status")
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if requestPath != "/api/v1/user/status" {
		t.Errorf("expected /api/v1/user/status, got %s", requestPath)
	}
}

func TestLogoutCommand(t *testing.T) {
	tc := setupTest(t)
	// Setup profile and session for logout
	viper.Set("profiles", []map[string]interface{}{
		{"name": "test", "base_url": "http://localhost", "default_output": "table", "active": true},
	})
	viper.Set("session", map[string]interface{}{
		"profile_name": "test", "username": "admin", "token": "tok",
	})
	viper.WriteConfigAs(filepath.Join(tc.homeDir, ".octopus", "config.yaml"))

	err := tc.run("remote", "logout")
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}
	viper.Reset()
	os.Remove(filepath.Join(tc.homeDir, ".octopus", "config.yaml"))
	session := cli.GetSession()
	if session != nil {
		t.Error("session should be nil after logout")
	}
}

func TestChangePassword(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	requestPath := ""
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok",
		})
	})
	err := tc.run("remote", "user", "change-password",
		"--old-password", "old1", "--new-password", "new1")
	if err != nil {
		t.Fatalf("change-password failed: %v", err)
	}
	if requestPath != "/api/v1/user/change-password" {
		t.Errorf("expected /api/v1/user/change-password, got %s", requestPath)
	}
}

func TestChangeUsername(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	requestPath := ""
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok",
		})
	})
	err := tc.run("remote", "user", "change-username", "--new-username", "newadmin")
	if err != nil {
		t.Fatalf("change-username failed: %v", err)
	}
	if requestPath != "/api/v1/user/change-username" {
		t.Errorf("expected /api/v1/user/change-username, got %s", requestPath)
	}
}

func TestValidationErrorOnMissingFlags(t *testing.T) {
	tc := setupTest(t)

	err := tc.run("remote", "user", "change-password", "--old-password", "old1")
	if err == nil {
		t.Fatal("expected error for missing --new-password")
	}
}

func TestUnauthorizedWithoutSession(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	viper.Reset()
	os.MkdirAll(filepath.Join(homeDir, ".octopus"), 0755)

	viper.Set("profiles", []map[string]interface{}{
		{"name": "test", "base_url": "http://localhost", "default_output": "table", "active": true},
	})
	viper.WriteConfigAs(filepath.Join(homeDir, ".octopus", "config.yaml"))

	rootCmd := &cobra.Command{Use: "octopus"}
	rootCmd.AddCommand(RemoteCmd)
	rootCmd.SetArgs([]string{"remote", "channel", "list"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
	if !strings.Contains(err.Error(), "not authenticated") {
		t.Errorf("expected 'not authenticated', got: %v", err)
	}
}

func TestExitCodeInputOnMissingFlag(t *testing.T) {
	tc := setupTest(t)

	err := tc.run("remote", "channel", "create")
	if err == nil {
		t.Fatal("expected error for missing flags")
	}
	if cliErr, ok := err.(*cli.CLIError); ok {
		if cliErr.ExitCode != cli.ExitCodeInput {
			t.Errorf("expected ExitCodeInput (3), got %d", cliErr.ExitCode)
		}
	}
}

func TestExitCodeServerError(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 500, "message": "internal error",
		})
	})
	err := tc.run("remote", "channel", "list")
	if err == nil {
		t.Fatal("expected error for server error")
	}
	if cliErr, ok := err.(*cli.CLIError); ok {
		if cliErr.ExitCode != cli.ExitCodeServerError {
			t.Errorf("expected ExitCodeServerError (4), got %d", cliErr.ExitCode)
		}
	}
}

func TestJSONOutputFormat(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": []map[string]interface{}{
				{"name": "gpt-4", "input_price": 10, "output_price": 30},
			},
		})
	})

	cmd, err := tc.runCapture("remote", "model", "list", "--output", "json")
	if err != nil {
		t.Fatalf("model list with JSON output failed: %v", err)
	}
	out := cmd.OutOrStdout()
	_ = out
}

func TestJSONOutputApikeyStats(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": map[string]interface{}{"total_keys": 5, "active_keys": 3},
		})
	})
	err := tc.run("remote", "apikey", "stats", "--output", "json")
	if err != nil {
		t.Fatalf("apikey stats json failed: %v", err)
	}
}

// -- Channel commands --

func TestChannelList(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	verified := false
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/channel/list" {
			verified = true
		}
		if r.Header.Get("Authorization") != "Bearer test-auth-token" {
			t.Errorf("expected Bearer test-auth-token, got %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": []map[string]interface{}{
				{"id": 1, "name": "ch1", "type": 0, "enabled": true, "model": "gpt-4"},
			},
		})
	})
	err := tc.run("remote", "channel", "list")
	if err != nil {
		t.Fatalf("channel list failed: %v", err)
	}
	if !verified {
		t.Error("expected GET /api/v1/channel/list")
	}
}

func TestChannelCreate(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	verified := false
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/channel/create" {
			body, _ := io.ReadAll(r.Body)
			var data map[string]interface{}
			json.Unmarshal(body, &data)
			if data["name"] == "test-channel" {
				verified = true
			}
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": map[string]interface{}{"id": 1},
		})
	})
	err := tc.run("remote", "channel", "create",
		"--name", "test-channel", "--type", "0", "--model", "gpt-4",
		"--base-url", "https://api.openai.com", "--api-key", "sk-test")
	if err != nil {
		t.Fatalf("channel create failed: %v", err)
	}
	if !verified {
		t.Error("expected POST /api/v1/channel/create with name 'test-channel'")
	}
}

func TestChannelUpdate(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	verified := false
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/channel/update" {
			verified = true
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok",
		})
	})
	err := tc.run("remote", "channel", "update", "--id", "1", "--name", "updated")
	if err != nil {
		t.Fatalf("channel update failed: %v", err)
	}
	if !verified {
		t.Error("expected POST /api/v1/channel/update")
	}
}

func TestChannelEnable(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	verified := false
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/api/v1/channel/enable") {
			verified = true
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok",
		})
	})
	err := tc.run("remote", "channel", "enable", "--id", "1", "--enabled", "false")
	if err != nil {
		t.Fatalf("channel enable failed: %v", err)
	}
	if !verified {
		t.Error("expected POST /api/v1/channel/enable")
	}
}

func TestChannelDelete(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	verified := false
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/api/v1/channel/delete/1") {
			verified = true
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok",
		})
	})
	err := tc.run("remote", "channel", "delete", "--id", "1", "--force")
	if err != nil {
		t.Fatalf("channel delete failed: %v", err)
	}
	if !verified {
		t.Error("expected DELETE /api/v1/channel/delete/1")
	}
}

func TestChannelSync(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	verified := false
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/channel/sync" {
			verified = true
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok",
		})
	})
	err := tc.run("remote", "channel", "sync")
	if err != nil {
		t.Fatalf("channel sync failed: %v", err)
	}
	if !verified {
		t.Error("expected POST /api/v1/channel/sync")
	}
}

func TestChannelLastSyncTime(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": "2025-01-15T10:00:00Z",
		})
	})
	err := tc.run("remote", "channel", "last-sync-time")
	if err != nil {
		t.Fatalf("channel last-sync-time failed: %v", err)
	}
}

// -- Group commands --

func TestGroupList(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": []map[string]interface{}{
				{"id": 1, "name": "g1", "mode": 1, "match_regex": ".*"},
			},
		})
	})
	err := tc.run("remote", "group", "list")
	if err != nil {
		t.Fatalf("group list failed: %v", err)
	}
}

func TestGroupCreate(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": map[string]interface{}{"id": 1},
		})
	})
	err := tc.run("remote", "group", "create", "--name", "test-group", "--mode", "1")
	if err != nil {
		t.Fatalf("group create failed: %v", err)
	}
}

func TestGroupUpdate(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	verified := false
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/group/update" {
			verified = true
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok",
		})
	})
	err := tc.run("remote", "group", "update", "--id", "1", "--name", "renamed")
	if err != nil {
		t.Fatalf("group update failed: %v", err)
	}
	if !verified {
		t.Error("expected POST /api/v1/group/update")
	}
}

func TestGroupDelete(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	verified := false
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/api/v1/group/delete/1") {
			verified = true
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok",
		})
	})
	err := tc.run("remote", "group", "delete", "--id", "1", "--force")
	if err != nil {
		t.Fatalf("group delete failed: %v", err)
	}
	if !verified {
		t.Error("expected DELETE /api/v1/group/delete/1")
	}
}

// -- API Key commands --

func TestApikeyList(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": []map[string]interface{}{
				{"id": 1, "name": "my-key", "api_key": "sk-abc123", "enabled": true, "expire_at": ""},
			},
		})
	})
	err := tc.run("remote", "apikey", "list")
	if err != nil {
		t.Fatalf("apikey list failed: %v", err)
	}
}

func TestApikeyCreate(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": map[string]interface{}{"name": "my-key", "api_key": "sk-abc123"},
		})
	})
	err := tc.run("remote", "apikey", "create", "--name", "my-key")
	if err != nil {
		t.Fatalf("apikey create failed: %v", err)
	}
}

func TestApikeyUpdate(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	verified := false
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/apikey/update" {
			verified = true
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok",
		})
	})
	err := tc.run("remote", "apikey", "update", "--id", "1", "--name", "renamed")
	if err != nil {
		t.Fatalf("apikey update failed: %v", err)
	}
	if !verified {
		t.Error("expected POST /api/v1/apikey/update")
	}
}

func TestApikeyDelete(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	verified := false
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/api/v1/apikey/delete/1") {
			verified = true
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok",
		})
	})
	err := tc.run("remote", "apikey", "delete", "--id", "1", "--force")
	if err != nil {
		t.Fatalf("apikey delete failed: %v", err)
	}
	if !verified {
		t.Error("expected DELETE /api/v1/apikey/delete/1")
	}
}

func TestApikeyStats(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	verified := false
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/apikey/stats" {
			verified = true
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": map[string]interface{}{"total_keys": 5, "active_keys": 3},
		})
	})
	err := tc.run("remote", "apikey", "stats")
	if err != nil {
		t.Fatalf("apikey stats failed: %v", err)
	}
	if !verified {
		t.Error("expected GET /api/v1/apikey/stats")
	}
}

// -- Model commands --

func TestModelList(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": []map[string]interface{}{
				{"name": "gpt-4", "input_price": 10, "output_price": 30},
			},
		})
	})
	err := tc.run("remote", "model", "list")
	if err != nil {
		t.Fatalf("model list failed: %v", err)
	}
}

func TestModelChannel(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	verified := false
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/model/channel" {
			verified = true
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": []map[string]interface{}{
				{"name": "gpt-4", "channel_id": 1, "channel_name": "ch1", "enabled": true},
			},
		})
	})
	err := tc.run("remote", "model", "channel")
	if err != nil {
		t.Fatalf("model channel failed: %v", err)
	}
	if !verified {
		t.Error("expected GET /api/v1/model/channel")
	}
}

func TestModelCreate(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok",
		})
	})
	err := tc.run("remote", "model", "create", "--name", "gpt-4", "--input", "10", "--output", "30")
	if err != nil {
		t.Fatalf("model create failed: %v", err)
	}
}

func TestModelUpdate(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	verified := false
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/model/update" {
			verified = true
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok",
		})
	})
	err := tc.run("remote", "model", "update", "--name", "gpt-4", "--input", "15")
	if err != nil {
		t.Fatalf("model update failed: %v", err)
	}
	if !verified {
		t.Error("expected POST /api/v1/model/update")
	}
}

func TestModelDelete(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	verified := false
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/model/delete" {
			verified = true
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok",
		})
	})
	err := tc.run("remote", "model", "delete", "--name", "gpt-4")
	if err != nil {
		t.Fatalf("model delete failed: %v", err)
	}
	if !verified {
		t.Error("expected POST /api/v1/model/delete")
	}
}

func TestModelUpdatePrice(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok",
		})
	})
	err := tc.run("remote", "model", "update-price")
	if err != nil {
		t.Fatalf("model update-price failed: %v", err)
	}
}

func TestModelLastUpdateTime(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": "2025-01-15T10:00:00Z",
		})
	})
	err := tc.run("remote", "model", "last-update-time")
	if err != nil {
		t.Fatalf("model last-update-time failed: %v", err)
	}
}

// -- Setting commands --

func TestSettingList(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": []map[string]interface{}{
				{"key": "max_tokens", "value": "4096"},
			},
		})
	})
	err := tc.run("remote", "setting", "list")
	if err != nil {
		t.Fatalf("setting list failed: %v", err)
	}
}

func TestSettingSet(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	verified := false
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/setting/set" {
			verified = true
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok",
		})
	})
	err := tc.run("remote", "setting", "set", "--key", "max_tokens", "--value", "8192")
	if err != nil {
		t.Fatalf("setting set failed: %v", err)
	}
	if !verified {
		t.Error("expected POST /api/v1/setting/set")
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": map[string]interface{}{"reset_count": 1},
		})
	})
	err := tc.run("remote", "setting", "reset-circuit-breaker")
	if err != nil {
		t.Fatalf("circuit breaker reset failed: %v", err)
	}
}

// -- Stats commands --

func TestStatsToday(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": map[string]interface{}{"requests": 100, "tokens": 5000},
		})
	})
	err := tc.run("remote", "stats", "today")
	if err != nil {
		t.Fatalf("stats today failed: %v", err)
	}
}

func TestStatsDaily(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": []map[string]interface{}{},
		})
	})
	err := tc.run("remote", "stats", "daily")
	if err != nil {
		t.Fatalf("stats daily failed: %v", err)
	}
}

func TestStatsHourly(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": []map[string]interface{}{},
		})
	})
	err := tc.run("remote", "stats", "hourly")
	if err != nil {
		t.Fatalf("stats hourly failed: %v", err)
	}
}

func TestStatsTotal(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": map[string]interface{}{"requests": 10000},
		})
	})
	err := tc.run("remote", "stats", "total")
	if err != nil {
		t.Fatalf("stats total failed: %v", err)
	}
}

func TestStatsApikey(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": []map[string]interface{}{},
		})
	})
	err := tc.run("remote", "stats", "apikey")
	if err != nil {
		t.Fatalf("stats apikey failed: %v", err)
	}
}

func TestStatsModel(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": []map[string]interface{}{},
		})
	})
	err := tc.run("remote", "stats", "model")
	if err != nil {
		t.Fatalf("stats model failed: %v", err)
	}
}

// -- Log commands --

func TestLogList(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": []map[string]interface{}{
				{"id": 1, "request_model_name": "gpt-4", "channel_name": "ch1", "input_tokens": 10, "output_tokens": 20, "use_time": 500, "cost": 0.01},
			},
		})
	})
	err := tc.run("remote", "log", "list")
	if err != nil {
		t.Fatalf("log list failed: %v", err)
	}
}

func TestLogListPagination(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok", "data": []map[string]interface{}{},
		})
	})
	err := tc.run("remote", "log", "list", "--page", "2", "--page-size", "50")
	if err != nil {
		t.Fatalf("log list with pagination failed: %v", err)
	}
}

func TestLogClear(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	verified := false
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/api/v1/log/clear") {
			verified = true
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0, "message": "ok",
		})
	})
	err := tc.run("remote", "log", "clear", "--force")
	if err != nil {
		t.Fatalf("log clear failed: %v", err)
	}
	if !verified {
		t.Error("expected DELETE /api/v1/log/clear")
	}
}

// -- Backup commands --

func TestBackupExport(t *testing.T) {
	tc := setupTest(t)
	srv := tc.setupServer(t)
	defer srv.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":"test backup"}`))
	})
	err := tc.run("remote", "backup", "export")
	if err != nil {
		t.Fatalf("backup export failed: %v", err)
	}
	os.Remove("octopus-export.json")
}

// -- Profile commands --

func TestProfileAdd(t *testing.T) {
	tc := setupTest(t)
	err := tc.run("remote", "profile", "add", "--name", "integ-test", "--url", "http://test.local", "--output", "json")
	if err != nil {
		t.Fatalf("profile add failed: %v", err)
	}
	profiles := cli.GetProfiles()
	found := false
	for _, p := range profiles {
		if p.Name == "integ-test" {
			found = true
			if p.BaseURL != "http://test.local" {
				t.Errorf("expected BaseURL 'http://test.local', got %q", p.BaseURL)
			}
		}
	}
	if !found {
		t.Fatal("profile 'integ-test' not found after add")
	}
}

func TestProfileList(t *testing.T) {
	tc := setupTest(t)
	viper.Set("profiles", []map[string]interface{}{
		{"name": "p1", "base_url": "http://a.com", "active": true},
	})
	viper.WriteConfigAs(filepath.Join(tc.homeDir, ".octopus", "config.yaml"))
	err := tc.run("remote", "profile", "list")
	if err != nil {
		t.Fatalf("profile list failed: %v", err)
	}
}

// -- Command validation tests (no HTTP call needed) --

func TestChannelCreateValidatesFlags(t *testing.T) {
	tc := setupTest(t)
	err := tc.run("remote", "channel", "create", "--type", "0")
	if err == nil {
		t.Fatal("expected error for missing --name")
	}
}

func TestChannelUpdateValidatesID(t *testing.T) {
	tc := setupTest(t)
	err := tc.run("remote", "channel", "update", "--name", "x")
	if err == nil {
		t.Fatal("expected error for missing --id")
	}
}

func TestChannelDeleteValidatesID(t *testing.T) {
	tc := setupTest(t)
	err := tc.run("remote", "channel", "delete")
	if err == nil {
		t.Fatal("expected error for missing --id")
	}
}

func TestGroupCreateValidatesFlags(t *testing.T) {
	tc := setupTest(t)
	err := tc.run("remote", "group", "create", "--mode", "1")
	if err == nil {
		t.Fatal("expected error for missing --name")
	}
}

func TestGroupUpdateValidatesID(t *testing.T) {
	tc := setupTest(t)
	err := tc.run("remote", "group", "update", "--name", "x")
	if err == nil {
		t.Fatal("expected error for missing --id")
	}
}

func TestGroupDeleteValidatesID(t *testing.T) {
	tc := setupTest(t)
	err := tc.run("remote", "group", "delete")
	if err == nil {
		t.Fatal("expected error for missing --id")
	}
}

func TestApikeyCreateValidatesFlags(t *testing.T) {
	tc := setupTest(t)
	err := tc.run("remote", "apikey", "create")
	if err == nil {
		t.Fatal("expected error for missing --name")
	}
}

func TestApikeyUpdateValidatesID(t *testing.T) {
	tc := setupTest(t)
	err := tc.run("remote", "apikey", "update", "--name", "x")
	if err == nil {
		t.Fatal("expected error for missing --id")
	}
}

func TestApikeyDeleteValidatesID(t *testing.T) {
	tc := setupTest(t)
	err := tc.run("remote", "apikey", "delete")
	if err == nil {
		t.Fatal("expected error for missing --id")
	}
}

func TestModelCreateValidatesFlags(t *testing.T) {
	tc := setupTest(t)
	err := tc.run("remote", "model", "create", "--input", "10")
	if err == nil {
		t.Fatal("expected error for missing --name")
	}
}

func TestModelUpdateValidatesName(t *testing.T) {
	tc := setupTest(t)
	err := tc.run("remote", "model", "update", "--input", "10")
	if err == nil {
		t.Fatal("expected error for missing --name")
	}
}

func TestModelDeleteValidatesName(t *testing.T) {
	tc := setupTest(t)
	err := tc.run("remote", "model", "delete")
	if err == nil {
		t.Fatal("expected error for missing --name")
	}
}

func TestSettingSetValidatesFlags(t *testing.T) {
	tc := setupTest(t)
	err := tc.run("remote", "setting", "set", "--key", "k")
	if err == nil {
		t.Fatal("expected error for missing --value")
	}
}

func TestBackupImportValidatesFlags(t *testing.T) {
	tc := setupTest(t)
	err := tc.run("remote", "backup", "import")
	if err == nil {
		t.Fatal("expected error for missing --file")
	}
}

func TestUserChangePassValidatesFlags(t *testing.T) {
	tc := setupTest(t)
	err := tc.run("remote", "user", "change-password", "--old-password", "old")
	if err == nil {
		t.Fatal("expected error for missing --new-password")
	}
}

func TestUserChangeNameValidatesFlags(t *testing.T) {
	tc := setupTest(t)
	err := tc.run("remote", "user", "change-username")
	if err == nil {
		t.Fatal("expected error for missing --new-username")
	}
}
