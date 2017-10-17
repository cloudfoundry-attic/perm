package cmd

type MigrateCommand struct {
	Logger LagerFlag

	SQL SQLFlag `group:"SQL" namespace:"sql"`
}

func (cmd MigrateCommand) Execute([]string) error {
	return nil
}
