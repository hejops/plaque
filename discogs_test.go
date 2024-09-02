package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// note: tests are run in order of filename

func TestDiscogsSearch(t *testing.T) {
	// r := discogsReq("/releases/14056043", "GET", nil)
	// debugResponse(r)

	// database/search?artist=++Tokyo+Jihen++&release_title=
	// database/search?artist=++Tokyo+Jihen++&release_title=遭難

	// r := discogsReq("/releases/9544961", "GET", nil)
	// debugResponse(r)

	// assert.Equal(t, discogsReq("", "GET", nil).StatusCode, 200)
	assert.Equal(t, discogsReq("/releases/4319735", "GET", nil).StatusCode, 200)

	noResults := discogsSearch("Pyrrhic Salvation", "Demo")
	assert.Equal(t, noResults.Primary().Id, 0) // TODO: Primary() should return nil

	rtl := discogsSearch("Metallica", "Ride the Lightning")
	pri := rtl.Primary()
	assert.Equal(t, pri.Id, 6440)
	assert.Equal(t, pri.Primary, 377464)
	assert.Equal(t, pri.Artists[0].Name, "Metallica")

	// no master
	kyw := discogsSearch("natsumen", "kill your winter")
	assert.Equal(t, kyw.Primary().Id, 12578164)
}

func TestDiscogsSearchArtist(t *testing.T) {
	pyr := discogsSearchArtist("Pyrrhic Salvation")
	assert.Len(t, pyr, 1)

	graal := discogsSearchArtist("Graal")[0]
	// assert.Equal(t, graal.UserData, nil)
	assert.Len(t, graal.Releases(), 95)

	met := discogsSearchArtist("Metallica")[0]
	assert.Equal(t, met.Id, 18839)

	met1st := met.Releases()[0]
	assert.Equal(t, met1st.Id, 7430321)
	assert.Equal(t, met1st.Title, "Live Metal Up Your Ass / No Life 'Til Leather")
	assert.Equal(t, met1st.Artist, "Metallica")
	assert.Equal(t, met1st.Artists, []Artist(nil)) // no such field
	// assert.Equal(t, met1st.Artists, []Artist{}) // field exists, but empty

	// artist releases
	rv := discogsSearchArtist("red velvet")[0].Releases()

	assert.Equal(t, rv[0].Artist, "Red Velvet (3)")
	assert.Equal(t, rv[0].ArtistsSort, "")
	assert.Equal(t, rv[0].ReleaseType, "master")
	assert.Equal(t, rv[5].Artist, "Red Velvet (3)")
	assert.Equal(t, rv[5].ReleaseType, "release")
	assert.Len(t, rv[0].Artists, 0)
	assert.Len(t, rv[5].Artists, 0)

	assert.Equal(t, discogsSearchArtist("lil peep")[0].Releases()[0].Id, 11270776)

	assert.Equal(t, discogsSearchArtist("red velvet")[0].Title, "Red Velvet (3)")
}
