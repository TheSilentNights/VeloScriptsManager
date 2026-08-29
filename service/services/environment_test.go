package services

import (
	"path/filepath"
	"strings"
	"testing"

	"github/TheSilentNights/VeloScriptsManager/service/storage"
)

func newTestService(t *testing.T) *Server {
	t.Helper()
	db, err := storage.OpenOrCreate(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewServerController(storage.CreateScriptRepo(db), storage.CreateEnvironmentRepo(db))
}

func envAsMap(env []string) map[string]string {
	out := make(map[string]string, len(env))
	for _, kv := range env {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			out[parts[0]] = parts[1]
		}
	}
	return out
}

func TestResolveEnvironmentVars(t *testing.T) {
	service := newTestService(t)

	_ = service.environmentRepo.Upsert(storage.Environment{ID: "A", Name: "A", Env: []storage.EnvVar{{Key: "K1", Value: "v1"}, {Key: "K2", Value: "v2"}}, Paths: []string{`C:\jdk\bin`}})
	_ = service.environmentRepo.Upsert(storage.Environment{ID: "B", Name: "B", Env: []storage.EnvVar{{Key: "K1", Value: "override"}}})

	env, apiErr := service.resolveEnvironmentVars([]string{"A", "B"})
	if apiErr != nil {
		t.Fatalf("unexpected error: %v", apiErr)
	}

	got := envAsMap(env)
	if got["K1"] != "override" {
		t.Fatalf("expected K1=override (later env wins), got %q", got["K1"])
	}
	if got["K2"] != "v2" {
		t.Fatalf("expected K2=v2 from earlier env, got %q", got["K2"])
	}
	if got["Path"] != `C:\jdk\bin` {
		t.Fatalf("expected Path collected from paths, got %q", got["Path"])
	}
}

// TestResolveEnvironmentVarsLaterWins checks that an environment listed later
// overrides an earlier one, and its paths are appended after existing ones.
func TestResolveEnvironmentVarsLaterWins(t *testing.T) {
	service := newTestService(t)

	_ = service.environmentRepo.Upsert(storage.Environment{ID: "C", Name: "C", Env: []storage.EnvVar{{Key: "K", Value: "first"}}, Paths: []string{`C:\a`}})
	_ = service.environmentRepo.Upsert(storage.Environment{ID: "P", Name: "P", Env: []storage.EnvVar{{Key: "K", Value: "second"}}, Paths: []string{`C:\b`, `C:\a`}})

	env, apiErr := service.resolveEnvironmentVars([]string{"C", "P"})
	if apiErr != nil {
		t.Fatalf("unexpected error: %v", apiErr)
	}

	got := envAsMap(env)
	if got["K"] != "second" {
		t.Fatalf("expected K=second (later env wins), got %q", got["K"])
	}
	if got["Path"] != `C:\a;C:\b` {
		t.Fatalf("expected Path=C:\\a;C:\\b (first occurrence kept), got %q", got["Path"])
	}
}

func TestResolveEnvironmentVarsMissing(t *testing.T) {
	service := newTestService(t)

	_, apiErr := service.resolveEnvironmentVars([]string{"does-not-exist"})
	if apiErr == nil || apiErr.Code != 404 {
		t.Fatalf("expected 404 for missing environment, got %+v", apiErr)
	}
}
