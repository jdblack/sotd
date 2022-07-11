package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-co-op/gocron"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Jukebox struct {
	ready   bool
	db      *gorm.DB
	cron    *gocron.Scheduler
	Playset chan Play
}

type Play struct {
	URL         string
	user        string
	channel     string
	chanid      string
	description string
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

// RandomSong picks a song not in the playlist
func (j *Jukebox) RandomSong() Song {
	song := Song{}
	j.db.Take(&song)
	return song
}

// PickSong gets the next song for a playlist
func (j *Jukebox) PickSong(channel string) Song {
	playlist := Playlist{Channel: channel}
	result := j.db.Where(playlist).First(&playlist)
	if result.RowsAffected == 0 {
		return j.RandomSong()
	}

	return Song{}
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

	song := channel.Songs[rand.Intn(len(channel.Songs))]

	fmt.Printf("I have %d  songs to choose from, but I chose %d, %s \n", len(channel.Songs), song.ID, song.URL)

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
