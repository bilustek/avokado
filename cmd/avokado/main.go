package main

import (
	"fmt"
	"io"
	"os"

	"github.com/bilustek/avokado"
	"github.com/bilustek/avokado/avokadoerror"
)

const usage = `avokado - Django inspired Go Web REST API framework

Usage: avokado [command|-flags]

  help,    -h, --help        show this help
  version, -v, --version     show version information
  create,  -c, --create      create shiny new rest api server!

Examples:
  
  avokado create github.com/bilustek/splitray
  avokado create github.com/vigo/weatherapi

`

func main() {
	if err := run(os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(w io.Writer, args []string) error {
	if len(args) == 0 {
		fmt.Fprint(w, usage)

		return nil
	}

	switch args[0] {
	case "version", "--version", "-v":
		fmt.Fprintln(w, avokado.Version)

		return nil
	case "help", "--help", "-h":
		fmt.Fprint(w, usage)

		return nil
	default:
		return avokadoerror.New("unknown command")
	}
}
