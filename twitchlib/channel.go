package twitchlib

type Channel struct {
    client      *Client      // So we can interact with the client

    Name        string      // Name of channel
    commands    map[string]*Command     // Channel-specific commands
}


func NewChannel(name string, cl *Client) *Channel {
    c := new(Channel)
    c.Name = name
    c.client = cl

    return c
}
