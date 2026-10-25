# DeepStack
 
## Overview

DeepStack is a structured logging and error management library for the Quollix ecosystem based on the Go SDK library [log/slog](https://go.dev/blog/slog).

### Error Design

Here is the DeepStack error data structure:

```go
type DeepStackError struct {
    Message    string
    StackTrace string
    Context    map[string]any
}
```

* **Log Or Return Principle**: To avoid duplication, either log or return an error, but not both. Therefore, errors are created in a low-level function and passed up to a higher-level function, where they are logged once. However, when an error is logged in a high-level function, it must contain information about which function caused it. Therefore, a **stack trace is included in the DeepStack error** upon creation.
* **DeepStack error implements Go `error` interface**, allowing it to be used seamlessly with Go's error handling mechanisms and reducing its coupling with the code in which it is used.
* **DeepStackError Structure**: Other logging libraries often encode context and stack trace information in a single error string, adding encoding complexity. By contrast, DeepStack errors are rich data structures that contain additional fields for context and stack traces. This makes them easy to understand and avoids unnecessary complexity.
* **Adding Error Context**: DeepStack error data structures have a context field that can store key-value pairs. These pairs can be added to extend the context during DeepStack error creation or by intermediate functions passing up the DeepStack error. As these operations are performed directly on the error data structure, the process is much lighter than the costly encoding operations performed by other logging libraries.

### Logging Design

* **Structured Logging** is the general use case of the DeepStack library that allows for easy filtering and searching of logs.
* **Error Logging** is a special case in which the logger recognizes `*DeepStackError` values and includes their context and stack trace in the log record. The handlers passed to `NewDeepStackLogger` determine the output format and destination.

## Usage Overview

### Basic Structured Logging

```go
package main

import (
    "log/slog"

    "github.com/quollix/deepstack"
)

func main() {
    logger := deepstack.NewDeepStackLogger(deepstack.NewRawConsoleHandler(slog.LevelInfo))
    logger.Info("user logged in", "name", "john", "age", 23)
}
```

Output:

```text
2025-08-31 16:58:52.135 INFO main.go:10 "user logged in" age=23 name=john
```

### Error Management

Use a simple **create → propagate → handle** lifecycle with DeepStack. Stack traces are captured **once** at creation. Logging happens **once** at a boundary, e.g. an HTTP handler.

#### 1) Error creation

Create a DeepStack error at the first failure point so the stack trace points to the origin.
- New failure: `logger.NewError("resource not found")`
- From a Go error: `logger.NewError(err.Error())`

#### 2) Error propagation

Return the error upward. Intermediate layers **do not log**, but they may enrich the error with additional context via `AddContext()` and return it.

#### 3) Error handling

Log the error at the boundary by passing it directly to a log method, for example `logger.Error(err)`. A `*DeepStackError` is logged with its context and stack trace. Other Go errors are logged using their `Error()` text. Use `logger.NewError(err.Error())` when you need to capture a DeepStack stack trace at the wrapping point.

#### Context

At any stage you can add context to a DeepStack error as key–value pairs. This context travels with the error and is included when it’s finally logged.

### Full Logging Example

```go
package main

import (
    "log/slog"
    "os/exec"

    "github.com/quollix/deepstack"
)

var logger = deepstack.NewDeepStackLogger(deepstack.NewRawConsoleHandler(slog.LevelInfo))

const (
    // good practice to hard code field names for reusability
    AccessDeniedField = "access_denied"
)

func main() {
    err := func1()
    logger.Error(err)
}

// intermediate function passing up the error
func func1() error {
    err := func2()
    if err != nil {
        return err
    }
    // do some other stuff here
    return nil
}

// intermediate function adding context and then passing up the error
func func2() error {
    return logger.AddContext(func3(), AccessDeniedField, "access token not found")
}

var someCondition = true

// the function where the error occurs the first time
func func3() error {
    if someCondition {
        // create own error
        return logger.NewError("access token not found", AccessDeniedField, "no access token provided")
    } else {
        // wrap error from external library
        err := exec.Command("not-existing-command").Run()
        return logger.NewError(err.Error())
    }
}
```

The console handler prints the structured fields and, for DeepStack errors, the stack trace. Exact file paths and line numbers depend on where the example is run.

Example output:

```text
main.func3
    /path/to/project/main.go:42
main.func2
    /path/to/project/main.go:33
main.func1
    /path/to/project/main.go:23
main.main
    /path/to/project/main.go:17
runtime.main
    /path/to/go/src/runtime/proc.go:283
runtime.goexit
    /path/to/go/src/runtime/asm_amd64.s:1700

2026-05-15 12:00:00.000 ERROR main.go:18 "access token not found" access_denied="no access token provided"
main.func3
    /path/to/main.go:...
main.func2
    /path/to/main.go:...
```

### Asserting DeepStack Errors in Tests

Use the provided assertion helper to verify DeepStack errors in tests. It requires an exact match of the error message and the context.

```go
func TestStuff(t *testing.T) {
    err := someFunction()
    deepstack.AssertDeepStackError(t, err, "some error message", "name", "john", "age", "23")
}
```

### Register New Log Handlers

Pass one or more `log/slog` handlers to `NewDeepStackLogger()` to choose where logs go. The package provides `NewRawConsoleHandler` and `NewJsonConsoleHandler`, and both write to standard output. You can also pass handlers for files, databases, or other destinations by using your own `slog.Handler` implementations.

### Contributing

Please read the [Community](https://quollix.org/docs/project/community/) articles for more information on how to contribute to the project and interact with others.

### License

This project is licensed under the [MIT License](LICENSE).
