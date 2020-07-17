package twitchgo

type Command struct {
	Name     string
	function func(ctx *Context) (string, error)
}

func NewCommand(name string, f func(ctx *Context) (string, error)) *Command {
	com := &Command{
		Name:     name,
		function: f,
	}
	return com
}

func (com *Command) Construct(ctx *Context) (string, error) {
	output, err := com.function(ctx)
	return output, err
}
