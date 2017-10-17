package main

import (
	"os"

	"code.cloudfoundry.org/perm/cmd"

	"github.com/jessevdk/go-flags"
)

type options struct {
	Serve   cmd.ServeCommand   `command:"serve" alias:"s" description:"Start the server"`
	Migrate cmd.MigrateCommand `command:"migrate" alias:"m" description:"Migrate the database"`
}

func main() {
	parserOpts := &options{}
	parser := flags.NewParser(parserOpts, flags.Default)
	parser.NamespaceDelimiter = "-"

	// Show actual help message when no command specified
	// Instead of the default unhelpful help message
	parser.SubcommandsOptional = true
	parser.CommandHandler = func(command flags.Commander, args []string) error {
		if command == nil {
			parser.WriteHelp(os.Stderr)
			os.Exit(1)
		}

		err := command.Execute(args)

		// go-flags prints the error itself, but we should already be logging it
		if err != nil {
			os.Exit(1)
		}
		return nil
	}

	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
}
