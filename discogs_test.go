package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDescend(t *testing.T) {
	// dirs, _ := descend(config.Library.Root, false)
	// assert.Equal(t, dirs[0], "!T.O.O.H.!")

	// assert.Equal(t, discogsGet("").StatusCode, 200) // noop
	// assert.Equal(t, discogsGet("/releases/4319735").StatusCode, 200)

	// assert.Equal(t, discogsReq("", "GET", nil).StatusCode, 200)
	// assert.Equal(t, discogsReq("/releases/4319735", "GET", nil).StatusCode, 200)

	sr := discogsSearch("Metallica", "Ride the Lightning")
	firstMaster := sr.Master()
	assert.Equal(t, firstMaster.id, 6440)
	// pri := firstMaster.Primary()
	// assert.Equal(t, pri.id, 377464)
}
