package main

import (
	"fmt"
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
func (c *Config) GetInt(name string) (int, error) {
	val, err := c.GetStr(name)
	if err != nil {
		return 0, err
	}
	r, err := strconv.Atoi(val)
	return r, err
}

// GetBool  fetch config as a strong
func (c *Config) GetBool(name string) (bool, error) {
	val, err := c.GetStr(name)
	if err != nil {
		return false, err
	}
	r, err := strconv.ParseBool(val)
	return r, err
}

// GetStr fetch config as a strong
func (c *Config) GetStr(name string) (string, error) {
	var res string

	res = os.Getenv("SOTD_" + name)
	if len(res) > 0 {
		return res, nil
	}
	if c.ini == nil {
		return "", fmt.Errorf("Config option %s not found", name)
	}
	return c.getIni(name)
}

func (c *Config) getIni(name string) (string, error) {

	section, key, found := strings.Cut(name, "_")
	if found != true {
		key = section
		section = ""
	}
	res, err := c.ini.Section(section).GetKey(key)
	if err != nil {
		return "", err
	}
	return res.String(), nil
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
