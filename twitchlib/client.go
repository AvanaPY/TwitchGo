package twitchlib

import (
    "fmt"
    "net"
    "bufio"
    "errors"
    "strconv"
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

func NewClient(pass string, nick string, prefix string) (c *Client) {
    c = new(Client)
    c.HOST  = "irc.twitch.tv"
    c.PORT  = "6667"
    c.Active = true
    c.commandMap = make(map[string]*Command)
    c.cmdPrefix = prefix

    var err error
    c.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%s", c.HOST, c.PORT))

    if err != nil {
        fmt.Printf("An error occured when dialing the host: %s\n", err)
        return
    } else {
        fmt.Printf("Connected successfully to %s:%s\n", c.HOST, c.PORT)
    }
    c.reader = bufio.NewReader(c.conn)

    err = c.authenticate(pass, nick)
    if err != nil {
        fmt.Printf("An error occured when authenticating: %s\n", err)
        c.Active = false
    }else {
        fmt.Printf("Authentication successful: %s\n", nick)
    }
    return c
}

func (c *Client) Log(s string) {
    fmt.Print(s)
}

func (c *Client) Logf(format string, a ...interface{}) {
    s := fmt.Sprintf(format+"\n", a...)
    c.Log(s)
}

func (c *Client) LogContext(context *Context) {
    var maxLen int          = 10
    var maxLenStr string    = strconv.Itoa(maxLen)

    sender  := context.Sender
    channel := context.Channel
    if len(context.Sender) > maxLen {
        sender = context.Sender[:maxLen-3] + "..."
    }
    if len(context.Channel) > maxLen{
        channel = context.Channel[:maxLen-3] + "..."
    }

    if context.Valid {
        c.Logf("%-" + maxLenStr+ "s [ %-" + maxLenStr + "s ]: %s", sender, channel, context.Msg)
    }
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
    c.Logf("Joined channel %s", channel)
}

func (c *Client) PartChannel(channel string) {
    c.WriteString(fmt.Sprintf("PART #%s", channel))
    c.Logf("Parted channel %s", channel)
}

func (c *Client) Send(channel string, message string) {
    c.WriteString(fmt.Sprintf("PRIVMSG #%s :%s", channel, message))
}

func (c *Client) Start() {
    c.startReadLoop()
}

func (c *Client) Read() (s string, err error) {
    s, err = c.reader.ReadString('\n')
    return
}

func (c *Client) startReadLoop() {
    for ;c.Active; {
        s, err := c.Read()
        if err == nil {
            context := NewContext(s, c.cmdPrefix)
            c.LogContext(context)


            go c.HandleContext(context)
        }else {
            c.Logf("Error occured when reading: %s", err)
        }
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
    c.Logf("Invalid command name: %s", name)
}

func (c *Client) HandleInvalidCommandCall (com *Command, err error) {
    c.Logf("Could not execute command %s: %s", com.Name, err)
}
