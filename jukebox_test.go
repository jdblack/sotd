package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	assert.NotNil(t, jukebox)
}

func TestLoadFile(t *testing.T) {
	jb := testNewJB()
	songs, err := jb.songsFromJSON("testuser", "testchan", "testing/songs.json")
	assert.NotNil(t, songs)
	require.NoError(t, err)

	pls := jb.GetPlaylists()
	assert.Equal(t, 1, len(pls))

	pl, err := jb.GetPlaylist("testchan")
	require.NoError(t, err)
	require.Equal(t, len(pl.Songs), len(songs))
}

func TestLoadURL(t *testing.T) {
	jb := testNewJB()
	songs, err := jb.songsFromJSON("testuser", "testchan", testurl)
	assert.NotNil(t, songs)
	require.NoError(t, err)

	pls := jb.GetPlaylists()
	require.Equal(t, 1, len(pls))

	pl, err := jb.GetPlaylist("testchan")
	require.NoError(t, err)
	require.NotNil(t, pl)
	require.Equal(t, len(pl.Songs), len(songs))
}
