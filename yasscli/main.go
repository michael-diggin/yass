package main

/*
import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/michael-diggin/yass"
	"github.com/michael-diggin/yass/yasscli/interpreter"
)

func main() {
	nodeAddress := flag.String("l", "localhost:8080", "location of a yass server node")
	flag.Parse()

	ctx := context.Background()
	client, err := yass.NewClient(ctx, *nodeAddress)
	if err != nil {
		fmt.Printf("Could not connect to yass server: %v", err)
		os.Exit(1)
	}
	defer client.Close()
	interpreter := interpreter.New(client)

	fmt.Println("Welcome to the YassCLI interpreter")
	fmt.Println("Enter \\halp for help, \\q to quit")

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\nyasscli=> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Encountered error when reading command: %v", err)
		}
		command := strings.TrimSpace(text)
		if command == "" {
			continue
		}

		exit := interpreter.ProcessCommand(command, os.Stdout)
		if exit {
			break
		}
	}
}

*/
