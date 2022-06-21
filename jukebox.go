package main
import(
//  "time"
  "gorm.io/driver/sqlite"
  "gorm.io/gorm"
//  "errors"
)

type JukeBox struct {
  ready bool
  db *gorm.DB
}

func (j *JukeBox) Init(config map[string]string) (error) {
  db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
  db.AutoMigrate(&Song{}, &PlayList{})
  j.db = db
  if err == nil {
    j.ready = true
  }
  return err
}

// PlayList is self explanative
type PlayList struct {
  gorm.Model
  Channel string
  Songs []Song
}

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


func (j *JukeBox) RandomOldSong() (Song) {
  song := Song{}
  j.db.Take(&song)
  return song
}

func (j *JukeBox) PickSong(channel string) (Song) {
  playlist := PlayList{Channel: channel}
  result := j.db.Where(playlist).First(&playlist)
  if result.RowsAffected == 0 {
    return j.RandomOldSong()
  }

  if result.RowsAffected == 0 {
    return j.RandomOldSong()
  }
  return Song{}
}

// Get jj
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


