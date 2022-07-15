package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelloName(t *testing.T) {
	cfg, _ := loadConfig("testing/test1.ini")
	assert.NotNil(t, cfg)
}
