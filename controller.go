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
	o string                // help options
	f func(FromBot, string) // Callout function
}

func (c *Controller) start() {
	c.idleTimeout = 60
	c.mainMenu = map[string]menuItem{
		"hello":     {f: c.hello, o: "", h: "Say hello"},
		"add":       {f: c.addSong, o: "CHANNEL* URL [Song description]", h: "Add a song to a play with optional description"},
		"delete":    {f: c.deleteSong, o: "URL", h: "Delete song matching URL"},
		"playlists": {f: c.listPlaylists, o: "", h: "List all running playlists"},
		"load":      {f: c.loadPlaylist, o: "CHANNEL URL", h: "Import a json playlist from an URL"},
		"stop":      {f: c.leaveChannel, o: "CHANNEL", h: "Tell SOTD to remove a playlist for a channel"},
		"show":      {f: c.showPlaylist, o: "CHANNEL", h: "Show the playlist for a given channel"},
		//		"channels":  {f: c.listChannels, h: "List Channels"},
	}
	var err error
	fmt.Println("Controller  starting")
	c.bot, err = NewBot()
	if err != nil {
		panic(err)
	}
	fmt.Println("Bot started")
	c.bot.Run()
	c.jukebox, err = NewJukebox(&Cfg)
	if err != nil {
		panic(err)
	}
	fmt.Println("Jukebox started")
	c.mainloop()

}

func (c *Controller) spinAPlay(play Play) {
	desc := "from"
	s := play.song
	msg := "\n\n\nTime for *Sotd*!"
	if play.backfill {
		msg += " _Your channel is out of songs, so we stole one for ya_ "
	}
	msg += ": " + s.URL
	if len(s.Description) > 1 {
		desc = s.Description
	}
	msg += fmt.Sprintf("\n\n```\n\n         %s -- %s\n\n```\n\n", desc, s.RealName)

	c.Tell(play.channel, msg)
}

func (c *Controller) mainloop() {
	for {
		select {
		case in := <-c.bot.frombot:
			c.Commands(in)
		case play := <-c.jukebox.Playset:
			c.spinAPlay(play)
		case <-time.After(time.Duration(c.idleTimeout) * time.Second):
			fmt.Printf("All quiet here for the last %d seconds\n", c.idleTimeout)
		}
	}
}

func (c *Controller) sendHelp(in FromBot, args string) {
	c.Tell(in.user, "I don't understand what you meant by"+in.message)
}

func (c *Controller) showPlaylist(in FromBot, args string) {
	_, ch := ParseChannel(args)
	pl, err := c.jukebox.GetPlaylist(ch)
	if err != nil {
		c.Tell(in.user, fmt.Sprintf("I could not find %s : %s", ch, err))
		return
	}
	m := []string{
		"> ",
		fmt.Sprintf("*Playlist Channel Name*: %s", pl.Channel),
		fmt.Sprintf("*Play Schedule* : %s", pl.Cron),
		fmt.Sprintf("*%d Songs to come!*\n\n", len(pl.Songs)),
		"> ",
	}
	for _, s := range pl.Songs {
		m = append(m, fmt.Sprintf("<@%s> _(%s)_ : `%s` %s", s.User, s.RealName, s.URL, s.Description))
	}
	c.Tell(in.user, strings.Join(m, "\n"))
}

func (c *Controller) loadPlaylist(in FromBot, args string) {
	if !Cfg.GetBool("insecure_loading") {
		c.Tell(in.user, "Loading is not enabled in the insecure section of the config")
		return
	}

	songs, err := c.jukebox.loadSongs(in, args)
	if err != nil {
		c.Tell(in.user, fmt.Sprintf("There was a problem, but I was able to add %d songs. The error was %s", len(songs), err.Error()))
		return
	}
	c.Tell(in.user, fmt.Sprintf("I loaded %d songs", len(songs)))
}

func (c *Controller) listPlaylists(in FromBot, args string) {
	playlists := c.jukebox.GetPlaylists()
	for _, pl := range playlists {
		c.Tell(in.user, fmt.Sprintf("%s : %s", pl.Channel, pl.Cron))
	}
}

func (c *Controller) leaveChannel(in FromBot, args string) {
	id, name := ParseChannel(args)
	id = strings.Trim(id, "#")

	_, err := c.jukebox.DeleteChannel(in, name)
	if err != nil {
		c.Tell(in.user, err.Error())
		return
	}

	_, err = c.bot.leaveChannel(id)
	if err != nil {
		c.Tell(in.user, err.Error())
		return
	}
	c.Tell(in.user, "I have left "+name)
	return
}

func (c *Controller) hello(in FromBot, args string) {
	c.Tell(in.user, "Hello back!")
}

func (c *Controller) contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// deleteSong is given an url and will delete those songs
func (c *Controller) deleteSong(in FromBot, args string) {

	count, err := c.jukebox.DeleteSongByURL(args)
	if err != nil {
		c.Tell(in.user, "I wasnt able to delete that")
		return
	}
	c.Tell(in.user, fmt.Sprintf("I deleted %d songs", count))
}

// AddSong blah
func (c *Controller) addSong(in FromBot, args string) {
	s := regexp.MustCompile(" +").Split(args, 3)
	_, channel := ParseChannel(s[0])

	channels, _ := c.bot.ChannelNames()
	fmt.Printf("I want %s and we have %+v\n", channel, channels)
	//	if !c.contains(channels, channel) {
	//		msg := fmt.Sprintf("You need to invite me to %s first!", s[0])
	//		c.Tell(in.user, msg)
	//		return
	//	}

	if len(s) < 2 {
		c.Tell(in.user, "To add a song, message me again with the following: ")
		msg :=
			"add #channel https://some-song/url " +
				"_An optional description of your song_"
		c.Tell(in.user, msg)
		return
	}
	desc := ""
	if len(s) > 2 {
		desc = s[2]
	}

	song := Song{
		User:        in.user,
		RealName:    in.fullName,
		URL:         s[1],
		Description: desc,
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
		msg = append(msg, fmt.Sprintf("*%s %s*\n\t\t%s\n", name, data.o, data.h))
	}
	c.Tell(in.user, strings.Join(msg, "\n"))
}
