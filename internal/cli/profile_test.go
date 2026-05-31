package cli

import (
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	viper.Reset()
	return filepath.Join(dir, ".octopus")
}

func TestAddProfile(t *testing.T) {
	setupTestDir(t)

	err := AddProfile("test", "http://localhost:8080", "table")
	if err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}

	profiles := GetProfiles()
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].Name != "test" {
		t.Errorf("expected name 'test', got %q", profiles[0].Name)
	}
	if profiles[0].BaseURL != "http://localhost:8080" {
		t.Errorf("expected base_url 'http://localhost:8080', got %q", profiles[0].BaseURL)
	}
	if !profiles[0].Active {
		t.Error("expected first profile to be active")
	}
}

func TestAddProfileDuplicate(t *testing.T) {
	setupTestDir(t)

	AddProfile("dup", "http://a.com", "table")
	AddProfile("dup", "http://b.com", "table")

	profiles := GetProfiles()
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile after duplicate, got %d", len(profiles))
	}
}

func TestSetActiveProfile(t *testing.T) {
	setupTestDir(t)

	AddProfile("first", "http://a.com", "table")
	AddProfile("second", "http://b.com", "json")

	SetActiveProfile("second")

	profiles := GetProfiles()
	var firstActive, secondActive bool
	for _, p := range profiles {
		if p.Name == "first" && p.Active {
			firstActive = true
		}
		if p.Name == "second" && p.Active {
			secondActive = true
		}
	}
	if firstActive {
		t.Error("first profile should not be active")
	}
	if !secondActive {
		t.Error("second profile should be active")
	}
}

func TestRemoveProfile(t *testing.T) {
	setupTestDir(t)

	AddProfile("keep", "http://a.com", "table")
	AddProfile("remove", "http://b.com", "table")

	RemoveProfile("remove")

	profiles := GetProfiles()
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].Name != "keep" {
		t.Errorf("expected 'keep', got %q", profiles[0].Name)
	}
}

func TestGetActiveProfile(t *testing.T) {
	setupTestDir(t)

	p := GetActiveProfile()
	if p != nil {
		t.Error("expected nil active profile when no profiles exist")
	}

	AddProfile("alpha", "http://a.com", "table")
	AddProfile("beta", "http://b.com", "json")

	p = GetActiveProfile()
	if p == nil {
		t.Fatal("expected non-nil active profile")
	}
	if p.Name != "alpha" {
		t.Errorf("expected first profile to be active, got %q", p.Name)
	}
}

func TestSessionPersistence(t *testing.T) {
	setupTestDir(t)

	s := GetSession()
	if s != nil {
		t.Error("expected nil session initially")
	}

	session := Session{
		ProfileName: "test",
		Username:    "admin",
		Token:       "tok_abc123",
		ExpireAt:    "2030-01-01",
	}
	SaveSession(session)

	loaded := GetSession()
	if loaded == nil {
		t.Fatal("expected non-nil session after save")
	}
	if loaded.ProfileName != "test" {
		t.Errorf("expected profile 'test', got %q", loaded.ProfileName)
	}
	if loaded.Token != "tok_abc123" {
		t.Errorf("expected token 'tok_abc123', got %q", loaded.Token)
	}

	ClearSession()
	cleared := GetSession()
	if cleared != nil {
		t.Error("expected nil session after clear")
	}
}
