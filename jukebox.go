package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-co-op/gocron"
	"gopkg.in/ini.v1"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Jukebox is the main jukebox struct
type Jukebox struct {
	ready   bool
	db      *gorm.DB
	cron    *gocron.Scheduler
	config  *ini.File
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
	Cron  string `gorm:"default:* * * * *"`
	Songs []Song `gorm:"many2many:song_playlist;"`
}

// Song is self descripitive
type Song struct {
	gorm.Model
	URL         string
	Description string
	User        string
	Playlists   []Playlist `gorm:"many2many:song_playlist;"`
}

// Playhistory remembers when songs were played
type Playhistory struct {
	gorm.Model
	Song     Song
	Playlist Playlist
}

func (j *Jukebox) songsFromJSON(user string, channel string, path string) ([]Song, error) {
	var songs []Song
	var body []byte
	var err error

	body, err = LoadFile(path)
	if err != nil {
		return songs, err
	}

	err = json.Unmarshal(body, &songs)
	if err != nil {
		return songs, err
	}
	for _, song := range songs {
		song.User = user
		err = j.AddSong(song, channel)
		if err != nil {
			return songs, err
		}
	}
	return songs, nil
}

//Init Set up the jukebox
func (j *Jukebox) Init(cfg *ini.File) error {
	var err error
	j.config = cfg

	dtype := j.config.Section("database").Key("type").String()

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

	path := j.config.Section("database").Key("path").String()
	j.db, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	j.db.AutoMigrate(&Song{}, &Playlist{})
	if err == nil {
		j.ready = true
	}
	return err
}

func (j *Jukebox) openMySQL() error {
	var err error

	host := j.config.Section("database").Key("host").String()
	port := j.config.Section("database").Key("port").String()
	user := j.config.Section("database").Key("user").String()
	pass := j.config.Section("database").Key("pass").String()
	db := j.config.Section("database").Key("db").String()
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
func NewJukebox(cfg *ini.File) (*Jukebox, error) {
	j := Jukebox{}
	err := j.Init(cfg)
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
	fmt.Printf("Spin a song from  %+s\n", name)
	channel, err := j.ensurePlaylist(name)
	if err != nil {
		return err
	}

	if len(channel.Songs) > 0 {
		song := channel.Songs[rand.Intn(len(channel.Songs))]
		fmt.Printf("I chose %d:%s from %d\n", song.ID, song.URL, len(channel.Songs))
		j.Playset <- Play{channel: name, song: song}

		var play = Playhistory{
			Song:     song,
			Playlist: channel,
		}
		res := j.db.Create(&play)
		if res.Error != nil {
			return res.Error
		}

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
func (j *Jukebox) AddSong(song Song, channel string) error {
	playlist, err := j.ensurePlaylist(channel)
	if err != nil {
		return err
	}
	res := j.db.Create(&song)
	if res.Error != nil {
		return res.Error
	}

	err = j.db.Model(&playlist).Association("Songs").Append(&song)

	return err
}
