package twitchlib

import (
    "fmt"
    "net"
    "bufio"
    "errors"
    "strconv"
)

type Client struct {
    HOST            string
    PORT            string

    conn            net.Conn
    reader          *bufio.Reader

    Active          bool

    CommandPrefix   string
    commandMap      map[string]*Command

    channels        map[string]*Channel
}

func NewClient(pass string, nick string, prefix string) (c *Client) {
    c = &Client {
        HOST:           "irc.twitch.tv",
        PORT:           "6667",
        Active:         true,
        commandMap:     make(map[string]*Command),
        CommandPrefix:  prefix,
        channels:       make(map[string]*Channel),
    }

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

// LOGGING

func (c *Client) Log(s string) {
    if s[len(s)-1] == '\n' {
        fmt.Print(s)
    } else {
        fmt.Println(s)
    }
}

func (c *Client) Logf(format string, a ...interface{}) {
    s := fmt.Sprintf(format, a...)
    c.Log(s)
}

func (c *Client) LogContext(context *Context) {
    var maxLen int          = 10
    var maxLenStr string    = strconv.Itoa(maxLen)

    if context.MsgType == PrivMsg {
        sender  := LimitStringLength(context.Sender      , maxLen, "...")
        channel := LimitStringLength(context.Channel.Name, maxLen, "...")

        if context.Valid {
            c.Logf("%-" + maxLenStr+ "s [ %-" + maxLenStr + "s ]: %s", sender, channel, context.Msg)
        }
    }
}

// TWITCH FUNCS

func (c *Client) authenticate(pass string, nick string) error {
    var auth string = fmt.Sprintf("PASS %s\r\nNICK %s\r\n", pass, nick)
    _, err := c.WriteBytes([]byte(auth))
    return err
}

func (c *Client) Channel(name string) (*Channel) {
    if val, ok := c.channels[name]; ok {
        return val
    }
    return nil
}

// WRITERS

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

// COMMANDS

func (c *Client) Join(channels []string) {
    for _, channels := range channels {
        c.JoinChannel(channels)
    }
}

func (c *Client) JoinChannel(channel string) {
    ch := NewChannel(channel, c)
    c.channels[ch.Name] = ch
    c.WriteString(fmt.Sprintf("JOIN #%s", ch.Name))
}

func (c *Client) PartChannel(channel string) {
    c.WriteString(fmt.Sprintf("PART #%s", channel))
}

func (c *Client) Send(channel string, message string) {
    c.WriteString(fmt.Sprintf("PRIVMSG #%s :%s", channel, message))
}

// GENERALS

func (c *Client) Start() {
    for key, val := range c.channels {
        fmt.Printf("%s: %s\n", key, val.Name)
    }
    c.startReadLoop()
}

func (c *Client) Read() (string, error) {
    return c.reader.ReadString('\n')
}

func (c *Client) startReadLoop() {
    for ;c.Active; {
        s, err := c.Read()
        if err == nil {
            context := NewContext(s, c)
            go c.HandleContext(context)
            c.LogContext(context)
        }else {
            c.Logf("Error occured when reading: %s", err)
        }
    }
}

// User Commands

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
    if ctx.IsPing {
        c.HandlePing(ctx.ORG)
    } else if ctx.MsgType == PrivMsg {
        if ctx.IsCommand {
            c.CallCommand(ctx.CommandName, ctx.Channel.Name, ctx.CommandArgs)
        }
    }
    return
}

func (c *Client) HandlePing (ping string) {
    var resp string = "PONG :tmi.twitch.tv"
    c.Logf("Responded to %q with %q", ping, resp)
    c.WriteString(resp)
}

func (c *Client) HandleInvalidCommandName (name string) {
    c.Logf("Invalid command name: %s", name)
}

func (c *Client) HandleInvalidCommandCall (com *Command, err error) {
    c.Logf("Could not execute command %s: %s", com.Name, err)
}
