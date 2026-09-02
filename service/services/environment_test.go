package services

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github/TheSilentNights/VeloScriptsManager/service/executor"
	"github/TheSilentNights/VeloScriptsManager/service/ierrors"
	"github/TheSilentNights/VeloScriptsManager/service/models"
	"github/TheSilentNights/VeloScriptsManager/service/storage"
)

func newEnvTestService(t *testing.T) (*ScriptService, *storage.EnvironmentRepo) {
	t.Helper()
	db, err := storage.OpenOrCreate(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	environmentRepo := storage.CreateEnvironmentRepo(db)
	scriptRepo := storage.CreateScriptRepo(db)
	environmentService := NewEnvironmentService(environmentRepo)
	scriptService := NewScriptService(scriptRepo, executor.NewExecutionManager(), environmentService)
	return scriptService, environmentRepo
}

func mustInsertEnvironment(t *testing.T, repo *storage.EnvironmentRepo, environment storage.Environment) {
	t.Helper()
	if _, err := repo.Insert(environment); err != nil {
		t.Fatalf("insert environment %q: %v", environment.ID, err)
	}
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
	service, repo := newEnvTestService(t)

	mustInsertEnvironment(t, repo, storage.Environment{
		ID:    "A",
		Name:  "A",
		Env:   []storage.EnvVar{{Key: "K1", Value: "v1"}, {Key: "K2", Value: "v2"}},
		Paths: []string{`C:\jdk\bin`},
	})
	mustInsertEnvironment(t, repo, storage.Environment{
		ID:   "B",
		Name: "B",
		Env:  []storage.EnvVar{{Key: "K1", Value: "override"}},
	})

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
	service, repo := newEnvTestService(t)

	mustInsertEnvironment(t, repo, storage.Environment{
		ID:    "C",
		Name:  "C",
		Env:   []storage.EnvVar{{Key: "K", Value: "first"}},
		Paths: []string{`C:\a`},
	})
	mustInsertEnvironment(t, repo, storage.Environment{
		ID:    "P",
		Name:  "P",
		Env:   []storage.EnvVar{{Key: "K", Value: "second"}},
		Paths: []string{`C:\b`, `C:\a`},
	})

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

func TestResolveEnvironmentVarsEmpty(t *testing.T) {
	service, _ := newEnvTestService(t)

	env, apiErr := service.resolveEnvironmentVars(nil)
	if apiErr != nil {
		t.Fatalf("unexpected error: %v", apiErr)
	}
	if len(env) != 0 {
		t.Fatalf("expected no vars for empty id list, got %v", env)
	}
}

func TestResolveEnvironmentVarsMissing(t *testing.T) {
	service, _ := newEnvTestService(t)

	_, apiErr := service.resolveEnvironmentVars([]string{"does-not-exist"})
	if !errors.Is(apiErr, ierrors.EnvironmentNotFound) {
		t.Fatalf("expected EnvironmentNotFound, got %v", apiErr)
	}
}

func TestEnvironmentServiceCrud(t *testing.T) {
	_, repo := newEnvTestService(t)
	service := NewEnvironmentService(repo)

	count, apiErr := service.AddEnvironment(&models.AddEnvironmentRequest{
		Name:  "JDK",
		Paths: []string{`C:\jdk\bin`},
		Env:   []storage.EnvVar{{Key: "JAVA_HOME", Value: `C:\jdk`}},
	})
	if apiErr != nil {
		t.Fatalf("add environment failed: %v", apiErr)
	}
	if count != int64(1) {
		t.Fatalf("expected 1 row affected, got %v", count)
	}

	list, apiErr := service.ListEnvironments()
	if apiErr != nil {
		t.Fatalf("list environments failed: %v", apiErr)
	}
	environments, ok := list.([]storage.Environment)
	if !ok || len(environments) != 1 {
		t.Fatalf("expected 1 environment, got %#v", list)
	}
	if environments[0].Name != "JDK" || len(environments[0].Paths) != 1 || len(environments[0].Env) != 1 {
		t.Fatalf("unexpected stored environment: %#v", environments[0])
	}

	id := environments[0].ID
	updated, apiErr := service.UpdateEnvironment(&models.UpdateEnvironmentRequest{
		Id:    id,
		Name:  "JDK2",
		Paths: []string{`C:\jdk2\bin`},
		Env:   []storage.EnvVar{{Key: "JAVA_HOME", Value: `C:\jdk2`}},
	})
	if apiErr != nil {
		t.Fatalf("update environment failed: %v", apiErr)
	}
	if updated != int64(1) {
		t.Fatalf("expected 1 row affected, got %v", updated)
	}

	deleted, apiErr := service.DeleteEnvironment(id)
	if apiErr != nil {
		t.Fatalf("delete environment failed: %v", apiErr)
	}
	if deleted != int64(1) {
		t.Fatalf("expected 1 row affected, got %v", deleted)
	}

	list, apiErr = service.ListEnvironments()
	if apiErr != nil {
		t.Fatalf("list environments failed: %v", apiErr)
	}
	if environments, ok := list.([]storage.Environment); !ok || len(environments) != 0 {
		t.Fatalf("expected 0 environments after delete, got %#v", list)
	}
}
