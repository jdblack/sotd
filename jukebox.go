package main
import(
  "gorm.io/driver/sqlite"
  "gopkg.in/ini.v1"
  "gorm.io/gorm"
)

// JukeBox main data object
type JukeBox struct {
  ready bool
  db *gorm.DB
  config *ini.File
}

//OpenSqlite Opens up sqlite
func (j *JukeBox) OpenSqlite() (error) {
  var err error

  path := j.config.Section("database").Key("path").String()
  j.db, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
  j.db.AutoMigrate(&Song{}, &PlayList{})
  if err == nil {
    j.ready = true
  }
  return err
}

//Init Set up the jukebox
func (j *JukeBox) Init(config *ini.File) (error) {
  var err error
  j.config = config

  switch {
  case j.config.Section("database").Key("type").String() == "sqlite":
    err =  j.OpenSqlite()
  }
  return err
}

// PlayList is self explanative
type PlayList struct {
  gorm.Model
  Channel string
  Songs []Song
}

// PlayHistory remembers when songs were played
type PlayHistory struct {
  gorm.Model
  SongID int
  Song Song
  PlayListID int
  PlayList PlayList
}

// Song is self descripitive
type Song struct {
  gorm.Model
  URL string
  Description string
  User string
}


// RandomSong picks a song not in the playlist
func (j *JukeBox) RandomSong() (Song) {
  song := Song{}
  j.db.Take(&song)
  return song
}

// PickSong gets the next song for a playlist
func (j *JukeBox) PickSong(channel string) (Song) {
  playlist := PlayList{Channel: channel}
  result := j.db.Where(playlist).First(&playlist)
  if result.RowsAffected == 0 {
    return j.RandomSong()
  }

  return Song{}
}

// GetPlayList get a playlist 
func (j *JukeBox) GetPlayList(channel string) (PlayList, error) {
  playlist := PlayList{Channel: channel}
  res := j.db.Where(playlist).FirstOrCreate(&playlist)
  if res.Error != nil {
    return playlist, res.Error
  }
  return playlist, nil
}

// CreateSong creates a song if it doesnt exist
func (j *JukeBox) CreateSong(songIn map[string]string) (Song, error)  {

  song := Song{ 
    URL: songIn["url"],
    Description: songIn["description"],
    User: songIn["user"],
  }

  res := j.db.Create(song)
  if res.Error != nil {
    return song, res.Error
  }
  return song, nil
}

// AddSong creates a song and adds it to a channel
func (j *JukeBox) AddSong(songIn map[string]string, channel string) (error){
  playlist, err := j.GetPlayList(channel)
  if err != nil  {
    return err
  }
  song, err := j.CreateSong(songIn)
  if err != nil {
    return err
  }

  j.db.Model(&playlist).Association("Songs").Append(song)

  return nil
}


