package main
import (
  "errors"
  "fmt"
  "strings"
  "gopkg.in/ini.v1"
  "time"
)

// FromBot messages from the Bot
type FromBot struct {
  message string
  user string
}

// ToBot struct
type ToBot struct {
  message string
  user string
  channel string
}

// Controller struct
type Controller struct {
  frombot chan FromBot
  tobot chan ToBot
  config *ini.File
  bot *SlackBot
}

func (c *Controller) startBot() (error) {
  var err error
  c.bot, err = NewSotdBot(c.config, c.frombot, c.tobot)
  if err != nil {
    return err
  }
  go c.bot.Run()
  return err
}

func (c *Controller) start() {
  c.frombot  = make(chan FromBot, 100) 
  c.tobot = make(chan ToBot, 100)
  fmt.Println("starting bot")
  err := c.startBot()
  fmt.Println("bot started")
  if err != nil {
    panic(err)
  }
  c.mainloop()
}

func (c *Controller) mainloop() {
  for {
    select {
    case in  := <- c.frombot :
      fmt.Printf("%+v\n", in)
      c.Commands(in) 
    case <-time.After(5 * time.Second):
      fmt.Println("Tick Tock")
    }
  }
}


// AddSong blah
// FIXME we need a jukebox too =)
func (c *Controller) AddSong(in FromBot, args string) {
  data, err := ParseStrIntoMap(in.message)

  if err != nil {
    panic("Error adding " + in.message)
  }
  res := []string{}
  for key, element := range data {
    res = append(res, key + ":" + element)
  }

  c.Tell(in.user, "Adding song" + strings.Join(res, ", "))
  // FIXME: add the song to jukebox here
}

func (c *Controller) Tell(user string, message string) {
  c.tobot <- ToBot { message: message , user: user }
}

func (c *Controller) Channels(in FromBot, args string) {
  channels, err := c.bot.Channels()
  if err != nil {
    fmt.Println("error")
  }

  chans := []string{}

  for _,channel := range channels {
    chans = append(chans, "#" + channel.Name)
  }
  str := strings.Join(chans,", ")
  c.Tell(in.user, "I am in channels: " + str)
}

// Commands Here we strip off the first atom as the wanted command
// and pack the rest into a string
func (c *Controller) Commands(in FromBot)  {
  fmt.Println("Parsing command " + in.message)
  parsed := strings.SplitN(in.message," ", 2)
  cmd := parsed[0]
  args := ""

  if len(parsed) == 2 {
    args = parsed[1]
  }

  // FIXME This should be a function table, not a switch
  // FIXME This should be in a controller together bot and jukebox together
  switch(cmd) {
    case "channels": 
    c.Channels(in, args)

  case "add":
    c.AddSong(in,args)
  }
}

// ParseStrIntoMap way to do this.
func ParseStrIntoMap(in string) (map[string]string, error) {
  answer := make(map[string]string)
  atoms := strings.Split(in,";")
  for _,atom := range atoms {
    subs := strings.SplitN(atom,"=", 2)
    if len(subs) < 2 {
      return nil, errors.New("Incorrect song format. Please ask me for help \""+atom+"\"")
    }
    answer[strings.TrimSpace(subs[0])] = strings.TrimSpace(subs[1])
  }

  return answer, nil
}
