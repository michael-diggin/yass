package interpreter

import (
	"context"
	"fmt"
	"io"

	"github.com/michael-diggin/yass"
)

// Interpreter is the object that handles the connection to the data store
// And processes the incoming commands from the user in the inputCommand line.
type Interpreter struct {
	client *yass.Client
}

// New returns a new Interpreter instance.
func New(client *yass.Client) *Interpreter {
	return &Interpreter{client: client}
}

// ProcessCommand will convert the input from the user and excute it.
func (i *Interpreter) ProcessCommand(inputCommand string, w io.Writer) (exit bool) {
	if inputCommand == "\\q" {
		fmt.Fprintf(w, "\nQuitting...")
		return true
	}
	if inputCommand == "\\halp" {
		fmt.Fprintf(w, "\n"+helpOutput+"\n")
		return
	}

	command, err := parseInputToCommand(inputCommand)
	if err != nil {
		fmt.Fprintf(w, "%v\n", err)
		return
	}
	switch command.request {
	case PutRequest:
		err := i.client.Put(context.Background(), command.key, command.value)
		if err != nil {
			fmt.Fprintf(w, "%v\n", err)
		}
		return

	case FetchRequest:
		val, err := i.client.Fetch(context.Background(), command.key)
		if err != nil {
			fmt.Fprintf(w, "%v\n", err)
		} else {
			fmt.Fprintf(w, "%s\n", val)
		}
		return
	}
	fmt.Fprintf(w, "unknown command %v\n", command)
	return
}
