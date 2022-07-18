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

	pl, err := jb.GetPlaylist("nochannel")
	require.NotNil(t, err)

	pl, err = jb.GetPlaylist("testchan")
	require.NoError(t, err)
	require.Equal(t, len(pl.Songs), len(songs))
}

func TestLoadURL(t *testing.T) {
	jb := testNewJB()
	songs, err := jb.songsFromJSON("testuser", "testchan", testurl)
	assert.NotNil(t, songs)
	require.NoError(t, err)

	pl, err := jb.GetPlaylist("testchan")
	require.NoError(t, err)
	require.Equal(t, len(pl.Songs), len(songs))
}

func TestDeleteSong(t *testing.T) {
	jb := testNewJB()
	_, err := jb.songsFromJSON("testuser", "testchan", "testing/songs.json")
	require.NoError(t, err)

	pl, err := jb.GetPlaylist("testchan")
	oldlen := len(pl.Songs)
	targetSong := pl.Songs[0]

	jb.DeleteSongByURL(targetSong.URL)
	assert.Equal(t, oldlen, len(pl.Songs)) //Because its stale

	pl, err = jb.GetPlaylist("testchan")
	assert.Equal(t, oldlen-1, len(pl.Songs))

}
