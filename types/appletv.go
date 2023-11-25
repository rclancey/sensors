package types

import (
	"time"
)

type AppleTVStatus struct {
	Result            string    `json:"result"`
	Hash              string    `json:"hash"`
	MediaType         string    `json:"media_type"`
	DeviceState       string    `json:"device_state"`
	Title             *string   `json:"title,omitempty"`
	Artist            *string   `json:"artist,omitempty"`
	Album             *string   `json:"album,omitempty"`
	Genre             *string   `json:"genre,omitempty"`
	TotalTime         *int      `json:"total_time,omitempty"`
	Position          *int      `json:"position,omitempty"`
	Shuffle           string    `json:"shuffle,omitempty"`
	Repeat            string    `json:"repeat,omitempty"`
	SeriesName        *string   `json:"series_name,omitempty"`
	SeasonNumber      *int      `json:"season_number,omitempty"`
	EpisodeNumber     *int      `json:"episode_number,omitempty"`
	ContentIdentifier *string   `json:"content_identifier,omitempty"`
	App               *string   `json:"app,omitempty"`
	AppID             *string   `json:"app_id,omitempty"`
	PowerState        *string   `json:"power_state,omitempty"`
	PushUpdates       *string   `json:"push_updates,omitempty"`
	Connection        *string   `json:"connection,omitempty"`
	LastUpdate        time.Time `json:"now,omitempty"`
}

func (state *AppleTVStatus) Clone() *AppleTVStatus {
	clone := *state
	return &clone
}

func (state *AppleTVStatus) ContentChanged(prev *AppleTVStatus) bool {
	if state.MediaType != prev.MediaType {
		return true
	}
	if !nullStringEqual(state.Title, prev.Title) {
		return true
	}
	if !nullStringEqual(state.Artist, prev.Artist) {
		return true
	}
	if !nullStringEqual(state.Album, prev.Album) {
		return true
	}
	if !nullStringEqual(state.Genre, prev.Genre) {
		return true
	}
	if !nullStringEqual(state.SeriesName, prev.SeriesName) {
		return true
	}
	if !nullStringEqual(state.ContentIdentifier, prev.ContentIdentifier) {
		return true
	}
	if !nullStringEqual(state.App, prev.App) {
		return true
	}
	if !nullStringEqual(state.AppID, prev.AppID) {
		return true
	}
	if !nullIntEqual(state.TotalTime, prev.TotalTime) {
		return true
	}
	if !nullIntEqual(state.SeasonNumber, prev.SeasonNumber) {
		return true
	}
	if !nullIntEqual(state.EpisodeNumber, prev.EpisodeNumber) {
		return true
	}
	return false
}

func nullStringEqual(a, b *string) bool {
	if a == nil {
		return b == nil
	}
	if b == nil {
		return false
	}
	return *a == *b
}

func nullIntEqual(a, b *int) bool {
	if a == nil {
		return b == nil
	}
	if b == nil {
		return false
	}
	return *a == *b
}


/*
{"result": "success",
 "datetime": "2022-09-24T16:06:56.540054-07:00",
 "hash": "ca496c14642c78af6dd4250191fe175f6dafd72b4c33bcbab43c454aae051da1",
 "media_type": "unknown",
 "device_state": "idle",
 "title": null,
 "artist": null,
 "album": null,
 "genre": null,
 "total_time": null,
 "position": null,
 "shuffle": "off",
 "repeat": "off",
 "series_name": null,
 "season_number": null,
 "episode_number": null,
 "content_identifier": null,
 "app": null,
 "app_id": null}
{"result": "success",
 "datetime": "2022-09-24T16:07:06.035189-07:00",
 "hash": "c313595b-f220-4821-a569-c6cb0eaf0744",
 "media_type": "video",
 "device_state": "playing",
 "title": "That Would Be Me",
 "artist": null,
 "album": null,
 "genre": null,
 "total_time": 2265,
 "position": 15,
 "shuffle": "off",
 "repeat": "off",
 "series_name": null,
 "season_number": null,
 "episode_number": null,
 "content_identifier": "c313595b-f220-4821-a569-c6cb0eaf0744",
 "app": "Disney+",
 "app_id": "com.disney.disneyplus"}
{"result": "success",
 "datetime": "2022-09-24T16:07:10.348736-07:00",
 "hash": "c313595b-f220-4821-a569-c6cb0eaf0744",
 "media_type": "video",
 "device_state": "paused",
 "title": "That Would Be Me",
 "artist": null,
 "album": null,
 "genre": null,
 "total_time": 2265,
 "position": 17,
 "shuffle": "off",
 "repeat": "off",
 "series_name": null,
 "season_number": null,
 "episode_number": null,
 "content_identifier": "c313595b-f220-4821-a569-c6cb0eaf0744",
 "app": "Disney+",
 "app_id": "com.disney.disneyplus"}
*/
