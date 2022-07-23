package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/ini.v1"
)

func loadConfig(src string) (*ini.File, error) {
	fn, err := homedir.Expand(src)
	if err != nil {
		log.Println("Unable to parse filename " + src)
		os.Exit(1)
	}
	return ini.Load(fn)
}

// ParseStrIntoMap Takes a string in the format of  this=that;this=theother and
// Makes a map from it
func ParseStrIntoMap(in string) (map[string]string, error) {
	answer := make(map[string]string)
	atoms := strings.Split(in, ";")
	for _, atom := range atoms {
		subs := strings.SplitN(atom, "=", 2)
		if len(subs) < 2 {
			return nil, errors.New("Incorrect song format. Please ask me for help \"" + atom + "\"")
		}
		answer[strings.TrimSpace(subs[0])] = strings.TrimSpace(subs[1])
	}

	return answer, nil
}

// LoadFile loads a file from either the filesystem or http
// FIXME There should be an option to limit file access
// to http
func LoadFile(path string) ([]byte, error) {

	var body []byte
	var err error

	if !strings.HasPrefix(strings.ToLower(path), "http") {
		return os.ReadFile(path) /* #nosec G304 Protected by option*/
	}

	resp, err := http.Get(path) /* #nosec G107 Protected by option*/
	if err != nil {
		return body, err
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}
	return body, nil
}

func ParseChannel(channel string) (string, string) {
	cleaned := strings.Trim(channel, "<>")
	id, name, found := strings.Cut(cleaned, "|")
	if !found {
		name = id
		id = ""
	}
	return id, name
}

func ParseURL(url string) string {
	return strings.Trim(url, "<>")
}
