package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rafaelespinoza/ged/internal/entity"
	"github.com/rafaelespinoza/ged/internal/log"
)

var errNoFZF = errors.New("fzf binary not available")

type fzfChooser struct {
	bin        string
	inputLines []string
}

const fzfLineFieldSeparator = "\t"

func newFZFChooser(ctx context.Context, people []*entity.Person) (personIDChooser, error) {
	fzf := os.Getenv("FZF_BIN")
	if len(fzf) == 0 {
		fzf = "fzf"
	}
	pathToBin, err := exec.LookPath(fzf)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errNoFZF, err)
	}

	return &fzfChooser{bin: pathToBin, inputLines: makeFZFPeople(fzfLineFieldSeparator, people)}, nil
}

func (c *fzfChooser) choosePersonID(ctx context.Context) (out string, err error) {
	// this part was adapted from https://github.com/junegunn/fzf/wiki/Language-bindings#go.
	const fzfOptions = "--header 'pick person to compare (try fuzzy search)' --layout reverse --height 66% --info=inline --border=bold --margin=2 --padding=2 --cycle --with-nth='2..'"
	shell := os.Getenv("SHELL")
	if len(shell) == 0 {
		shell = "sh"
	}

	cmd := exec.CommandContext(ctx, shell, "-c", c.bin+" "+fzfOptions)
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Error(ctx, nil, err, "stdin error")
	}
	go func() {
		for _, line := range c.inputLines {
			fmt.Fprintln(stdin, line)
		}
		if cerr := stdin.Close(); cerr != nil {
			log.Error(ctx, nil, cerr, "could not close input pipe")
		}
	}()

	var userChoice string
	if result, rerr := cmd.Output(); rerr != nil {
		err = rerr
		return
	} else {
		before, after, found := strings.Cut(string(result), "\n")
		log.Info(ctx, map[string]any{"before": before, "after": after, "found": found}, "")
		userChoice = before
	}

	// parsing the user choice depends highly on the formatting done to the
	// input line.
	fields := strings.Fields(userChoice)
	if len(fields) < 1 {
		err = errors.New("it seems like an early exit")
		return
	}

	out = fields[0]
	return
}

func makeFZFPeople(fieldDelimiter string, people []*entity.Person) (out []string) {
	out = make([]string, len(people))
	var line []string

	for i, person := range people {
		// Set cap to maximum number of non-empty fields you might have (assuming 2 parents).
		line = make([]string, 0, 6)

		line = append(line, person.ID)
		line = append(line, "Name:"+person.Name.Full())
		if d := formatDate(person.Birthdate); d != nil {
			line = append(line, "Birth:"+*d)
		}
		if d := formatDate(person.Deathdate); d != nil {
			line = append(line, "Death:"+*d)
		}
		for _, parent := range person.Parents {
			line = append(line, "Parent:"+parent.Name.Full())
		}

		out[i] = strings.Join(line, fieldDelimiter)
	}

	return
}
