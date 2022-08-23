package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTitleFetch(t *testing.T) {
	title, err := HTMLTitle("http://google.com")
	require.NoError(t, err)
	assert.Equal(t, "Google", title)
}

func TestPageTitleNXdomain(t *testing.T) {
	title, err := HTMLTitle("http://this-domain-shouldnt-exist-anywhere.com")
	require.Error(t, err)
	assert.Equal(t, "", title)
}
