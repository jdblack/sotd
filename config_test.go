package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var TestDir string

func testFile(filepath string) string {
	wd, _ := os.Getwd()
	return wd + "/" + filepath
}

func TestConfigLoad(t *testing.T) {
	fn := testFile("testing/test1.ini")
	cfg, err := loadConfig(fn)
	assert.Nil(t, err)
	assert.Equal(t, "sqlite", cfg.Section("database").Key("type").String())
}
