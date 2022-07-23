package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testurl = "https://for-your-perusal.s3.amazonaws.com/sotd/songs_test1.json"

func testNewJB(t *testing.T) *Jukebox {
	err := Cfg.load("testing/test1.ini")
	require.NoError(t, err)
	jb, err := NewJukebox(&Cfg)
	require.NoError(t, err)
	return jb
}

func TestInit(t *testing.T) {
	err := Cfg.load("testing/test1.ini")
	assert.Nil(t, err)
	jukebox, err := NewJukebox(&Cfg)
	require.NoError(t, err)
	assert.NotNil(t, jukebox)
}

func mockFB(user string, channel string, path string) (FromBot, string) {
	args := fmt.Sprintf("%s %s", channel, path)
	return FromBot{user: user, message: args}, args
}

func TestLoadFile(t *testing.T) {
	jb := testNewJB(t)
	songs, err := jb.loadSongs(mockFB("testuser", "testchan", "testing/songs_jblack.json"))
	require.NoError(t, err)
	assert.NotNil(t, songs)

	pls := jb.GetPlaylists()
	assert.Equal(t, 1, len(pls))

	pl, err := jb.GetPlaylist("nochannel")
	require.NotNil(t, err)

	pl, err = jb.GetPlaylist("testchan")
	require.NoError(t, err)
	require.Equal(t, len(pl.Songs), len(songs))
}

func TestLoadURL(t *testing.T) {
	jb := testNewJB(t)
	songs, err := jb.loadSongs(mockFB("testuser", "testchan", testurl))
	assert.NotNil(t, songs)
	require.NoError(t, err)

	pl, err := jb.GetPlaylist("testchan")
	require.NoError(t, err)
	require.Equal(t, len(pl.Songs), len(songs))
}

func TestDeleteSong(t *testing.T) {
	jb := testNewJB(t)
	_, err := jb.loadSongs(mockFB("testuser", "testchan", "testing/songs_jblack.json"))
	require.NoError(t, err)

	pl, err := jb.GetPlaylist("testchan")
	oldlen := len(pl.Songs)
	targetSong := pl.Songs[0]

	jb.DeleteSongByURL(targetSong.URL)
	assert.Equal(t, oldlen, len(pl.Songs)) //Because its stale

	pl, err = jb.GetPlaylist("testchan")
	assert.Equal(t, oldlen-1, len(pl.Songs))

}
