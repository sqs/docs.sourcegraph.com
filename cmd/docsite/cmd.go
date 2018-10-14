package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
)

// command is a subcommand handler and its flag set.
type command struct {
	// flagSet is the flag set for the command.
	FlagSet *flag.FlagSet

	// ShortDescription is the short description for the command shown in the top-level help
	// message.
	ShortDescription string

	// LongDescription is the long description for the command shown in the command's help message.
	LongDescription string

	// aliases for the command.
	aliases []string

	// handler is the function that is invoked to handle this command.
	handler func(args []string) error

	// flagSet.Usage function to invoke on e.g. -h flag. If nil, a default one
	// one is used.
	usageFunc func()
}

func (c *command) NameAndAliases() string {
	v := make([]string, 1+len(c.aliases))
	v[0] = c.FlagSet.Name()
	copy(v[1:], c.aliases)
	return strings.Join(v, ",")
}

// matches tells if the given name matches this command or one of its aliases.
func (c *command) matches(name string) bool {
	if name == c.FlagSet.Name() {
		return true
	}
	for _, alias := range c.aliases {
		if name == alias {
			return true
		}
	}
	return false
}

// commander represents a top-level command with subcommands.
type commander []*command

// run runs the command.
func (c commander) run(flagSet *flag.FlagSet, cmdName string, usage *template.Template, args []string) {
	// Parse flags.
	flagSet.Usage = func() {
		data := struct {
			FlagUsage func() string
			Commands  []*command
		}{
			FlagUsage: func() string { flag.PrintDefaults(); return "" },
			Commands:  c,
		}
		if err := usage.Execute(flag.CommandLine.Output(), data); err != nil {
			log.Fatal(err)
		}
	}
	if !flagSet.Parsed() {
		flagSet.Parse(args)
	}

	// Print usage if the command is "help".
	if flagSet.Arg(0) == "help" || flagSet.NArg() == 0 {
		flagSet.Usage()
		os.Exit(0)
	}

	// Configure default usage funcs for commands.
	for _, cmd := range c {
		cmd.FlagSet.Usage = func() {
			fmt.Fprintln(flag.CommandLine.Output(), "Usage:")
			fmt.Fprintln(flag.CommandLine.Output())
			fmt.Fprintf(flag.CommandLine.Output(), "  %s [options] %s [command options]\n", cmdName, cmd.FlagSet.Name())
			if cmd.LongDescription != "" {
				fmt.Fprintln(flag.CommandLine.Output())
				fmt.Fprintln(flag.CommandLine.Output(), cmd.LongDescription)
				fmt.Fprintln(flag.CommandLine.Output())
			}
			fmt.Fprintln(flag.CommandLine.Output(), "The command options are:")
			fmt.Fprintln(flag.CommandLine.Output())
			cmd.FlagSet.PrintDefaults()
		}
	}

	// Find the subcommand to execute.
	name := flagSet.Arg(0)
	for _, cmd := range c {
		if !cmd.matches(name) {
			continue
		}

		// Parse subcommand flags.
		args := flagSet.Args()[1:]
		if err := cmd.FlagSet.Parse(args); err != nil {
			panic(fmt.Sprintf("all registered commands should use flag.ExitOnError: error: %s", err))
		}

		// Execute the subcommand.
		if err := cmd.handler(flagSet.Args()[1:]); err != nil {
			if _, ok := err.(*usageError); ok {
				log.Println(err)
				cmd.FlagSet.Usage()
				os.Exit(2)
			}
			if e, ok := err.(*exitCodeError); ok {
				if e.error != nil {
					log.Println(e.error)
				}
				os.Exit(e.exitCode)
			}
			log.Fatal(err)
		}
		os.Exit(0)
	}
	log.Printf("%s: unknown subcommand %q", cmdName, name)
	log.Fatalf("Run '%s help' for usage.", cmdName)
}

// usageError is an error type that subcommands can return in order to signal
// that a usage error has occurred.
type usageError struct {
	error
}

// exitCodeError is an error type that subcommands can return in order to
// specify the exact exit code.
type exitCodeError struct {
	error
	exitCode int
}