package test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github/TheSilentNights/VeloScriptsManager/service/executor"
)

func TestSplitCommand(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{"empty", "", nil, false},
		{"single", "echo", []string{"echo"}, false},
		{"multiple", "git status --short", []string{"git", "status", "--short"}, false},
		{"quoted space", `"C:\Program Files\tool.exe" run`, []string{`C:\Program Files\tool.exe`, "run"}, false},
		{"extra spaces", "  git   log  ", []string{"git", "log"}, false},
		{"unbalanced quote", `"foo bar`, nil, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := executor.SplitCommand(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestExecEmptyCommand(t *testing.T) {
	_, err := executor.Exec(context.Background(), "   ", "", ".")
	if err == nil {
		t.Fatal("expected error for empty command")
	}
	if !errors.Is(err, errors.New("empty command")) {
		// SplitCommand itself returns the empty-command error only when 0 tokens.
		if !strings.Contains(err.Error(), "empty command") {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestExecCommand(t *testing.T) {
	_, err := executor.Exec(context.Background(), "/c C:\\develop\\test_env\\neoforge26.1.2\\run.bat", "cmd.exe", "C:\\develop\\test_env\\neoforge26.1.2")
	if err != nil {
		println(err.Error())
	}
}
