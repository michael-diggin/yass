package interpreter

import (
	"errors"
	"fmt"
	"strings"
)

// Request is the type of request to make
type Request int

const (
	// UknownRequest is the default request type
	UknownRequest Request = iota
	// PutRequest is the put request type
	PutRequest
	// FetchRequest is the fetch request type
	FetchRequest
)

// Command is the struct that contains the type of request, along with the key and value
type Command struct {
	request Request
	key     string
	value   string
}

func parseInputToCommand(input string) (*Command, error) {
	c := &Command{}
	if input == "" {
		return nil, errors.New("no input command provided")
	}

	args := strings.Split(input, " ")
	switch reqType := strings.ToLower(args[0]); reqType {
	case "put":
		if len(args) != 3 {
			return nil, fmt.Errorf("incorrect number of inputs for request of type PUT")
		}
		c.request = PutRequest
		c.key = args[1]
		c.value = args[2]
		return c, nil

	case "fetch":
		if len(args) != 2 {
			return nil, fmt.Errorf("incorrect number of inputs for request of type fetch")
		}
		c.request = FetchRequest
		c.key = args[1]
		return c, nil
	}
	return nil, fmt.Errorf("unknown request type %s", args[0])
}
