package sonos

import (
	//"encoding/xml"
	//"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
func TestStruct(t *testing.T) {
	container := &didlContainer{
		Items: []*didlItem{&didlItem{
			ID: "1234",
			ParentID: "5678",
			ObjectClass: "obj.cls",
			Res: didlRes{
				ProtocolInfo: "http",
				Duration: "1:02:34.567",
				URI: "http://google.com/",
			},
			Title: "title",
			Artist: "artist",
			Album: "album",
			Genre: "genre",
			AlbumArtURI: "http://album.art/",
		}},
	}
	data, err := xml.Marshal(container)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(data))
	t.Fail()
}
*/

func TestMarshal(t *testing.T) {
	tr := &Track{
		URI: "http://napster.com/The_Beatles/The_White_Album/While_My_Guitar_Gently_Weeps.mp3",
		Artist: "The Beatles",
		Album: "The White Album",
		Title: "While My Guitar Gently Weeps",
		Genre: "Rock",
		Time: 285.178,
		AlbumArtURI: "http://napster.com/The_Beatles/The_White_Album.jpg",
	}
	data, err := tr.MarshalDIDLLite()
	assert.Nil(t, err)
	exp := `<DIDL-Lite xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:upnp="urn:schemas-upnp-org:metadata-1-0/upnp/" xmlns:r="urn:schemas-rinconnetworks-com:metadata-1-0/" xmlns="urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/"><item id="346ee6a615936d25f9dac3514dd87d5b9803ceab" parentID="346ee6a615936d25f9dac3514dd87d5b9803ceab"><upnp:class>object.item.audioItem.musicTrack</upnp:class><res protocolInfo="http-get:*:audio/mpeg:*" duration="0:04:45.178">http://napster.com/The_Beatles/The_White_Album/While_My_Guitar_Gently_Weeps.mp3</res><dc:title>While My Guitar Gently Weeps</dc:title><dc:creator>The Beatles</dc:creator><upnp:album>The White Album</upnp:album><upnp:genre>Rock</upnp:genre><upnp:albumArtURI>http://napster.com/The_Beatles/The_White_Album.jpg</upnp:albumArtURI></item></DIDL-Lite>`
	assert.Equal(t, exp, string(data))
}

func TestUnmarshal(t *testing.T) {
	tr := &Track{
		URI: "http://napster.com/The_Beatles/The_White_Album/While_My_Guitar_Gently_Weeps.mp3",
		Artist: "The Beatles",
		Album: "The White Album",
		Title: "While My Guitar Gently Weeps",
		Genre: "Rock",
		Time: 285.178,
		AlbumArtURI: "http://napster.com/The_Beatles/The_White_Album.jpg",
	}
	data, err := tr.MarshalDIDLLite()
	assert.Nil(t, err)
	xt := &Track{}
	err = xt.UnmarshalDIDLLite(data)
	assert.Nil(t, err)
	assert.Equal(t, tr.URI, xt.URI)
	assert.Equal(t, tr.Artist, xt.Artist)
	assert.Equal(t, tr.Album, xt.Album)
	assert.Equal(t, tr.Title, xt.Title)
	assert.Equal(t, tr.Genre, xt.Genre)
	assert.Equal(t, tr.Time, xt.Time)
	assert.Equal(t, tr.AlbumArtURI, xt.AlbumArtURI)
}
