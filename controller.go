package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Controller struct
type Controller struct {
	bot         *SlackBot
	jukebox     *Jukebox
	mainMenu    map[string]menuItem
	idleTimeout int
}

type menuItem struct {
	h string                // The help to give
	f func(FromBot, string) // Callout function
}

func (c *Controller) start() {
	c.idleTimeout = 60
	c.mainMenu = map[string]menuItem{
		"hello":    {f: c.hello, h: "Say hello"},
		"hi":       {f: c.hello, h: "Alias for hello"},
		"add":      {f: c.addSong, h: "Add a song to a play"},
		"playlist": {f: c.playlist, h: "Run a playlist subcommand"},
		"channels": {f: c.listChannels, h: "List Channels"},
	}
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
		case in := <-c.bot.frombot:
			c.Commands(in)
		case play := <-c.jukebox.Playset:
			s := play.song
			msg := fmt.Sprintf("Please thank <@%s> for today's SOTD %s\n", s.User, s.URL)
			msg += s.Description
			c.Tell(play.channel, msg)
		case <-time.After(time.Duration(c.idleTimeout) * time.Second):
			fmt.Printf("All quiet here for the last %d seconds\n", c.idleTimeout)
		}
	}
}

func (c *Controller) sendHelp(in FromBot, args string) {
	c.Tell(in.user, "I don't understand what you meant by"+in.message)
}

func (c *Controller) showPlaylist(in FromBot, args string) {
	_, ch := c.bot.ParseChannel(args)
	pl, err := c.jukebox.GetPlaylist(ch)
	if err != nil {
		c.Tell(in.user, fmt.Sprintf("I could not find %s : %s", ch, err))
		return
	}
	m := []string{
		fmt.Sprintf("Request: %s", ch),
		fmt.Sprintf("Name: %s", pl.Channel),
		fmt.Sprintf("Play Schedule: %s", pl.Cron),
	}
	for _, s := range pl.Songs {
		m = append(m, fmt.Sprintf("<@%s> : %s %s", s.User, s.URL, s.Description))
	}
	c.Tell(in.user, strings.Join(m, "\n"))
}

func (c *Controller) playlist(in FromBot, message string) {
	cmd, args, _ := strings.Cut(message, " ")
	switch cmd {
	case "list":
		playlists := c.jukebox.GetPlaylists()
		for _, pl := range playlists {
			c.Tell(in.user, fmt.Sprintf("#%s : %s", pl.Channel, pl.Cron))
		}
	case "show":
		c.showPlaylist(in, args)
		return
	}
}

func (c *Controller) hello(in FromBot, args string) {
	c.Tell(in.user, "Hello back to you!")
}

func (c *Controller) contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// AddSong blah
func (c *Controller) addSong(in FromBot, args string) {
	s := regexp.MustCompile(" +").Split(args, 3)
	_, channel := c.bot.ParseChannel(s[0])

	channels, _ := c.bot.ChannelNames()
	if !c.contains(channels, channel) {
		msg := fmt.Sprintf("You need to invite me to %s first!", s[0])
		c.Tell(in.user, msg)
		return
	}

	if len(s) != 3 {
		c.Tell(in.user, "To add a song, message me again with the following: ")
		msg :=
			"add #channel https://some-song/url " +
				"A description of your song"
		c.Tell(in.user, msg)
		return
	}

	song := Song{
		User:        in.user,
		URL:         s[1],
		Description: s[2],
	}

	err := c.jukebox.AddSong(song, channel)
	if err != nil {
		c.Tell(in.user, "I had trouble adding "+song.URL+" to "+channel)
		return
	}
	c.Tell(in.user, "I have added "+song.URL+" to "+channel)
}

// Tell a user something
func (c *Controller) Tell(user string, message string) {
	c.bot.tobot <- ToBot{message: message, user: user}
}

func (c *Controller) addplaylist(in FromBot, args string) {
	channels, _ := c.bot.Channels()
	c.Tell(in.user, fmt.Sprintf("I am in %d channels: ", len(channels)))
	playlist, err := c.jukebox.GetPlaylist(args)
	if err != nil {
		c.Tell(in.user, fmt.Sprintf("I got error : %v", err))
	}
	c.Tell(in.user, fmt.Sprintf("I have playlist :  %v", playlist))
}

func (c *Controller) listChannels(in FromBot, args string) {
	channels, err := c.bot.ChannelNames()
	chans := []string{}
	if err != nil {
		fmt.Println("error")
	}

	for _, channel := range channels {
		chans = append(chans, "#"+channel)
	}
	c.Tell(in.user, "I am in channels"+strings.Join(chans, ", "))
}

//Commands is the main command handler. To add a command, add to the
//jump table in start()
func (c *Controller) Commands(in FromBot) {
	cmd, args, _ := strings.Cut(in.message, " ")

	if opt, ok := c.mainMenu[cmd]; ok {
		opt.f(in, args)
	} else {
		c.printHelp(in, c.mainMenu)
	}
}

//Display the help for a menu
func (c *Controller) printHelp(in FromBot, menu map[string]menuItem) {
	msg := []string{"Help for  in.message"}

	for name, data := range menu {
		msg = append(msg, fmt.Sprintf("%s : %s", name, data.h))
	}
	c.Tell(in.user, strings.Join(msg, "\n"))
}
