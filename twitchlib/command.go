package twitchlib

import (
    "regexp"
    "strings"
    "strconv"
    "errors"
)
type Command struct {

    Name        string
    Msg         string
}

func NewCommand(name string, msg string) *Command {
    com := new(Command)
    com.Name = name
    com.Msg = msg
    return com
}

func (com *Command) Construct(args []string) (string, error) {
    var out string = ""
    wrds := strings.Fields(com.Msg)

    comExpr, _ := regexp.Compile("^\\$\\{(.+?)\\}$")
    for _, wrd := range wrds {
        mtch := comExpr.FindStringSubmatch(wrd)
        if mtch != nil {
            add, err := parseFunc(mtch[1], args)
            if err != nil{
                return "", err
            }
            out += add + " "
        }else {
            out += wrd + " "
        }
    }
    return out, nil
}

func parseFunc(in string, args []string) (string, error) {
    var val string  = ""
    var err error   = nil

    funcs := []func(in string, args []string)(string, error) {isInt}
    // Loop through all functions and stop if a non-error case occurs
    for i:=0; i < len(funcs) && err == nil; i++{
        val, err = funcs[i](in, args)
    }
    if err != nil {
        return val, err
    }
    return val, nil
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
