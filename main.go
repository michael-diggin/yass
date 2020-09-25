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
	fmt.Printf("Server successfully running on addr localhost:%s", os.Args[2])
}

func run(args []string, command string, out io.Writer, outErr io.Writer) error {
	if len(args) != 2 || args[0] != "deploy" {
		return errors.New("Incorrect command: usage 'yass deploy PORT'")
	}

	cmd := exec.Command("/usr/bin/make", command, "PORT="+args[1])
	cmd.Stdout = out
	cmd.Stderr = outErr
	return cmd.Run()
}
