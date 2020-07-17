package twitchgo

import (
    "strconv"
    "errors"
)
type Command struct {

    Name        string
    Msg         string
}

func NewCommand(name string, msg string) *Command {
    com := &Command{
        Name: name,
        Msg:  msg,
    }
    return com
}

func (com *Command) Construct(args []string) (string, error) {
    var (
        output      string  = ""
        funcString          = ""
        startedFunc bool    = false
    )
    // Loop over all characters and see if the pattern ${...} is found
    // If we find a pattern evaluate it as a func string, otherwise just append the character to the output
    //
    // Result: Replace all ${...} patterns with function outputs
    for _, char := range com.Msg {
        if char == '$'{
            startedFunc = true
        } else if char == '{'{
            continue
        } else if char == '}' {
            startedFunc = false
            o, err := EvaluateFuncString(funcString, args)
            if err != nil {
                return "", err
            }
            output += o
            funcString = ""
        } else if startedFunc {
            funcString += string(char)
        } else {
            output += string(char)
        }
    }
    return output, nil
}

func EvaluateFuncString(in string, args []string) (string, error) {
    var (
        val string  = ""
        err error   = nil
    )

    // Loop through all functions and stop if a non-error case occurs
    funcs := []func(in string, args []string)(string, error) {isInt}
    for i:=0; i < len(funcs) && err == nil; i++{
        val, err = funcs[i](in, args)
    }
    return val, err
}

func isInt(in string, args []string) (string, error) {
    i, err := strconv.Atoi(in)
    if err != nil {
        return "", err
    }
    if i >= len(args){
        return "", errors.New("Invalid command arguments: Out of range")
    }
    return args[i], nil
}
