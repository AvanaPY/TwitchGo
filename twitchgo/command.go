package twitchgo

type Command struct {
	Name     string
	function func(args []string) (string, error)
}

func NewCommand(name string, f func(args []string) (string, error)) *Command {
	com := &Command{
		Name:     name,
		function: f,
	}
	return com
}

func (com *Command) Construct(args []string) (string, error) {
	output, err := com.function(args)
	return output, err
}
