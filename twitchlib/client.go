package twitchlib

import (
    "fmt"
    "net"
    "bufio"
    "errors"
)


type Client struct {
    HOST        string
    PORT        string

    conn        net.Conn
    reader      *bufio.Reader

    Active      bool

    cmdPrefix   string
    commandMap  map[string]*Command
}

func NewClient(pass string, nick string, prefix string) *Client {
    var c *Client = new(Client)
    c.HOST  = "irc.twitch.tv"
    c.PORT  = "6667"
    c.Active = true
    c.commandMap = make(map[string]*Command)
    c.cmdPrefix = prefix

    var err error
    c.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%s", c.HOST, c.PORT))

    if err != nil {
        fmt.Println("An error occured when dialing the host")
    }
    c.reader = bufio.NewReader(c.conn)

    err = c.authenticate(pass, nick)
    if err != nil {
        fmt.Printf("An error occured when authenticating: %s\n", err)
        c.Active = false
    }
    return c
}

func (c *Client) authenticate(pass string, nick string) error {
    var auth string = fmt.Sprintf("PASS %s\r\nNICK %s\r\n", pass, nick)
    _, err := c.WriteBytes([]byte(auth))
    return err
}

func (c *Client) WriteBytes(bytes []byte) (int, error) {
    if c.conn == nil{
        return 0, errors.New("Conn is not initialized")
    }
    wb, err := c.conn.Write(bytes)
    return wb, err
}

func (c *Client) WriteString(s string) (int, error){
    s += "\r\n"
    wb, err := c.WriteBytes([]byte(s))
    return wb, err
}

func (c *Client) Join(channels []string) {
    for _, channels := range channels {
        c.JoinChannel(channels)
    }
}

func (c *Client) JoinChannel(channel string) {
    c.WriteString(fmt.Sprintf("JOIN #%s", channel))
}

func (c *Client) PartChannel(channel string) {
    c.WriteString(fmt.Sprintf("PART #%s", channel))
}

func (c *Client) Send(channel string, message string) {
    c.WriteString(fmt.Sprintf("PRIVMSG #%s :%s", channel, message))
}

func (c *Client) Start() {
    c.startReadLoop()
}

func (c *Client) Read() string {
    s, err := c.reader.ReadString('\n')
    if err != nil {
        fmt.Printf("Error occured during reading connection, %s", err)
    }
    return s
}

func (c *Client) startReadLoop() {
    fmt.Println("Starting read loop")
    for {
        s := c.Read()
        context := NewContext(s, c.cmdPrefix)
        c.HandleContext(context)
    }
}
// Commands

func (c *Client) CreateCommand(name string, msg string) {
    com := NewCommand(name, msg)
    c.commandMap[name] = com
}

func (c *Client) CallCommand(name string, channel string, args []string) {
    if com, ok := c.commandMap[name]; ok {
        resp, err := com.Construct(args)
        if err == nil {
            c.Send(channel, resp)
        } else {
            c.HandleInvalidCommandCall(com, err)
        }
    } else {
        c.HandleInvalidCommandName(name)
    }
}


// Handlers
func (c *Client) HandleContext (ctx *Context) {
    if ctx.MsgType == "PING" {
        fmt.Printf("Responded to PING\n")
        c.WriteString("PONG :tmi.twitch.tv")
    } else if ctx.MsgType == "PRIVMSG" {
        if ctx.IsCommand {
            c.CallCommand(ctx.CommandName, ctx.Channel, ctx.CommandArgs)
        }
    }
    return
}

func (c *Client) HandleInvalidCommandName (name string) {
    fmt.Printf("Invalid command name: %s\n", name)
}

func (c *Client) HandleInvalidCommandCall (com *Command, err error) {
    fmt.Printf("Could not execute command %s: %s\n", com.Name, err)
}
