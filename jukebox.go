package main
import(
  "time"
  "gorm.io/driver/sqlite"
  "gorm.io/gorm"

)

type JukeBox struct {
  ready bool
  db *gorm.DB
}

func (j *JukeBox) Init(config map[string]string) (error) {
  db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
  db.AutoMigrate(&Song{}, &Playlist{})
  j.db = db
  if err == nil {
    j.ready = true
  }
  return err
}

type Song struct {
  gorm.Model
  URL string
  Description string
  User string
  PlayedOn time.Time
}

type Playlist struct {
  gorm.Model
  Channel string
  Songs []Song
}

func (j *JukeBox) NewSong(song *Song) (uint, error) {
  res := j.db.Create(song)
  return song.ID, res.Error
}

func (j *JukeBox) Checksong(url string) (bool) {
  song := Song{URL: url}

  result := j.db.Where(song).First(&song)
  return result.RowsAffected == 1
}

func (j *JukeBox) RandomOldSong() (Song) {
  song := Song{}
  j.db.Take(&song)
  return song
}

func (j *JukeBox) PickSong(channel string) (Song) {
  playlist := Playlist{Channel: channel}
  result := j.db.Where(playlist).First(&playlist)
  if result.RowsAffected == 0 {
    return j.RandomOldSong()
  }

  // take from list

  if result.RowsAffected == 0 {
    return j.RandomOldSong()
  }

  // return it

  return Song{}
}

func (j *JukeBox) NewPlaylist(channel string) (Playlist){
  playlist := Playlist{Channel: channel}
  j.db.Where(playlist).FirstOrCreate(&playlist)
  return playlist
}


func (j *JukeBox) AddSong(channel string, song *Song) {
  // Find playlist
  // create if not
  playlist := j.NewPlaylist(channel)
  j.db.Where(playlist).FirstOrCreate(&playlist)

  j.NewSong(song)
  j.db.Model(&playlist).Association("Songs").Append(song)
}

// add! channel url  description
// delete channel url  description
// list channel
// history


