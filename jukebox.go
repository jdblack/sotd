package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
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
	config  *Config
	Playset chan Play
}

// Play is a instruction to play a song in a channel
type Play struct {
	channel  string
	backfill bool
	song     Song
}

// Playlist is the channel representation of a playlist
// FIXME Need a way to change the cron for a playlist
type Playlist struct {
	gorm.Model
	Channel  string `gorm:"unique"`
	Cron     string `gorm:"default:0 18 * * 1-5"`
	Songs    []Song `gorm:"many2many:song_playlist;"`
	Playlogs []Playlog
}

// Playlog remembers when songs were played
// FIXME We need a way to recite this history
type Playlog struct {
	gorm.Model
	SongID     uint
	PlaylistID uint
	backfill   bool
}

// Song is self descripitive
type Song struct {
	gorm.Model
	URL         string `gorm:"unique"`
	Description string
	User        string
	RealName    string
	Playlists   []Playlist `gorm:"many2many:song_playlist;"`
	Playlogs    []Playlog
}

func (j *Jukebox) loadSongs(in FromBot, args string) ([]Song, error) {
	// FIXME Need a way to safely override Real Name
	var songs []Song
	var loaded []Song
	var body []byte
	var err error
	first, second, found := strings.Cut(args, " ")
	if !found {
		return songs, errors.New("Please give me channel and path")
	}
	_, channel := ParseChannel(first)
	path := ParseURL(second)

	body, err = LoadFile(path)
	if err != nil {
		return songs, err
	}

	err = json.Unmarshal(body, &songs)
	if err != nil {
		return songs, err
	}
	for _, song := range songs {
		song.User = in.user
		song.RealName = in.fullName
		err = j.AddSong(song, channel)
		if err != nil {
			return loaded, err
		}
		loaded = append(loaded, song)
	}
	return loaded, nil
}

//Init Set up the jukebox
func (j *Jukebox) Init(cfg *Config) error {
	var err error
	j.config = cfg

	dtype, err := j.config.GetStr("database_type")

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

	path, err := j.config.GetStr("database_path")
	if err != nil {
		return err
	}
	j.db, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return err
	}
	err = j.db.AutoMigrate(&Song{}, &Playlist{}, &Playlog{})
	if err == nil {
		j.ready = true
	}
	return err
}

func (j *Jukebox) openMySQL() error {
	var err error

	host, err := j.config.GetStr("database_host")
	if err != nil {
		return err
	}
	port, err := j.config.GetStr("database_port")
	if err != nil {
		return err
	}
	user, err := j.config.GetStr("database_user")
	if err != nil {
		return err
	}
	pass, err := j.config.GetStr("database_pass")
	if err != nil {
		return err
	}
	db, err := j.config.GetStr("database_db")
	if err != nil {
		return err
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, pass, host, port, db)
	dsn += "?charset=utf8mb4&parseTime=True&loc=Local"

	j.db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	err = j.db.AutoMigrate(&Song{}, &Playlist{})
	if err == nil {
		j.ready = true
	}
	return err
}

// NewJukebox creates a new jukebox
func NewJukebox(cfg *Config) (*Jukebox, error) {
	j := Jukebox{}
	err := j.Init(cfg)
	return &j, err
}

// GetPlaylist get a playlist
func (j *Jukebox) GetPlaylist(channel string) (Playlist, error) {
	playlist := Playlist{Channel: channel}
	res := j.db.Preload("Songs").Where(playlist).Find(&playlist)
	if res.RowsAffected == 0 {
		return Playlist{}, errors.New("Unable to find playlist")
	}
	return playlist, res.Error
}

//ChannelHistory is supposed to list the songs that have previously
// played in the channel  FIXME It does not
func (j *Jukebox) ChannelHistory(channel string) error {
	playlist := Playlist{Channel: channel}

	res := j.db.Preload("Playlogs").Find(&playlist)
	fmt.Printf("============\n")
	for _, pl := range playlist.Playlogs {
		fmt.Printf("%+v\n", pl)
	}
	fmt.Printf("============\n")
	return res.Error
}

func (j *Jukebox) ensurePlaylist(channel string) (Playlist, error) {
	playlist := Playlist{Channel: channel}
	res := j.db.Find(&playlist)
	found := res.RowsAffected > 0

	res = j.db.Preload("Songs").Where(playlist).FirstOrCreate(&playlist)
	if !found {
		j.schedulePlaylists()
	}
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
	j.cron.Clear()
	j.db.Find(&playlists)

	for _, pl := range playlists {
		channel := pl.Channel
		_, err := j.cron.Cron(pl.Cron).Tag(pl.Channel).Do(func() {
			err := j.spinPlaylist(channel)
			if err != nil {
				fmt.Printf("Got spin error :" + err.Error())
			}
		})
		if err != nil {
			fmt.Printf("Error scheduling playlist %s: %s", pl.Channel, err.Error())
		}
	}
}

func (j *Jukebox) randInt(max int) int64 {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		panic(err)
	}
	return nBig.Int64()

}

func (j *Jukebox) randomSong() (Song, error) {
	var songs []Song
	j.db.Find(&songs)
	fmt.Println("Spin a random song")
	if len(songs) == 0 {
		return Song{}, errors.New("Unable to find new song to play")
	}
	return songs[j.randInt(len(songs))], nil
}

func (j *Jukebox) spinPlaylist(name string) error {
	var err error
	channel, err := j.ensurePlaylist(name)
	play := Play{channel: name, backfill: false}
	fmt.Println("Spin a song for " + name)

	if err != nil {
		return err
	}

	if len(channel.Songs) == 0 {
		play.backfill = true
		play.song, err = j.randomSong()
	} else {
		play.song = channel.Songs[j.randInt(len(channel.Songs))]
		err = j.db.Model(&channel).Association("Songs").Delete(&play.song)
	}
	if err != nil {
		return err
	}

	pl, err := j.GetPlaylist(name)
	if err != nil {
		return err
	}

	fmt.Printf("Adding play for %+v\n", play)
	j.Playset <- play

	fmt.Printf("Storing record\n")

	log := Playlog{SongID: play.song.ID, PlaylistID: pl.ID, backfill: play.backfill}
	res := j.db.Create(&log)
	return res.Error
}

//DeleteChannel removes a channel
func (j *Jukebox) DeleteChannel(in FromBot, channel string) (bool, error) {
	var pl Playlist
	res := j.db.Where("Channel LIKE ?", channel).Delete(&pl)
	j.schedulePlaylists()
	return (res.RowsAffected == 1), res.Error
}

// DeleteSongByURL finds and removes all songs that have this url
func (j *Jukebox) DeleteSongByURL(url string) (int64, error) {
	var songs []Song
	res := j.db.Where("URL LIKE ?", url).Delete(&songs)
	return res.RowsAffected, res.Error
}

func (j *Jukebox) ScheduleChannel(channel string, cron string) (bool, error) {
	var pl Playlist
	created := false

	res := j.db.Where("Channel LIKE ?", channel).FirstOrCreate(&pl)
	fmt.Printf("============")
	fmt.Printf("%+v", res)
	fmt.Printf("============")
	if res.Error != nil {
		return created, res.Error
	}
	res = j.db.Model(&pl).Update("Cron", cron)
	return created, res.Error
}

// AddSong creates a song and adds it to a channel
func (j *Jukebox) AddSong(song Song, channel string) error {
	playlist, err := j.ensurePlaylist(channel)
	if err != nil {
		return err
	}
	res := j.db.FirstOrCreate(&song, &song)
	if res.Error != nil {
		return res.Error
	}

	err = j.db.Model(&playlist).Association("Songs").Append(&song)

	return err
}
