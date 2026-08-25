package services

import (
	"path/filepath"
	"strings"
	"testing"

	"github/TheSilentNights/VeloScriptsManager/service/storage"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	db, err := storage.OpenOrCreate(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewService(storage.CreateScriptRepo(db), storage.CreateEnvironmentRepo(db))
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

	// B is a base env; A inherits from B; C inherits from A and overrides K1.
	_ = service.environmentRepo.Upsert(storage.Environment{ID: "B", Name: "B", Env: []storage.EnvVar{{Key: "K2", Value: "v2"}}})
	_ = service.environmentRepo.Upsert(storage.Environment{ID: "A", Name: "A", Env: []storage.EnvVar{{Key: "K1", Value: "v1"}}, Children: []string{"B"}})
	_ = service.environmentRepo.Upsert(storage.Environment{ID: "C", Name: "C", Env: []storage.EnvVar{{Key: "K1", Value: "override"}}, Children: []string{"A"}})

	env, apiErr := service.resolveEnvironmentVars([]string{"C"})
	if apiErr != nil {
		t.Fatalf("unexpected error: %v", apiErr)
	}

	got := envAsMap(env)
	if got["K1"] != "override" {
		t.Fatalf("expected K1=override (own value overrides child), got %q", got["K1"])
	}
	if got["K2"] != "v2" {
		t.Fatalf("expected K2=v2 from inherited base env, got %q", got["K2"])
	}
}

// TestResolveEnvironmentVarsLaterWins checks that a top-level environment listed
// later overrides an earlier one, even when it is the child of a preceding env.
func TestResolveEnvironmentVarsLaterWins(t *testing.T) {
	service := newTestService(t)

	_ = service.environmentRepo.Upsert(storage.Environment{ID: "C", Name: "C", Env: []storage.EnvVar{{Key: "K", Value: "child"}}})
	_ = service.environmentRepo.Upsert(storage.Environment{ID: "P", Name: "P", Env: []storage.EnvVar{{Key: "K", Value: "parent"}}, Children: []string{"C"}})

	env, apiErr := service.resolveEnvironmentVars([]string{"P", "C"})
	if apiErr != nil {
		t.Fatalf("unexpected error: %v", apiErr)
	}

	if got := envAsMap(env); got["K"] != "child" {
		t.Fatalf("expected K=child (later-listed child overrides parent), got %q", got["K"])
	}
}

func TestResolveEnvironmentVarsCycle(t *testing.T) {
	service := newTestService(t)

	_ = service.environmentRepo.Upsert(storage.Environment{ID: "X", Name: "X", Children: []string{"Y"}})
	_ = service.environmentRepo.Upsert(storage.Environment{ID: "Y", Name: "Y", Children: []string{"X"}})

	_, apiErr := service.resolveEnvironmentVars([]string{"X"})
	if apiErr == nil {
		t.Fatal("expected a cycle error, got nil")
	}
}

func TestResolveEnvironmentVarsMissing(t *testing.T) {
	service := newTestService(t)

	_, apiErr := service.resolveEnvironmentVars([]string{"does-not-exist"})
	if apiErr == nil || apiErr.Code != 404 {
		t.Fatalf("expected 404 for missing environment, got %+v", apiErr)
	}
}
