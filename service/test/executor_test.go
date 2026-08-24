package test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github/TheSilentNights/VeloScriptsManager/service/executor"
)

func TestExecEmptyCommand(t *testing.T) {
	_, err := executor.Exec(context.Background(), "", nil, ".")
	if err == nil {
		t.Fatal("expected error for empty command")
	}
	if !strings.Contains(err.Error(), "empty command") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecCommand(t *testing.T) {
	_, err := executor.Exec(context.Background(), "cmd.exe", []string{"/c", "C:\\develop\\test_env\\neoforge26.1.2\\run.bat"}, "C:\\develop\\test_env\\neoforge26.1.2")
	if err != nil {
		println(err.Error())
	}
}

// TestStartProcess verifies the async Process: output is delivered to
// subscribers and the process reports a clean exit.
func TestStartProcess(t *testing.T) {
	process, err := executor.Start(context.Background(), "cmd.exe", []string{"/c", "echo hello"}, "")
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	var output bytes.Buffer
	done := make(chan struct{})
	go func() {
		defer close(done)
		for chunk := range process.Subscribe() {
			output.Write(chunk)
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

// TestStartProcessStdin verifies stdin forwarding: the process is a real .bat
// file that reads a line from stdin (via delayed expansion) and echoes it back.
func TestStartProcessStdin(t *testing.T) {
	dir := t.TempDir()
	batPath := dir + "\\stdin_echo.bat"
	content := "@echo off\r\nsetlocal enabledelayedexpansion\r\nset /p line=\r\necho got:!line!\r\n"
	if err := os.WriteFile(batPath, []byte(content), 0644); err != nil {
		t.Fatalf("write bat failed: %v", err)
	}

	process, err := executor.Start(context.Background(), "cmd.exe", []string{"/c", batPath}, dir)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	var output bytes.Buffer
	done := make(chan struct{})
	go func() {
		defer close(done)
		for chunk := range process.Subscribe() {
			output.Write(chunk)
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

// TestExecError keeps the error-path assertion available.
func TestExecError(t *testing.T) {
	_, err := executor.Exec(context.Background(), "cmd.exe", []string{"/c", "exit 3"}, "")
	if err == nil {
		t.Fatal("expected non-nil error for failing command")
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *exec.ExitError, got %T", err)
	}
}
