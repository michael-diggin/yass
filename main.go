package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func main() {
	err := run(os.Args[1:], "dev-start", os.Stdout, os.Stderr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Server successfully running on addr localhost:8080")
}

func run(args []string, command string, out io.Writer, outErr io.Writer) error {
	if len(args) == 0 || args[0] != "deploy" {
		return errors.New("Incorrect command: usage 'yass deploy'")
	}
	cmd := exec.Command("/usr/bin/make", command)
	cmd.Stdout = out
	cmd.Stderr = outErr
	return cmd.Run()
}
