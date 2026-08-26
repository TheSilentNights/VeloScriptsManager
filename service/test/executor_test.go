package test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github/TheSilentNights/VeloScriptsManager/service/executor"
)

// TestStartEmptyCommand checks that a command with no params is rejected.
func TestStartEmptyCommand(t *testing.T) {
	_, err := executor.Start(context.Background(), "", nil, ".", nil)
	if err == nil {
		t.Fatal("expected error for empty command")
	}
	if !strings.Contains(err.Error(), "empty command") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestStartProcess verifies the async Process: output is delivered to
// subscribers and the process reports a clean exit.
func TestStartProcess(t *testing.T) {
	process, err := executor.Start(context.Background(), "cmd.exe", []string{"/c", "echo hello"}, "", nil)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	var output bytes.Buffer
	done := make(chan struct{})
	go func() {
		defer close(done)
		for chunk := range process.Subscribe() {
			output.Write(chunk.Data)
		}
	}()

	<-process.Done()
	<-done

	if process.ExitCode() != 0 {
		t.Fatalf("expected exit code 0, got %d (err: %v)", process.ExitCode(), process.Err())
	}
	if !strings.Contains(output.String(), "hello") {
		t.Fatalf("expected output to contain 'hello', got: %q", output.String())
	}
}

// TestStartProcessEnv verifies environment variables set before execution are
// visible to the process. cmd.exe expands %VAR% from the process environment,
// so the value passed via env must be printed.
func TestStartProcessEnv(t *testing.T) {
	env := []string{"VSM_TEST_VAR=hello-env"}
	process, err := executor.Start(context.Background(), "cmd.exe", []string{"/c", "echo %VSM_TEST_VAR%"}, "", env)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	var output bytes.Buffer
	done := make(chan struct{})
	go func() {
		defer close(done)
		for chunk := range process.Subscribe() {
			output.Write(chunk.Data)
		}
	}()

	<-process.Done()
	<-done

	if process.ExitCode() != 0 {
		t.Fatalf("expected exit code 0, got %d (err: %v)", process.ExitCode(), process.Err())
	}
	if !strings.Contains(output.String(), "hello-env") {
		t.Fatalf("expected output to contain 'hello-env', got: %q", output.String())
	}
}

// TestStartProcessStdin verifies stdin forwarding: the process is a real .bat
// file that reads a line from stdin (via delayed expansion) and echoes it back.
func TestStartProcessStdin(t *testing.T) {
	dir := t.TempDir()
	batPath := dir + "\\stdin_echo.bat"
	content := "@echo off\r\nsetlocal enabledelayedexpansion\r\nset /p line=\r\necho got:!line!\r\n"
	if err := os.WriteFile(batPath, []byte(content), 0644); err != nil {
		t.Fatalf("write bat failed: %v", err)
	}

	process, err := executor.Start(context.Background(), "cmd.exe", []string{"/c", batPath}, dir, nil)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	var output bytes.Buffer
	done := make(chan struct{})
	go func() {
		defer close(done)
		for chunk := range process.Subscribe() {
			output.Write(chunk.Data)
		}
	}()

	if _, err := process.WriteStdin([]byte("payload\r\n")); err != nil {
		t.Fatalf("write stdin failed: %v", err)
	}
	_ = process.CloseStdin()

	<-process.Done()
	<-done

	if process.ExitCode() != 0 {
		t.Fatalf("expected exit code 0, got %d (err: %v)", process.ExitCode(), process.Err())
	}
	if !strings.Contains(output.String(), "got:payload") {
		t.Fatalf("expected output to contain 'got:payload', got: %q", output.String())
	}
}
