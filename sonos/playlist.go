package sonos

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var extInfRe = regexp.MustCompile(`EXTINF:\s*(\d+),\s*<(.*?)>\s*<(.*?)>\s*<(.*?)>$`)

type Playlist struct {
	Name string
	Items []*Track
}

func (pl *Playlist) UnmarshalM3U(data []byte) error {
	tracks := []*Track{}
	var curTrack *Track
	lines := strings.Split(string(data), "\n")
	if len(lines) == 1 {
		// itunes exports with <CR> rather than <LF>
		lines = strings.Split(string(data), "\r")
	}
	for _, line := range lines {
		if curTrack == nil {
			curTrack = &Track{}
		}
		if strings.HasPrefix(line, "#EXTINF:") {
			parts := strings.SplitN(strings.TrimPrefix(line, "#EXTINF:"), ",", 2)
			if len(parts) == 2 {
				t, err := strconv.Atoi(strings.TrimSpace(parts[0]))
				if err == nil {
					curTrack.Time = float64(t)
				}
				info := strings.TrimSpace(parts[1])
				if strings.HasPrefix(info, "<") && strings.HasSuffix(info, ">") {
					infoParts := strings.Split(info[1:], "<")
					for i := range infoParts {
						infoParts[i] = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(infoParts[i]), ">"))
					}
					if len(infoParts) == 3 {
						curTrack.Artist = infoParts[0]
						curTrack.Album = infoParts[1]
						curTrack.Title = infoParts[2]
					}
				} else if strings.Contains(info, " - ") {
					// itunes format is title - artist
					infoParts := strings.SplitN(info, " - ", 2)
					if curTrack.Title == "" && curTrack.Artist == "" {
						curTrack.Title = strings.TrimSpace(infoParts[0])
						curTrack.Artist = strings.TrimSpace(infoParts[1])
					}
				} else {
					curTrack.Title = info
				}
			}
		} else if strings.HasPrefix(line, "#EXTALB:") {
			curTrack.Album = strings.TrimSpace(strings.TrimPrefix(line, "#EXTALB:"))
		} else if strings.HasPrefix(line, "#EXTART:") {
			curTrack.Artist = strings.TrimSpace(strings.TrimPrefix(line, "#EXTART:"))
		} else if strings.HasPrefix(line, "#EXTGENRE:") {
			curTrack.Genre = strings.TrimSpace(strings.TrimPrefix(line, "#EXTGENRE:"))
		} else if strings.HasPrefix(line, "#EXTIMG:") {
			curTrack.AlbumArtURI = strings.TrimSpace(strings.TrimPrefix(line, "#EXTIMG:"))
		} else if strings.HasPrefix(line, "http") {
			curTrack.URI = line
			tracks = append(tracks, curTrack)
			curTrack = nil
		}
	}
	pl.Items = tracks
	return nil
}

func (pl *Playlist) MarshalM3U() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("#EXTM3U\n")
	writeTag := func(tag, val string) {
		if val == "" {
			return
		}
		buf.WriteString(fmt.Sprintf("#EXT%s:%s\n", strings.ToUpper(tag), val))
	}
	for _, tr := range pl.Items {
		t := int(math.Round(tr.Time))
		title := tr.Title
		if tr.Artist != "" {
			title += " - " + tr.Artist
		}
		buf.WriteString(fmt.Sprintf("#EXTINF:%d,%s\n", t, title))
		writeTag("alb", tr.Album)
		writeTag("art", tr.Artist)
		writeTag("genre", tr.Genre)
		writeTag("img", tr.AlbumArtURI)
		buf.WriteString(tr.URI+"\n")
	}
	return buf.Bytes(), nil
}

type PListTrack struct {
	TrackID int `plist:"Track ID"`
	Name string
	Artist string
	Album string
	AlbumArtist string `plist:"Album Artist"`
	Genre string
	TotalTime int `plist:"Total Time"`
	Location string
}

type PListTrackRef struct {
	TrackID int `plist:"Track ID"`
}

type PListPlaylist struct {
	Name string
	Items []*PListTrackRef `plist:"Playlist Items"`
}

type PList struct {
	MajorVersion int `plist:"Major Version"`
	Date time.Time
	Tracks map[int]*PListTrack
	Playlists []*PListPlaylist
}

func (pl *Playlist) UnmarshalPlist(data []byte) error {
	plist := &PList{}
	err := UnmarshalPlist(data, plist)
	if err != nil {
		return err
	}
	if len(plist.Playlists) == 0 {
		return errors.New("no playlists")
	}
	if len(plist.Playlists) > 1 {
		return errors.New("multiple playlists")
	}
	pl.Name = plist.Playlists[0].Name
	pl.Items = make([]*Track, len(plist.Playlists[0].Items))
	for i, ref := range plist.Playlists[0].Items {
		tr := plist.Tracks[ref.TrackID]
		if tr == nil {
			return errors.New("missing track")
		}
		pl.Items[i] = &Track{
			URI: tr.Location,
			Title: tr.Name,
			Album: tr.Album,
			Artist: tr.Artist,
			Genre: tr.Genre,
			Time: float64(tr.TotalTime) / 1000.0,
		}
	}
	return nil
}

func (pl *Playlist) MarshalPlist() ([]byte, error) {
	tracks := map[int]map[string]interface{}{}
	seen := map[string]int{}
	items := make([]map[string]int, len(pl.Items))
	nextId := 1
	for i, t := range pl.Items {
		id, ok := seen[t.URI]
		if !ok {
			id = nextId
			nextId += 1
			seen[t.URI] = id
			tracks[id] = map[string]interface{}{
				"Track ID": id,
				"Name": t.Title,
				"Album": t.Album,
				"Artist": t.Artist,
				"Genre": t.Genre,
				"Total Time": int(t.Time * 1000),
				"Location": t.URI,
			}
		}
		items[i] = map[string]int{"Track ID": id}
	}
	playlist := map[string]interface{}{
		"Name": pl.Name,
		"Playlist Items": items,
	}
	plist := map[string]interface{}{
		"Major Version": int(1),
		"Minor Version": int(1),
		"Date": time.Now(),
		"Application Version": "1.0.6.10",
		"Features": int(5),
		"Show Content Ratings": true,
		"Tracks": tracks,
		"Playlists": []interface{}{playlist},
	}
	return MarshalPlist(plist)
}
