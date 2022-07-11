package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-co-op/gocron"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Jukebox is the main jukebox struct
type Jukebox struct {
	ready   bool
	db      *gorm.DB
	cron    *gocron.Scheduler
	Playset chan Play
}

// Play is a instruction to play a song in a channel
type Play struct {
	channel string
	song    Song
}

// Playlist is self explanativ
type Playlist struct {
	gorm.Model
	Channel string
	//Cron string `gorm:"default:0 18 * * 1-5"`
	Cron       string `gorm:"default:* * * * *"`
	LastPlayed time.Time
	Songs      []Song `gorm:"many2many:song_playlist;"`
}

// Song is self descripitive
type Song struct {
	gorm.Model
	URL         string
	Description string
	User        string
	Playlists   []Playlist `gorm:"many2many:song_playlist;"`
}

// PlayHistory remembers when songs were played
type PlayHistory struct {
	gorm.Model
	Song     Song
	Playlist Playlist
}

//Init Set up the jukebox
func (j *Jukebox) Init() error {
	var err error

	dtype := Config.Section("database").Key("type").String()

	j.Playset = make(chan Play, 100)

	switch {
	case dtype == "sqlite":
		err = j.openSQLite()
	case dtype == "mysql":
		err = j.openMySQL()
	}
	if err != nil {
		return err
	}

	j.cron = gocron.NewScheduler(time.UTC)
	j.schedulePlaylists()
	j.cron.StartAsync()
	return nil
}

//openSQLite Opens up sqlite
func (j *Jukebox) openSQLite() error {
	var err error

	path := Config.Section("database").Key("path").String()
	j.db, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	j.db.AutoMigrate(&Song{}, &Playlist{})
	if err == nil {
		j.ready = true
	}
	return err
}

func (j *Jukebox) openMySQL() error {
	var err error

	host := Config.Section("database").Key("host").String()
	port := Config.Section("database").Key("port").String()
	user := Config.Section("database").Key("user").String()
	pass := Config.Section("database").Key("pass").String()
	db := Config.Section("database").Key("db").String()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, pass, host, port, db)
	dsn += "?charset=utf8mb4&parseTime=True&loc=Local"

	j.db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	j.db.AutoMigrate(&Song{}, &Playlist{})
	if err == nil {
		j.ready = true
	}
	return err
}

// NewJukebox creates a new jukebox
func NewJukebox() (*Jukebox, error) {
	j := Jukebox{}
	err := j.Init()
	return &j, err
}

// GetPlaylist get a playlist
func (j *Jukebox) GetPlaylist(channel string) (Playlist, error) {
	playlist := Playlist{Channel: channel}
	res := j.db.Preload("Songs").Where(playlist).First(&playlist)
	return playlist, res.Error
}

func (j *Jukebox) ensurePlaylist(channel string) (Playlist, error) {
	playlist := Playlist{Channel: channel}
	res := j.db.Preload("Songs").Where(playlist).FirstOrCreate(&playlist)
	return playlist, res.Error
}

//GetPlaylists gets all the playlists
func (j *Jukebox) GetPlaylists() []Playlist {
	var playlists []Playlist
	j.db.Find(&playlists)
	return playlists
}

func (j *Jukebox) schedulePlaylists() {
	var playlists []Playlist
	j.db.Find(&playlists)
	for _, pl := range playlists {
		fmt.Printf("Set up cron schedule for %s with %s\n", pl.Channel, pl.Cron)
		channel := pl.Channel
		j.cron.Cron(pl.Cron).Tag(pl.Channel).Do(func() {
			j.spinPlaylist(channel)
		})
	}
}

func (j *Jukebox) spinPlaylist(name string) error {
	//FIXME We need to add to the playset channel
	//FIXME Remove played song
	fmt.Printf("Spin a song from  %+s\n", name)
	channel, err := j.ensurePlaylist(name)
	if err != nil {
		return err
	}

	if len(channel.Songs) > 0 {
		song := channel.Songs[rand.Intn(len(channel.Songs))]
		fmt.Printf("I chose %d:%s from %d\n", song.ID, song.URL, len(channel.Songs))
		j.Playset <- Play{channel: name, song: song}
		fmt.Printf("Deassociate song from channel %s : %+v\n", name, song)
		err = j.db.Model(&channel).Association("Songs").Delete(&song)
		if err != nil {
			fmt.Printf("I had an errorw with song (%s) %s : %+v\n", err, name, song)
		}
		return nil

	}

	var songs []Song
	j.db.Find(&songs)
	if len(songs) == 0 {
		return errors.New("Unable to find new song to play")
	}

	song := songs[rand.Intn(len(songs))]
	j.Playset <- Play{channel: name, song: song}
	fmt.Printf("%+v\n", song)

	return nil
}

// CreateSong creates a song if it doesnt exist
func (j *Jukebox) CreateSong(songIn map[string]string) (Song, error) {

	song := Song{
		URL:         songIn["url"],
		Description: songIn["description"],
		User:        songIn["user"],
	}

	return song, nil
}

// AddSong creates a song and adds it to a channel
//FIXME We need to send channelid & channel name, not just channel name
func (j *Jukebox) AddSong(song Song, channel string) error {
	playlist, err := j.ensurePlaylist(channel)
	if err != nil {
		return err
	}
	res := j.db.Create(&song)
	if res.Error != nil {
		return res.Error
	}

	j.db.Model(&playlist).Association("Songs").Append(&song)

	return nil
}
