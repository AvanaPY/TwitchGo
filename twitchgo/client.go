package twitchgo

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Client struct {
	HOST string
	PORT string

	conn   net.Conn
	reader *bufio.Reader

	Active bool

	CommandPrefix string
	commandMap    map[string]*Command
	channels      map[string]*Channel
}

func NewClient(pass string, nick string, prefix string) (c *Client) {
	c = &Client{
		HOST:          "irc.twitch.tv",
		PORT:          "6667",
		Active:        true,
		commandMap:    make(map[string]*Command),
		CommandPrefix: prefix,
		channels:      make(map[string]*Channel),
	}

	var err error
	c.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%s", c.HOST, c.PORT))

	if err != nil {
		fmt.Printf("An error occured when dialing the host: %s\n", err)
		c.Active = false
		return
	}
	c.reader = bufio.NewReader(c.conn)

	err = c.authenticate(pass, nick)
	if err != nil {
		fmt.Println(err)
		c.Active = false
	}
	return c
}

func (c *Client) authenticate(pass string, nick string) error {
	var auth string = fmt.Sprintf("PASS %s\r\nNICK %s\r\n", pass, nick)
	_, err := c.WriteBytes([]byte(auth))
	resp, _ := c.Read()
	resp = strings.Split(resp, ":")[2]

	success := resp == "Welcome, GLHF!"
	if success {
		// Wait for authentication messages
		for i := 0; i < 6; i++ {
			resp, _ = c.Read()
		}
		return err
	}
	return errors.New("Authentication failed, invalid Oauth.")
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
	var maxLen int = 10
	var maxLenStr string = strconv.Itoa(maxLen)

	if context.MsgType == PrivMsg {
		sender := LimitStringLength(context.Sender, maxLen, "...")
		channel := LimitStringLength(context.Channel.Name, maxLen, "...")

		if context.Valid {
			c.Logf("%-"+maxLenStr+"s [ %-"+maxLenStr+"s ]: %s", sender, channel, context.Msg)
		}
	} else {
		c.Logf(context.ORG)
	}
}

// TWITCH FUNCS

func (c *Client) Channel(name string) *Channel {
	if val, ok := c.channels[name]; ok {
		return val
	}
	return nil
}

// WRITERS

func (c *Client) WriteBytes(bytes []byte) (int, error) {
	if c.conn == nil {
		return 0, errors.New("Conn is not initialized")
	}
	wb, err := c.conn.Write(bytes)
	return wb, err
}

func (c *Client) WriteString(s string) (int, error) {
	s += "\r\n"
	wb, err := c.WriteBytes([]byte(s))
	return wb, err
}

// COMMANDS

func (c *Client) Join(channels []string) {
	for _, channels := range channels {
		c.JoinChannel(channels)
	}
	for i := 0; i < 3*len(channels); i++ {
		c.Read()
	}
}

func (c *Client) JoinChannel(channel string) {
	ch := NewChannel(channel, c)
	c.channels[ch.Name] = ch
	c.WriteString(fmt.Sprintf("JOIN #%s", ch.Name))
	c.Logf("Joined %s", channel)
}

func (c *Client) PartChannel(channel string) {
	c.WriteString(fmt.Sprintf("PART #%s", channel))
}

func (c *Client) Send(channel string, message string) {
	c.WriteString(fmt.Sprintf("PRIVMSG #%s :%s", channel, message))
}

// GENERALS

func (c *Client) Read() (string, error) {
	s, err := c.reader.ReadString('\n')
	s = strings.TrimSuffix(s, "\r\n")
	return s, err
}

func (c *Client) Start() {
	c.Logf("Started Client.")
	c.startReadLoop()
}

func (c *Client) startReadLoop() {
	for c.Active {
		s, err := c.Read()
		if err == nil {
			context := NewContext(s, c)
			go c.HandleContext(context)
			c.LogContext(context)
		} else {
			c.Logf("Error occured when reading: %s", err)
		}
	}
}

// User Commands

func (c *Client) CreateCommandFunction(name string, f func(ctx *Context) (string, error), channel ...string) {
	com := NewCommand(name, f)
	if len(channel) > 0 {
		if ch, ok := c.channels[channel[0]]; ok {
			ch.AddCommand(com)
		}
	} else {
		c.commandMap[com.Name] = com
	}
}

func (c *Client) CreateCommandString(name string, msg string, channel ...string) {
	f := func(ctx *Context) (string, error) {
		return msg, nil
	}
	c.CreateCommandFunction(name, f, channel...)
}

func (c *Client) CallCommand(ctx *Context) {
	com := c.getCommand(ctx.CommandName, ctx.Channel.Name)
	if com != nil {
		resp, err := com.Construct(ctx)
		if err == nil {
			c.Send(ctx.Channel.Name, resp)
		} else {
			c.HandleInvalidCommandCall(ctx, err)
		}
	} else {
		c.HandleInvalidCommandName(ctx)
	}
}

func (c *Client) getCommand(name string, channel string) *Command {

	// Check if command is a channel-specific command
	if ch, ok := c.channels[channel]; ok {
		if com, comOk := ch.GetCommand(name); comOk {
			return com
		}
	}
	// If it's not a channel command, check if it's a global command
	if com, comOk := c.commandMap[name]; comOk {
		return com
	}
	return nil
}

// Handlers
func (c *Client) HandleContext(ctx *Context) {
	if ctx.IsPing {
		c.HandlePing(ctx.ORG)
	} else if ctx.MsgType == PrivMsg {
		if ctx.IsCommand {
			c.CallCommand(ctx)
		}
	}
	return
}

func (c *Client) HandlePing(ping string) {
	var resp string = "PONG :tmi.twitch.tv"
	c.Logf("Responded to %q with %q", ping, resp)
	c.WriteString(resp)
}

func (c *Client) HandleInvalidCommandName(ctx *Context) {
	c.Logf("Invalid command name: %s", ctx.CommandName)
}

func (c *Client) HandleInvalidCommandCall(ctx *Context, err error) {
	c.Logf("Could not execute command %s: %s", ctx.CommandName, err)
}
