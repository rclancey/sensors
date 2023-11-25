package sonos

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Track struct {
	URI         string  `json:"uri"`
	Title       string  `json:"title,omitempty"`
	Artist      string  `json:"artist,omitempty"`
	Album       string  `json:"album,omitempty"`
	Genre       string  `json:"genre,omitempty"`
	Time        float64 `json:"time,omitempty"`
	AlbumArtURI string  `json:"cover,omitempty"`
}

func (t *Track) ID() string {
	h := sha1.Sum([]byte(t.URI))
	return hex.EncodeToString(h[:])
}

func (t *Track) TimeString() string {
	duration := "0:00"
	if t.Time != 0 {
		hours := int(t.Time) / 3600
		mins := (int(t.Time) % 3600) / 60
		secs := float64(int(t.Time * 1000) % 60000) / 1000.0
		duration = fmt.Sprintf("%d:%02d:%02.3f", hours, mins, secs)
	}
	return duration
}

type didlRes struct {
	ProtocolInfo string `xml:"protocolInfo,attr"`
	Duration     string `xml:"duration,attr"`
	URI          string `xml:",chardata"`
}

type didlItem struct {
	XMLName      xml.Name `xml:"item"`
	ID           string `xml:"id,attr"`
	ParentID     string `xml:"parentID,attr"`
	ObjectClass  string `xml:"class"`
	Res          didlRes `xml:"res"`
	Title        string `xml:"title"`
	Artist       string `xml:"creator"`
	Album        string `xml:"album"`
	Genre        string `xml:"genre"`
	AlbumArtURI  string `xml:"albumArtURI"`
}

type didlContainer struct {
	XMLName xml.Name    `xml:"DIDL-Lite"`
	Items   []*didlItem `xml:"item"`
}

func (t *Track) UnmarshalDIDLLite(data []byte) error {
	container := &didlContainer{}
	err := xml.Unmarshal(data, container)
	if err != nil {
		return err
	}
	if len(container.Items) == 0 {
		return errors.New("no items")
	}
	if len(container.Items) > 1 {
		return errors.New("multiple items")
	}
	if container.Items[0].Res.Duration != "" {
		durParts := strings.Split(container.Items[0].Res.Duration, ":")
		n := len(durParts)
		seconds, err := strconv.ParseFloat(durParts[n-1], 64)
		if err != nil {
			return err
		}
		var minutes int
		if n > 1 {
			minutes, err = strconv.Atoi(durParts[n-2])
			if err != nil {
				return err
			}
		}
		var hours int
		if n > 2 {
			hours, err = strconv.Atoi(durParts[n-3])
			if err != nil {
				return err
			}
		}
		t.Time = float64(hours * 3600) + float64(minutes * 60) + seconds
	}
	t.URI = string(container.Items[0].Res.URI)
	t.Title = container.Items[0].Title
	t.Artist = container.Items[0].Artist
	t.Album = container.Items[0].Album
	t.Genre = container.Items[0].Genre
	t.AlbumArtURI = container.Items[0].AlbumArtURI
	return nil
}

func (t *Track) MarshalDIDLLite() ([]byte, error) {
	id := t.ID()
	buf := bytes.NewBuffer(nil)
	buf.WriteString(`<DIDL-Lite`)
	buf.WriteString(` xmlns:dc="http://purl.org/dc/elements/1.1/"`)
	buf.WriteString(` xmlns:upnp="urn:schemas-upnp-org:metadata-1-0/upnp/"`)
	buf.WriteString(` xmlns:r="urn:schemas-rinconnetworks-com:metadata-1-0/"`)
	buf.WriteString(` xmlns="urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/">`)
	buf.WriteString(fmt.Sprintf(`<item id="%s" parentID="%s">`, id, id))
	buf.WriteString(`<upnp:class>object.item.audioItem.musicTrack</upnp:class>`)
	buf.WriteString(`<res protocolInfo="http-get:*:audio/mpeg:*"`)
	if t.Time > 0 {
		buf.WriteString(fmt.Sprintf(` duration="%s"`, t.TimeString()))
	}
	buf.WriteString(">")
	xml.Escape(buf, []byte(t.URI))
	buf.WriteString("</res>")
	writePart := func(tag, part string) {
		if part == "" {
			return
		}
		buf.WriteString("<"+tag+">")
		xml.Escape(buf, []byte(part))
		buf.WriteString("</"+tag+">")
	}
	writePart("dc:title", t.Title)
	writePart("dc:creator", t.Artist)
	writePart("upnp:album", t.Album)
	writePart("upnp:genre", t.Genre)
	writePart("upnp:albumArtURI", t.AlbumArtURI)
	buf.WriteString(`</item></DIDL-Lite>`)
	return buf.Bytes(), nil
}
