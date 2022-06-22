package main
import (
  "errors"
  "fmt"
  "strings"
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
}

func (c *Controller) start() (chan FromBot, chan ToBot) {
  c.frombot  = make(chan FromBot, 100) 
  c.tobot = make(chan ToBot, 100)
  go c.mainloop()
  return c.frombot, c.tobot
}

func (c *Controller) mainloop() {
  for {
    select {
    case in  := <- c.frombot :
      fmt.Printf("%+v\n", in)
      c.Commands(in.message)
      c.tobot <- ToBot { message: "I got message:" + in.message,  user: in.user }
    case <-time.After(5 * time.Second):
      fmt.Println("Tick Tock")
    }
  }
}


// AddSong blah
// FIXME We need to know the user
// FIXME we need a jukebox too =)
func AddSong(input string) (string, error) {
  data, err := ParseStrIntoMap(input)

  if err != nil {
    return "Error adding "+input, err
  }
  fmt.Println("Adding song")
  // FIXME: add the song to jukebox here
  return fmt.Sprint(data), err
}

// Commands Here we strip off the first atom as the wanted command
// and pack the rest into a string
func (c *Controller) Commands(msg string) (string, error){
  fmt.Println("Parsing command " + msg)
  parsed := strings.SplitN(msg," ", 2)
  cmd := parsed[0]
  args := ""

  if len(parsed) == 2 {
    args = parsed[1]
  }

  // FIXME This should be a function table, not a switch
  // FIXME This should be in a controller together bot and jukebox together
  switch {
  case cmd == "add":
    return AddSong(args)
  }
  return "", errors.New("Unknown command " + cmd)
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
