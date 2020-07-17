package twitchgo

import "fmt"

type Channel struct {
    client      *Client      // So we can interact with the client

    Name        string      // Name of channel
    commands    map[string]*Command     // Channel-specific commands
}

func NewChannel(name string, cl *Client) *Channel {
    c := &Channel {
        Name: name,
        client: cl,
        commands: make(map[string]*Command),
    }
    return c
}

func (ch *Channel) CreateCommand(name string, msg string) {
    com := NewCommand(name, msg)
    ch.commands[name] = com

    fmt.Printf("Added command %q to channel %q\n", name, ch.Name)
}

func (ch *Channel) Command(name string) (*Command, bool) {
    if com, ok := ch.commands[name]; ok {
            return com, ok
    }
    return nil, false
}
