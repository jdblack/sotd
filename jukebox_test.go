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
	err := jb.songsFromJSON("testuser", "testchan", "testing/songs.json")
	require.NoError(t, err)

	pls := jb.GetPlaylists()
	assert.Equal(t, 1, len(pls))

	pl, err := jb.GetPlaylist("testchan")
	require.NoError(t, err)
	assert.Greater(t, len(pl.Songs), 1)
}

func TestLoadURL(t *testing.T) {
	jb := testNewJB()
	err := jb.songsFromJSON("testuser", "testchan", testurl)
	require.NoError(t, err)

	pls := jb.GetPlaylists()
	require.Equal(t, 1, len(pls))

	pl, err := jb.GetPlaylist("testchan")
	require.NoError(t, err)
	require.NotNil(t, pl)
	require.Greater(t, len(pl.Songs), 1)
}
