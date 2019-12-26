package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os/exec"
)

var pass *string

type Command struct {
	Pass string
	Cmd  string
	Opt  string
	Args string
}

func handler(w http.ResponseWriter, r *http.Request) {
	var req Command
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if req.Pass != *pass {
		http.Error(w, "pass err", 400)
		return
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd := exec.Command(req.Cmd, req.Opt, req.Args)
	// cmd := exec.Command("sh", "-c", "echo stdout; echo 1>&2 stderr")

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	var errStdout, errStderr error
	stdout := io.MultiWriter(w, &stdoutBuf)
	stderr := io.MultiWriter(w, &stderrBuf)
	err := cmd.Start()
	if err != nil {
		log.Printf("cmd.Start() failed with '%s'\n", err)
		//r.Write()
		return
	}

	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
	}()

	go func() {
		_, errStderr = io.Copy(stderr, stderrIn)
	}()

	err = cmd.Wait()
	if err != nil {
		log.Printf("cmd.Run() failed with %s\n", err)
		return
	}
	if errStdout != nil || errStderr != nil {
		log.Printf("failed to capture stdout or stderr\n")
		return
	}
	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	log.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
}

func main() {
	pass = flag.String("pass", "mypass", "passwd")
	addr := flag.String("addr", ":9090", "bind addr and port")
	crt := flag.String("crt", "", "cert file")
	key := flag.String("key", "", "key file")
	flag.Parse()

	http.HandleFunc("/", handler)

	if *crt == "" || *key == "" {
		log.Printf("start http server\n")
		log.Fatal(http.ListenAndServe(*addr, nil))
	} else {
		log.Printf("start https server\n")
		log.Fatal(http.ListenAndServeTLS(*addr, *crt, *key, nil))
	}
}

// curl -X POST -d '{"pass": "mypass", "cmd": "bash", "opt": "-c", "args": "ls -l ~; echo hello"}' http://localhost:9090
