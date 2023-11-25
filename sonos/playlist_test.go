package sonos

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testpl = &Playlist{
	Name: "Test",
	Items: []*Track{
		&Track{
			URI: "http://napster.com/The_Beatles/The_White_Album/While_My_Guitar_Gently_Weeps.mp3",
			Title: "While My Guitar Gently Weeps",
			Artist: "The Beatles",
			Album: "The White Album",
			Genre: "Rock",
			Time: 285.178,
		},
		&Track{
			URI: "http://napster.com/Pearl_Jam/Ten/Jeremy.mp3",
			Title: "Jeremy",
			Artist: "Pearl Jam",
			Album: "Ten",
			Genre: "Grunge",
			Time: 319.85,
		},
	},
}

func TestMarshalM3U(t *testing.T) {
	data, err := testpl.MarshalM3U()
	assert.Nil(t, err)
	exp := `#EXTM3U
#EXTINF:285,While My Guitar Gently Weeps - The Beatles
#EXTALB:The White Album
#EXTART:The Beatles
#EXTGENRE:Rock
http://napster.com/The_Beatles/The_White_Album/While_My_Guitar_Gently_Weeps.mp3
#EXTINF:320,Jeremy - Pearl Jam
#EXTALB:Ten
#EXTART:Pearl Jam
#EXTGENRE:Grunge
http://napster.com/Pearl_Jam/Ten/Jeremy.mp3
`
	assert.Equal(t, exp, string(data))
}

func TestUnmarshalM3U(t *testing.T) {
	data, err := testpl.MarshalM3U()
	assert.Nil(t, err)
	pl := &Playlist{}
	err = pl.UnmarshalM3U(data)
	assert.Nil(t, err)
	assert.Equal(t, len(testpl.Items), len(pl.Items))
	for i, tr := range pl.Items {
		assert.Equal(t, testpl.Items[i].URI, tr.URI)
		assert.Equal(t, testpl.Items[i].Title, tr.Title)
		assert.Equal(t, testpl.Items[i].Artist, tr.Artist)
		assert.Equal(t, testpl.Items[i].Album, tr.Album)
		assert.Equal(t, testpl.Items[i].Genre, tr.Genre)
		assert.Equal(t, math.Round(testpl.Items[i].Time), tr.Time)
	}
}

func TestMarshalPlist(t *testing.T) {
	data, err := testpl.MarshalPlist()
	assert.Nil(t, err)
	pl := &Playlist{}
	err = pl.UnmarshalPlist(data)
	assert.Nil(t, err)
	assert.Equal(t, pl.Name, testpl.Name)
	assert.Equal(t, len(testpl.Items), len(pl.Items))
	for i, tr := range pl.Items {
		assert.Equal(t, testpl.Items[i].URI, tr.URI)
		assert.Equal(t, testpl.Items[i].Title, tr.Title)
		assert.Equal(t, testpl.Items[i].Artist, tr.Artist)
		assert.Equal(t, testpl.Items[i].Album, tr.Album)
		assert.Equal(t, testpl.Items[i].Genre, tr.Genre)
		assert.Equal(t, testpl.Items[i].Time, tr.Time)
	}
}
