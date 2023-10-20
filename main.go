package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rafaelespinoza/ged/internal/cmd"
)

const (
	ansiOpen   = "\033["
	ansiClose  = "m"
	ansiBoldOn = ansiOpen + "1" + ansiClose
	ansiReset  = ansiOpen + "0" + ansiClose
)

func main() {
	err := cmd.New().Run(context.Background(), os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%v%s\n", ansiBoldOn, err, ansiReset)
		os.Exit(1)
	}
}
