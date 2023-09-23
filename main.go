package main

import (
	"fmt"
	"github.com/mehrankamal/monkey/repl"
	"os"
	"os/user"
)

func main() {
	current, err := user.Current()

	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello %s! This is the Monkey programming language!\n", current.Username)
	fmt.Printf("Feel free to type in Monkey commands\n")
	repl.Start(os.Stdin, os.Stdout)
}
