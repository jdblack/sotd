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
  jukebox *Jukebox
}

func (c *Controller) start() {
  var err error
  fmt.Println("Controller  starting")
  c.bot, err = NewBot()
  if err != nil {
    panic(err)
  }
  fmt.Println("Bot started")
  c.bot.Run()
  c.jukebox, err = NewJukebox()
  if err != nil {
    panic(err)
  }
  fmt.Println("Jukebox started")
  c.mainloop()

}

func (c *Controller) mainloop() {
  for {
    select {
    case in  := <- c.bot.frombot :
      c.Commands(in) 
    case <-time.After(5 * time.Second):
      fmt.Println("Tick Tock")
    }
  }
}


func (c *Controller) sendHelp(in FromBot, args string) {
  c.Tell(in.user, "I don't understand what you meant by" + in.message)
}

func (c *Controller) showPlaylist(name string) (string){
    playlist,err := c.jukebox.GetPlaylist(name)
    if err != nil {
      return fmt.Sprintf("I could not find %s : %s", name, err)
    }
    return fmt.Sprintf("Name: #%s\nSchedule: %s", playlist.Channel, playlist.Cron)
}

func (c *Controller) playlist(in FromBot, message string) {
  cmd, args, _ := strings.Cut(message," ")
  switch(cmd) {
  case "list":
    playlists := c.jukebox.GetPlaylists() 
    for _, pl := range playlists {
      c.Tell(in.user, fmt.Sprintf("#%s : %s", pl.Channel, pl.Cron))
    }
  case "show":
    c.Tell(in.user, c.showPlaylist(args))
    return
  }
}

func (c *Controller) hello(in FromBot, args string) {
  c.Tell(in.user, "Hello back to you!")
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

func (c *Controller) addplaylist(in FromBot, args string) {
  channels, _ := c.bot.Channels()
  c.Tell(in.user, fmt.Sprintf( "I am in %d channels: ", len(channels)))
  playlist, err := c.jukebox.GetPlaylist(args)
  if err != nil {
    c.Tell(in.user, fmt.Sprintf("I got error : %v", err))
  }
  c.Tell(in.user,fmt.Sprintf("I have playlist :  %v", playlist))


}


func (c *Controller) botChannels(in FromBot, args string) {
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
  fmt.Printf("Parsing command: %v\n", in)
  cmd, args, _ := strings.Cut(in.message," ")

  // FIXME This should be a function table, not a switch
  // FIXME This should be in a controller together bot and jukebox together
  switch(cmd) {
    case "where": c.botChannels(in, args)
    case "subscribe": c.addplaylist(in, args)
    case "hello"   : c.hello(in,args)
    case "hi"      : c.hello(in,args)
    case "playlist" : c.playlist(in,args)
    case "add"     : c.addSong(in,args)
    default        : c.sendHelp(in,args)
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
