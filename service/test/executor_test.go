package test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github/TheSilentNights/VeloScriptsManager/service/executor"
)

// TestStartEmptyCommand checks that a command with no params is rejected.
func TestStartEmptyCommand(t *testing.T) {
	_, err := executor.Start(
		context.Background(),
		"",
		nil,
		".",
		nil,
	)
	if err == nil {
		t.Fatal("expected error for empty command")
	}
	if !strings.Contains(err.Error(), "empty command") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestStartProcessExitCode verifies that the exit code of the command running
// in the opened terminal window is reported once it is done.
func TestStartProcessExitCode(t *testing.T) {
	process, err := executor.Start(
		context.Background(),
		"exit-code-test",
		[]string{"cmd.exe", "/c", "exit", "7"},
		"",
		nil,
	)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	<-process.Done()

	if process.ExitCode() != 7 {
		t.Fatalf("expected exit code 7, got %d (err: %v)", process.ExitCode(), process.Err())
	}
}

// TestStartProcessEnv verifies environment variables set before execution are
// visible to the process. cmd.exe expands %VAR% from the process environment,
// so the value passed via env must end up in the redirected output file.
func TestStartProcessEnv(t *testing.T) {
	dir := t.TempDir()
	batPath := dir + "\\env.bat"
	content := "@echo off\r\nset VSM_TEST_VAR>env.txt\r\n"
	if err := os.WriteFile(batPath, []byte(content), 0644); err != nil {
		t.Fatalf("write bat failed: %v", err)
	}
	env := []string{"VSM_TEST_VAR=hello-env"}

	process, err := executor.Start(
		context.Background(),
		"env-test",
		[]string{"cmd.exe", "/c", batPath},
		dir,
		env,
	)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	<-process.Done()

	if process.ExitCode() != 0 {
		t.Fatalf("expected exit code 0, got %d (err: %v)", process.ExitCode(), process.Err())
	}

	data, err := os.ReadFile(dir + "\\env.txt")
	if err != nil {
		t.Fatalf("read output failed: %v", err)
	}
	if !strings.Contains(string(data), "hello-env") {
		t.Fatalf("expected output to contain 'hello-env', got: %q", string(data))
	}
}
