package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
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

// ParseChannel splits the slack channel string into the channel id and channel name
func ParseChannel(channel string) (string, string) {
	cleaned := strings.Trim(channel, "<>")
	id, name, found := strings.Cut(cleaned, "|")
	if !found {
		name = id
	}
	return id, name
}

// ParseURL removes the <> from slack urls
func ParseURL(url string) string {
	return strings.Trim(url, "<>")
}

// URLCache looks up urls, with caching
func URLCache() (f func(string) (string, error)) {
	cache := make(map[string]string)
	f = func(url string) (string, error) {
		if val, ok := cache[url]; ok {
			return val, nil
		}
		title, err := HTMLTitle(url)
		if err == nil {
			cache[url] = title
		}
		return title, err
	}
	return f
}

// HTMLTitle returns the title for a page
func HTMLTitle(page string) (string, error) {
	url := ParseURL(page)
	resp, err := http.Get(url)

	if err != nil {
		fmt.Printf("I got an error: -%s- %s\n", url, err)
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("failed to fetch data: %d %s", resp.StatusCode, resp.Status)
		return msg, errors.New(msg)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		log.Fatal(err)
	}
	title := doc.Find("title").Text()
	return title, nil
}
