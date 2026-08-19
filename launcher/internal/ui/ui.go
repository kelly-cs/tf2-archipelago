// Package ui is the launcher's interactive prompt layer. Each prompt takes the
// saved-config value as the default, so a returning operator hits Enter to keep
// what they had. Nothing here reads or writes the config file; the caller does.
package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Prompt is the interactive surface, backed by stdin/stdout. Methods return the
// value the operator entered or, on an empty line, the default.
type Prompt struct {
	reader *bufio.Reader

	// closed is set once stdin has no more to give: a pipe that ended, a
	// service with no terminal, a redirect from /dev/null. Every prompt after
	// that takes its default without asking.
	closed bool
}

// New returns a Prompt reading from stdin.
func New() *Prompt {
	return &Prompt{reader: bufio.NewReader(os.Stdin)}
}

// Closed reports whether stdin ended. The caller checks it before insisting on
// a value nobody can type.
func (p *Prompt) Closed() bool { return p.closed }

/*
line reads one answer, and says whether anybody was there to give it.

Every prompt used to drop the error from ReadString and loop on a blank
answer. Started without a terminal, that is not a loop that ends: the read
returns EOF and an empty string as fast as the process can ask, and the one
that wanted a room address wrote 849 MB of "the address is empty" in ninety
seconds before anybody noticed.

So the error is the loop's bound. Once stdin is done it stays done, and a
prompt that cannot be answered takes its default rather than asking again.
*/
func (p *Prompt) line() (string, bool) {
	if p.closed {
		return "", false
	}
	text, err := p.reader.ReadString('\n')
	if err != nil && text == "" {
		p.closed = true
		fmt.Println()
		return "", false
	}
	return strings.TrimSpace(text), true
}

// Text asks for a string. The default is shown in brackets and used on Enter.
func (p *Prompt) Text(label, def string) string {
	if def != "" {
		fmt.Printf("%s [%s]: ", label, def)
	} else {
		fmt.Printf("%s: ", label)
	}
	line, ok := p.line()
	if !ok || line == "" {
		return def
	}
	return line
}

// Password asks for a string without echoing it. The default is never shown;
// an empty answer keeps it. On Windows the no-echo is handled by termReadLine.
func (p *Prompt) Password(label, def string) string {
	fmt.Printf("%s%s: ", label, maskedDefault(def))
	line := p.readMaskedLine()
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}

// Int asks for an integer. The default is shown and used on Enter. A non-numeric
// answer re-prompts.
func (p *Prompt) Int(label string, def int) int {
	for {
		fmt.Printf("%s [%d]: ", label, def)
		line, ok := p.line()
		if !ok || line == "" {
			return def
		}
		n, err := strconv.Atoi(line)
		if err != nil {
			fmt.Println("  enter a whole number, or press Enter for the default")
			continue
		}
		return n
	}
}

// Bool asks for yes/no. The default is shown and used on Enter.
func (p *Prompt) Bool(label string, def bool) bool {
	d := "n"
	if def {
		d = "y"
	}
	for {
		fmt.Printf("%s [%s]: ", label, d)
		answer, ok := p.line()
		if !ok {
			return def
		}
		switch strings.ToLower(answer) {
		case "":
			return def
		case "y", "yes":
			return true
		case "n", "no":
			return false
		}
		fmt.Println("  enter y or n, or press Enter for the default")
	}
}

// Choice asks for one of a fixed set of options. The default is shown and used
// on Enter; an unknown answer re-prompts.
func (p *Prompt) Choice(label string, options []string, def string) string {
	for {
		fmt.Printf("%s %v [%s]: ", label, options, def)
		line, ok := p.line()
		if !ok || line == "" {
			return def
		}
		for _, opt := range options {
			if strings.EqualFold(line, opt) {
				return opt
			}
		}
		fmt.Printf("  pick one of %v, or press Enter for %s\n", options, def)
	}
}

func maskedDefault(def string) string {
	if def == "" {
		return ""
	}
	return " (set, press Enter to keep)"
}

func (p *Prompt) readMaskedLine() string {
	if line, err := termReadLine(p.reader); err == nil {
		return line
	}
	// No terminal to turn the echo off on, so this is the ordinary read, and
	// it has to notice the end of stdin like every other one.
	line, _ := p.line()
	return line
}

// Option is one row of a Select: the value that gets saved, and the line the
// player reads.
type Option struct {
	Value string
	Label string
}

// Select asks for one of a numbered list. It takes the number, the value, or
// Enter for the default, which is marked in the list. A list beats free text
// wherever the set is known and short: nobody has to remember how ghost_town
// is spelled.
func (p *Prompt) Select(label string, options []Option, def string) string {
	fmt.Printf("%s:\n", label)
	for i, option := range options {
		marker := " "
		if option.Value == def {
			marker = "*"
		}
		fmt.Printf("  %s %2d. %s\n", marker, i+1, option.Label)
	}
	for {
		fmt.Printf("  Number or name [%s]: ", def)
		line, ok := p.line()
		if !ok || line == "" {
			return def
		}
		if n, err := strconv.Atoi(line); err == nil {
			if n >= 1 && n <= len(options) {
				return options[n-1].Value
			}
			fmt.Printf("  pick a number between 1 and %d\n", len(options))
			continue
		}
		for _, option := range options {
			if strings.EqualFold(line, option.Value) {
				return option.Value
			}
		}
		fmt.Println("  pick a number, or type one of the names")
	}
}

// IntRange asks for a number inside a range, and says what the range is. An
// answer outside it re-prompts rather than being clamped, because a silently
// corrected number is one the player never sees.
func (p *Prompt) IntRange(label string, def, low, high int) int {
	if def < low {
		def = low
	}
	if def > high {
		def = high
	}
	for {
		fmt.Printf("%s (%d to %d) [%d]: ", label, low, high, def)
		line, ok := p.line()
		if !ok || line == "" {
			return def
		}
		n, err := strconv.Atoi(line)
		if err != nil {
			fmt.Println("  enter a whole number, or press Enter for the default")
			continue
		}
		if n < low || n > high {
			fmt.Printf("  pick a number between %d and %d\n", low, high)
			continue
		}
		return n
	}
}
