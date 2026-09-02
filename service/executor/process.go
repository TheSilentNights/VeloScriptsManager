package executor

//// Start launches the command
//func Start(ctx context.Context, title string, command []string, workDir string, env []string) (*Process, error) {
//	if len(command) == 0 {
//		return nil, errors.New("empty command")
//	}
//
//	if len(title) == 0 {
//		title = "Script"
//	}
//
//	//修复将标题识别为可执行程序的问题
//	title = strings.ReplaceAll(title, `"`, `'`)
//	if !strings.ContainsAny(title, " \t") {
//		title += " "
//	}
//
//	launchArgs := []string{"/c", "start", title, "/wait"}
//	launchArgs = append(launchArgs, command...)
//	cmd := exec.CommandContext(ctx, "cmd.exe", launchArgs...)
//	if workDir != "" {
//		cmd.Dir = workDir
//	}
//	if len(env) > 0 {
//		cmd.Env = mergeEnviron(os.Environ(), env)
//	}
//
//	if err := cmd.Start(); err != nil {
//		return nil, err
//	}
//
//	p := &Process{
//		cmd:      cmd,
//		done:     make(chan struct{}),
//		exitCode: -1,
//	}
//
//	go p.wait()
//
//	return p, nil
//}

// Kill force-terminates the process.

// Done is closed once the process has exited and all output has been flushed.

//// ExitCode returns the process exit code, or -1 while it is still running.
//func (p *Process) ExitCode() int {
//	p.mu.Lock()
//	defer p.mu.Unlock()
//	return p.exitCode
//}

// Err returns the error reported by cmd.Wait, or nil on a clean exit.
//func (p *Process) Err() error {
//	p.mu.Lock()
//	defer p.mu.Unlock()
//	return p.exitErr
//}

//type envEntry struct {
//	key   string
//	value string
//}

//func mergeEnviron(base, overrides []string) []string {
//	entries := linkedhashmap.New()
//
//	put := func(key, value string) {
//		lower := strings.ToLower(key)
//		if existing, found := entries.Get(lower); found {
//			key = existing.(envEntry).key
//		}
//		entries.Put(lower, envEntry{key: key, value: value})
//	}
//
//	for _, kv := range base {
//		key, value, _ := strings.Cut(kv, "=")
//		put(key, value)
//	}
//
//	for _, kv := range overrides {
//		key, value, _ := strings.Cut(kv, "=")
//		lower := strings.ToLower(key)
//		if lower == "path" {
//			if existing, found := entries.Get(lower); found {
//				entry := existing.(envEntry)
//				switch {
//				case entry.value == "":
//					entry.value = value
//				case value == "":
//				default:
//					entry.value = value + ";" + entry.value
//				}
//				entries.Put(lower, entry)
//				continue
//			}
//		}
//		put(key, value)
//	}
//
//	keys := entries.Keys()
//	out := make([]string, 0, len(keys))
//	for _, k := range keys {
//		raw, _ := entries.Get(k)
//		entry := raw.(envEntry)
//		out = append(out, entry.key+"="+entry.value)
//	}
//	return out
//}
