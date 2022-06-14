package main

import (
  "errors"
  "fmt"
  "strings"
)

// This takes user input for a song and adds
// it to the jukebox...
// FIXME We need to know the user
// FIXME we need a jukebox too =)
func AddSong(input string) (string, error) {
  data, err := ParseSongInput(input)

  if err != nil {
    return "Error adding "+input, err
  }
  fmt.Println("Adding song")
  // FIXME: add the song to jukebox here
  return fmt.Sprint(data), err
}

// Here we strip off the first atom as the wanted command
// and pack the rest into a string
func Commands(msg string) (string, error){
  fmt.Println("Parsing command " + msg)
  parsed := strings.SplitN(msg," ", 2)
  cmd := parsed[0]
  args := ""

  if len(parsed) == 2 {
    args = parsed[1]
  }

  // FIXME This should be a function table, not a switch
  // FIXME This should be in a controller together bot and jukebox together
  switch {
  case cmd == "add":
    return AddSong(args)
  }
  return "", errors.New("Unknown command " + cmd)
}

// This converts a string of thing=other;this=that into a map
// It feels like there should be a better way to do this.
func ParseSongInput(in string) (map[string]string, error) {
  answer := make(map[string]string)
  atoms := strings.Split(in,";")
  for _,atom := range atoms {
    subs := strings.SplitN(atom,"=", 2)
    if len(subs) < 2 {
      return nil, errors.New("Incorrect song format. Please ask me for help \""+atom+"\"")
    }
    answer[strings.TrimSpace(subs[0])] = strings.TrimSpace(subs[1])
  }

  return answer, nil
}
