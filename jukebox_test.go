package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testurl = "https://for-your-perusal.s3.amazonaws.com/sotd/songs_test1.json"

func testNewJB() *Jukebox {
	cfg, _ := loadConfig("testing/test1.ini")
	jb, _ := NewJukebox(cfg)
	return jb
}

func TestInit(t *testing.T) {
	cfg, err := loadConfig("testing/test1.ini")
	assert.Nil(t, err)
	jukebox, err := NewJukebox(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, jukebox)
}

func TestLoadSongs(t *testing.T) {
	jb := testNewJB()
	err := jb.songsFromJSON("testuser", "testchan", testurl)
	assert.NoError(t, err)

	pls := jb.GetPlaylists()
	assert.Equal(t, 1, len(pls))

	pl, err := jb.GetPlaylist("testchan")
	assert.NoError(t, err)
	assert.NotNil(t, pl)
	assert.Greater(t, len(pl.Songs), 1)
}
