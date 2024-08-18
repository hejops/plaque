package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscogs(t *testing.T) {
	// dirs, _ := descend(config.Library.Root, false)
	// assert.Equal(t, dirs[0], "!T.O.O.H.!")

	// assert.Equal(t, discogsReq("", "GET", nil).StatusCode, 200)
	assert.Equal(t, discogsReq("/releases/4319735", "GET", nil).StatusCode, 200)

	sr := discogsSearch("Metallica", "Ride the Lightning")
	// assert.Equal(t, firstMaster.Id, 6440)
	assert.Equal(t, sr.Primary(), 377464)
}
