package registory

import (
	"fmt"
	"os"

	"github.com/sasano8/kvtool/internal/core/repository"
)

var Commands = make(repository.Repository[Command])

type Command func([]string) error

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
