package registory

import (
	"errors"
	"fmt"
	"os"

	repository "github.com/sasano8/kvtool/internal/core/repositories"
)

var Commands = repository.New[Command]()

type Command func([]string) error

func ParseCli() (Command, []string, error) {
	if len(os.Args) < 2 {
		//usage()
		os.Exit(1)
	}
	cmd := os.Args[1]
	c, err := Commands.Get(cmd)
	if err == nil {
		return c, os.Args[2:], nil
	} else {
		msg := fmt.Sprintf("Not found command: %s\n", cmd)
		return nil, os.Args[2:], errors.New(msg)
	}
}

func (f Command) Run(args []string) error {
	return f(args)
}

func (f Command) RunAsCli(args []string) {
	err := f.Run(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
