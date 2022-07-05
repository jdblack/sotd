package main

// Playlist is self explanative
type Playlist struct {
  gorm.Model
  Channel string
  Cron string `gorm:"default:0 18 * * 1-5"`
  LastPlayed time.Time
  Songs []Song `gorm:"many2many:song_languages;"`
}

// Song is self descripitive
type Song struct {
  gorm.Model
  URL string
  Description string
  User string
  Playlists []Playlist `gorm:"many2many:song_languages;"`
}


// PlayHistory remembers when songs were played
type PlayHistory struct {
  gorm.Model
  Song Song
  Playlist Playlist
}

