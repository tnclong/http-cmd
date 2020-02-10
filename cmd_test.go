package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var server *httptest.Server

func TestMain(m *testing.M) {
	http.HandleFunc(pattern, Command)
	server = httptest.NewServer(http.DefaultServeMux)
	ret := m.Run()
	server.Close()
	os.Exit(ret)
}

const pattern = "/danger/cmd"

func ExampleHello() {
	setenv("echo, cat")
	request(server, "timeout=10&name=echo&arg=hello&arg=http-cmd")
	// Output:
	// Get /danger/cmd?timeout=10&name=echo&arg=hello&arg=http-cmd
	//
	// Response:
	//
	// name: echo
	// arg: [hello http-cmd]
	// timeout: 10
	//
	// stdout:
	// hello http-cmd
	//
	// stderr:
}

func ExampleAllowedAllCmds() {
	setenv("***")
	request(server, "timeout=10&name=ls&arg=cmd.go")
	// Output:
	// Get /danger/cmd?timeout=10&name=ls&arg=cmd.go
	//
	// Response:
	//
	// name: ls
	// arg: [cmd.go]
	// timeout: 10
	//
	// stdout:
	// cmd.go
	//
	// stderr:
}

func ExampleTimeout() {
	setenv("***")
	request(server, "timeout=1&name=sleep&arg=10")
	// Output:
	// Get /danger/cmd?timeout=1&name=sleep&arg=10
	//
	// Response:
	//
	// name: sleep
	// arg: [10]
	// timeout: 1
	//
	// stdout:
	//
	// stderr:
	//
	// err:
	// signal: killed
}

func ExampleRequiredTimeout() {
	setenv("***")
	request(server, "name=ls&arg=cmd.go")
	// Output:
	// Get /danger/cmd?name=ls&arg=cmd.go
	//
	// Response:
	//
	// parse timeout with err: strconv.ParseInt: parsing "": invalid syntax
}

func ExampleUnallowedCmds() {
	setenv("ls")
	request(server, "timeout=10&name=ps")
	// Output:
	// Get /danger/cmd?timeout=10&name=ps
	//
	// Response:
	//
	// name="ps" is unallowed while check env DANGER_HTTP_ALLOWED_CMDS=ls
}

func setenv(value string) {
	err := os.Setenv("DANGER_HTTP_ALLOWED_CMDS", value)
	if err != nil {
		panic(err)
	}
}

func request(server *httptest.Server, query string) {
	client := server.Client()
	path := pattern + "?" + query
	res, err := client.Get(server.URL + path)
	if err != nil {
		panic(fmt.Sprintf("Get %s with err: %v", path, err))
	}
	defer res.Body.Close()

	fmt.Println("Get", path)

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(fmt.Sprintf("Get %s read response with err: %v", path, err))
	}

	fmt.Println()
	fmt.Println("Response:")
	fmt.Println()

	data = bytes.ReplaceAll(data, []byte("\n\n"), []byte("\n"))
	fmt.Println(string(data))
}
