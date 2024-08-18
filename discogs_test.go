package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscogs(t *testing.T) {
	// assert.Equal(t, discogsReq("", "GET", nil).StatusCode, 200)
	assert.Equal(t, discogsReq("/releases/4319735", "GET", nil).StatusCode, 200)

	sr := discogsSearch("Metallica", "Ride the Lightning")
	// assert.Equal(t, firstMaster.Id, 6440)
	assert.Equal(t, sr.Primary(), 377464)

	bad := discogsSearch("fjadskjfdslk", "fjdaksjfsdlk")
	assert.Equal(t, bad.Primary(), 0)

	met := discogsSearchArtist("Metallica")[0]
	assert.Equal(t, met.Id, 18839)

	met1st := met.Releases()[0]
	assert.Equal(t, met1st.Id, 7430321)
	assert.Equal(t, met1st.Title, "Live Metal Up Your Ass / No Life 'Til Leather")
}
