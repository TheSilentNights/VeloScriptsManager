package executor

import (
	"context"
	"errors"
	"os/exec"
	"strings"
)

// Exec 注意，这里的exec会自动根据runner判断从哪里截取可执行。如果没有预设的runner会从params里找到第一个参数作为runner来执行
func Exec(ctx context.Context, runner string, params []string, workDir string) ([]byte, error) {
	if len(params) == 0 {
		return nil, errors.New("empty command")
	}

	var exe string
	var args []string

	if runner != "" {
		exe = runner
		args = params
	} else {
		exe = params[0]
		args = params[1:]
	}

	cmd := exec.CommandContext(ctx, exe, args...)

	if workDir != "" {
		cmd.Dir = workDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// err is a *exec.ExitError when the process exited non-zero.
		return output, err
	}
	return output, nil
}

// SplitCommand splits a command line into an executable and its arguments,
// honouring double quotes so that paths and arguments containing spaces remain
// a single token.
func SplitCommand(command string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	inQuotes := false

	for _, r := range command {
		switch {
		case r == '"':
			inQuotes = !inQuotes
		case (r == ' ' || r == '\t') && !inQuotes:
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if inQuotes {
		return nil, errors.New("unbalanced quotes in command")
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens, nil
}
