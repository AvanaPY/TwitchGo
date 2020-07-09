package twitchlib

import (
    "regexp"
    "strings"
)

type Context struct {
    Sender      string
    Channel     string
    MsgType     string
    Msg         string
    IsCommand   bool
    CommandName string
    CommandArgs []string
}


func NewContext(s string, comPrefix string) *Context {
    c := new(Context)
    c.IsCommand = false
    c.CommandName = ""

    f := []func(c *Context, s string, prefix string)(bool){ checkPrivmsg, checkPing, setNone}
    done := false
    for i := 0; !done; i+=1 {
        done = f[i](c, s, comPrefix)
    }
    return c
}

func checkPrivmsg(c *Context, s string, comPrefix string) bool {
    expr, _ := regexp.Compile(":(.+?)!.+ (.+?) #(.+?) :(.+)")
    data := expr.FindStringSubmatch(s[:len(s)-2])   // Trimming the last two characters \r\n is important
                                                    // Why? I hear you ask? Because it doesn't work otherwise
    if len(data) > 0 {
        c.Sender    = data[1]
        c.MsgType   = data[2]
        c.Channel   = data[3]
        c.Msg       = data[4]
        c.IsCommand = strings.HasPrefix(c.Msg, comPrefix)
        args := strings.Fields(c.Msg)
        c.CommandName = args[0][1:]
        c.CommandArgs = args[1:]
        return true
    }
    return false
}

func checkPing(c *Context, s string, comPrefix string) bool {
    expr, _ := regexp.Compile("^PING")
    mtch    := expr.FindAllString(s, 1)
    if mtch != nil {
        c.Sender    = "twitch"
        c.MsgType   = "PING"
        c.Channel   = "twitch"
        c.Msg       = "twitch"
        return true
    }
    return false
}

func setNone(c *Context, s string, comPrefix string) bool {
    c.Sender    = "none"
    c.Channel   = "none"
    c.MsgType   = "none"
    c.Msg       = "none"
    return true
}
