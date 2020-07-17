package twitchlib

import (
    "regexp"
    "strings"
)

type MsgType int

const (
    PrivMsg MsgType = iota
    Unknown         = iota
)

type Context struct {
    ORG         string      // Original message from twitch

    Sender      string      // Who sent?
    Channel     *Channel    // In which channel?
    MsgType     MsgType     // PRIVMSG?
    Msg         string      // Message

    IsCommand   bool        // Is this a command?
    CommandName string      // Name of command
    CommandArgs []string    // Arguments for the command

    Valid       bool        // Is valid message?
    IsPing      bool        // Is this a PING ?
}


func NewContext(s string, cl *Client) *Context {
    c := &Context {
        ORG:        s,
        MsgType:    Unknown,
        IsCommand:  false,
        Valid:      true,
        IsPing:     false,
    }
    f := []func(c *Context, s string, cl *Client)(bool){ checkPrivmsg, checkPing}
    var done bool = false
    for i := 0; !done && i < len(f); i+=1 {
        done = f[i](c, s, cl)
    }
    if !done {          // If no function matched the input, assume we got an incorrect message or something we can't handle
        c.Valid = false
    }
    return c
}

func checkPrivmsg(c *Context, s string, cl *Client) bool {
    expr, _ := regexp.Compile(":(.+?)!.+ (.+?) #(.+?) :(.+)")
    data := expr.FindStringSubmatch(s[:len(s)-2])   // Trimming the last two characters \r\n is important
                                                    // Why? I hear you ask? Because it doesn't work otherwise
    if len(data) > 0 {
        c.Sender    = data[1]
        c.MsgType   = PrivMsg
        c.Channel   = cl.Channel(data[3])
        c.Msg       = data[4]
        c.IsCommand = strings.HasPrefix(c.Msg, cl.CommandPrefix)
        args := strings.Fields(c.Msg)
        c.CommandName = args[0][1:]
        c.CommandArgs = args[1:]
        return true
    }
    return false
}

func checkPing(c *Context, s string, cl *Client) bool {
    expr, _ := regexp.Compile("^PING")
    mtch    := expr.FindAllString(s, 1)
    c.IsPing = mtch != nil
    return c.IsPing
}
