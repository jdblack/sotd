package main

import (
	"errors"
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
		"add":      {f: c.addSong, o: "CHANNEL URL [Song description]", h: "Add a song to a play with optional description"},
		"channels": {f: c.listPlaylists, o: "", h: "Show all my channels"},
		"delete":   {f: c.deleteSong, o: "URL", h: "Delete song matching URL"},
		"load":     {f: c.loadPlaylist, o: "CHANNEL URL", h: "Import a json playlist from an URL (requires insecure.load enabled)"},
		"schedule": {f: c.scheduleChannel, o: "CHANNEL CRON", h: "Change the schedule for a channel"},
		"stop":     {f: c.leaveChannel, o: "CHANNEL", h: "Tell SOTD to remove a playlist for a channel"},
		"show":     {f: c.showPlaylist, o: "CHANNEL", h: "Show the playlist for a given channel"},
		"hello":    {f: c.hello, o: "", h: "Say hello"},
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
		fmt.Sprintf("*Play Schedule* : `%s`", pl.Cron),
		fmt.Sprintf("*%d Songs to come!*\n\n", len(pl.Songs)),
		"> ",
	}
	if len(pl.Songs) > 0 {
		for _, s := range pl.Songs {
			m = append(m, fmt.Sprintf("<@%s> _(%s)_ : `%s` %s", s.User, s.RealName, s.URL, s.Description))
		}
	}

	if len(pl.History) > 0 {
		msg := "`%s` originally by %s on %s : %s"
		m = append(m, fmt.Sprintf("\n*%d Past Songs:*\n", len(pl.History)))
		for i := len(pl.History) - 1; i >= 0; i-- {
			s := pl.History[i]
			when := s.CreatedAt.Format(time.UnixDate)
			m = append(m, fmt.Sprintf(msg, s.URL, s.RealName, when, s.Description))
		}

	}

	c.Tell(in.user, strings.Join(m, "\n"))
}

func (c *Controller) loadPlaylist(in FromBot, args string) {
	insecure, _ := Cfg.GetBool("insecure_loading")
	if insecure != true {
		c.Tell(in.user, "Loading is not enabled in the insecure section of the config")
		return
	}

	songs, err := c.jukebox.loadSongs(in, args)
	if err != nil {
		msg := "I had trouble but I was able to add %d songs. The error was %s"
		c.Tell(in.user, fmt.Sprintf(msg, len(songs), err.Error()))
		return
	}
	c.Tell(in.user, fmt.Sprintf("I loaded %d songs", len(songs)))
}

func (c *Controller) listPlaylists(in FromBot, args string) {
	playlists := c.jukebox.GetPlaylists()
	if len(playlists) == 0 {
		c.Tell(in.user, "No active playlists yet! Please add some songs!")
	}
	for _, pl := range playlists {
		fmt.Printf("%+v\n", pl)
		c.Tell(in.user, fmt.Sprintf("Name: %s Queue: %d Schedule: %s", pl.Channel, len(pl.Songs), pl.Cron))
	}
	c.listChannels(in, args)
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

func (c *Controller) addSong(in FromBot, args string) {
	s := regexp.MustCompile(" +").Split(args, 3)

	channel, err := c.parseChannel(s[0])
	if err != nil {
		c.Tell(in.user, "We have with that channel "+err.Error())
		return
	}

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

	err = c.jukebox.AddSong(song, channel)
	if err != nil {
		c.Tell(in.user, "I had trouble adding "+song.URL+" to "+channel)
		return
	}
	c.Tell(in.user, "I have added "+song.URL+" to "+channel)
}

func (c *Controller) scheduleChannel(in FromBot, args string) {
	s := regexp.MustCompile(" +").Split(args, 2)
	if len(s) != 2 {
		c.Tell(in.user, "please use:  schedule  #channel_name new crontab")
		return
	}
	channel, err := c.parseChannel(s[0])
	if err != nil {
		fmt.Printf("I was unable to parse %s becuase of %s\n", channel, err.Error())
		return
	}
	cron := s[1]

	if len(regexp.MustCompile(" +").Split(cron, -1)) != 5 {
		c.Tell(in.user, "Wrong cron format. Please use  MIN HOUR DAY_OF_MONTH MONTH DAY_OF_WEEK")
		c.Tell(in.user, "Please see the Cron docs at https://github.com/jdblack/sotd for more details")
		return
	}

	_, err = c.jukebox.ScheduleChannel(channel, cron)
	if err != nil {
		c.Tell(in.user, "I tried to change your schedule, but "+err.Error())
	}
	c.Tell(in.user, "Schedule updated for "+channel)

}

func (c *Controller) parseChannel(in_channel string) (string, error) {
	channel, err := c.bot.parseChannel(in_channel)
	if err != nil {
		return channel, err
	}
	present, err := c.bot.InChannel(channel)
	if err != nil {
		return channel, err
	}
	if !present {
		msg := fmt.Sprintf("Not in channel %s ( %s)", in_channel, channel)
		return channel, errors.New(msg)
	}
	return channel, nil
}

// AddSong blah

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
	c.Tell(in.user, "I have been invited to :"+strings.Join(chans, ", "))
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
