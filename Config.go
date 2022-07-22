package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/ini.v1"
)

type Config struct {
	ini *ini.File
}

// GetInt  fetch config as a strong
func (c *Config) GetInt(name string) int {
	r, _ := strconv.Atoi(c.GetStr(name))
	return r
}

// GetBool  fetch config as a strong
func (c *Config) GetBool(name string) bool {
	r, _ := strconv.ParseBool(c.GetStr(name))
	return r
}

// GetStr fetch config as a strong
func (c *Config) GetStr(name string) string {
	var res string

	res = os.Getenv("SOTD_" + name)
	if len(res) > 0 {
		return res
	}
	if c.ini == nil {
		return ""
	}
	r := c.getIni(name)
	return r
}

func (c *Config) getIni(name string) string {

	section, key, found := strings.Cut(name, "_")
	if found != true {
		key = section
		section = ""
	}
	return c.ini.Section(section).Key(key).String()
}

func (c *Config) load(src string) error {
	fn, err := homedir.Expand(src)
	if err != nil {
		log.Println("Unable to parse filename " + src)
		os.Exit(1)
	}
	ini, err := ini.Load(fn)
	if err == nil {
		c.ini = ini
	}
	return err
}
