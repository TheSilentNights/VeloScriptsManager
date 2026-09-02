package ierrors

import "errors"

var GetEnvironmentDbError = errors.New("get environment fail")
var UpdateScriptDbError = errors.New("update script fail")
var DeleteScriptDbError = errors.New("delete script fail")
var AddEnvironmentDbError = errors.New("add environment fail")
var UpdateEnvironmentDbError = errors.New("update environment fail")
var DeleteEnvironmentDbError = errors.New("delete environment fail")
var ListEnvironmentsError = errors.New("list environments fail")
var GetScriptDbError = errors.New("get script fail")

var ScriptNotFound = errors.New("script not found")
var EnvironmentNotFound = errors.New("environment not found")
var ExecutionNotFound = errors.New("execution not found")

var ExecuteScriptError = errors.New("execute script error")

var ScriptIsRunningError = errors.New("script is running")
