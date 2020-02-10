package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Command is a http.HandlerFunc that exec commands by given query string.
//
// Usage:
//    https://github.com/tnclong/http-cmd/blob/master/cmd_test.go
func Command(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")

	q := r.URL.Query()
	name := q.Get("name")
	arg := q["arg"]
	timeout, err := strconv.ParseInt(q.Get("timeout"), 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("parse timeout with err: %s", err.Error()), http.StatusUnprocessableEntity)
		return
	}
	if timeout <= 0 {
		timeout = 10
	}

	env := os.Getenv("DANGER_HTTP_ALLOWED_CMDS")
	if !isAllowedCmds(env, name) && !isAllowedAllCmds(env) {
		http.Error(w, fmt.Sprintf("name=%q is unallowed while check env DANGER_HTTP_ALLOWED_CMDS=%s", name, env), http.StatusUnprocessableEntity)
		return
	}

	ctx, cf := context.WithTimeout(r.Context(), time.Duration(timeout)*time.Second)
	defer cf()
	cmd := exec.CommandContext(ctx, name, arg...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	w.WriteHeader(http.StatusOK)
	err = cmd.Run()
	fmt.Fprintf(w, "name: %s\n", name)
	fmt.Fprintf(w, "arg: %s\n", arg)
	fmt.Fprintf(w, "timeout: %d\n", timeout)
	fmt.Fprintln(w)
	fmt.Fprintln(w)

	w.Write([]byte("stdout:\n"))
	w.Write(stdout.Bytes())
	fmt.Fprintln(w)
	fmt.Fprintln(w)

	w.Write([]byte("stderr:\n"))
	w.Write(stderr.Bytes())
	fmt.Fprintln(w)
	fmt.Fprintln(w)

	if err != nil {
		w.Write([]byte("err:\n"))
		w.Write([]byte(err.Error()))
		fmt.Fprintln(w)
	}
}

func isAllowedAllCmds(env string) bool {
	return isAllowedCmds(env, "***")
}

func isAllowedCmds(env, name string) bool {
	i := strings.Index(env, name)
	return backOk(env, i) && forwardOk(env, i, name)
}

func backOk(env string, i int) bool {
	if i < 0 {
		return false
	}
	for b := i - 1; b >= 0; b-- {
		if env[b] == ' ' {
			continue
		}
		if env[b] == ',' {
			return true
		}
		return false
	}
	return true
}

func forwardOk(env string, i int, name string) bool {
	if i < 0 {
		return false
	}
	for f := i + len(name); f < len(env); f++ {
		if env[f] == ' ' {
			continue
		}
		if env[f] == ',' {
			return true
		}
		return false
	}
	return true
}
