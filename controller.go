package main
import (
  "errors"
  "fmt"
  "strings"
  "time"
)

// Controller struct
type Controller struct {
  bot *SlackBot
}

func (c *Controller) newBot() (error) {
  var err error
  c.bot, err = NewSotdBot()
  if err != nil {
    return err
  }
  go c.bot.Run()
  return err
}

func (c *Controller) start() {
  err := c.newBot()
  if err != nil {
    panic(err)
  }
  c.mainloop()
}

func (c *Controller) mainloop() {
  for {
    select {
    case in  := <- c.bot.frombot :
      fmt.Printf("%+v\n", in)
      c.Commands(in) 
    case <-time.After(5 * time.Second):
      fmt.Println("Tick Tock")
    }
  }
}


// AddSong blah
// FIXME we need a jukebox too =)
func (c *Controller) addSong(in FromBot, args string) {
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

// Tell a user something
func (c *Controller) Tell(user string, message string) {
  c.bot.tobot <- ToBot { message: message , user: user }
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

func (c *Controller) parseCommand(in string) (string, string){
  fmt.Println("Parsing command " + in)
  parsed := strings.SplitN(in," ", 2)
  cmd := parsed[0]
  args := ""

  if len(parsed) == 2 {
    args = parsed[1]
  }
  return cmd,args
}

// Commands Here we strip off the first atom as the wanted command
// and pack the rest into a string
func (c *Controller) Commands(in FromBot)  {
  cmd, args := c.parseCommand(in.message)

  // FIXME This should be a function table, not a switch
  // FIXME This should be in a controller together bot and jukebox together
  switch(cmd) {
    case "channels": c.Channels(in, args)
    case "add"     : c.addSong(in,args)
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
