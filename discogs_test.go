package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscogs(t *testing.T) {
	// assert.Equal(t, discogsReq("", "GET", nil).StatusCode, 200)
	assert.Equal(t, discogsReq("/releases/4319735", "GET", nil).StatusCode, 200)

	bad := discogsSearch("fjadskjfdslk", "fjdaksjfsdlk")
	assert.Equal(t, bad.Primary().Id, 0)

	sr := discogsSearch("Metallica", "Ride the Lightning")
	pri := sr.Primary()
	// fmt.Printf("%+v", pri)
	assert.Equal(t, pri.Id, 6440)
	assert.Equal(t, pri.Primary, 377464)
	assert.Equal(t, pri.Artists[0].Name, "Metallica")

	nat := discogsSearch("natsumen", "kill your winter")
	assert.Equal(t, nat.Primary().Id, 12578164)

	met := discogsSearchArtist("Metallica")[0]
	assert.Equal(t, met.Id, 18839)

	met1st := met.Releases()[0]
	assert.Equal(t, met1st.Id, 7430321)
	assert.Equal(t, met1st.Title, "Live Metal Up Your Ass / No Life 'Til Leather")
	assert.Equal(t, met1st.Artist, "Metallica")
	assert.Equal(t, met1st.Artists, []Artist(nil)) // no such field
	// assert.Equal(t, met1st.Artists, []Artist{}) // field exists, but empty
}
