package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rclancey/events"
	"github.com/rclancey/httpserver/v2"
	"github.com/rclancey/sensors/sonos"
)

type Sonos struct {
	cfg *Config
	sonos *sonos.Sonos
	eventSink events.EventSink
	state *sonos.State
	updates chan *sonos.State
}

func NewSonos(cfg *Config, eventSink events.EventSink) (*Sonos, error) {
	updates := make(chan *sonos.State, 100)
	api, err := sonos.NewSonos(cfg.Network, updates)
	if err != nil {
		return nil, err
	}
	s := &Sonos{
		cfg: cfg,
		sonos: api,
		updates: updates,
		eventSink: events.NewPrefixedEventSource("sonos", eventSink),
	}
	go func() {
		for {
			update := <-updates
			if s.state == nil {
				s.state = update
			} else {
				s.state = s.state.ApplyUpdate(update)
			}
			if update.State != nil {
				switch *update.State {
				case sonos.PlayStatePlaying:
					s.eventSink.Emit("play", s.state)
				case sonos.PlayStatePaused:
					s.eventSink.Emit("pause", s.state)
				case sonos.PlayStateStopped:
					s.eventSink.Emit("stop", s.state)
				}
			}
			if update.Speed != nil {
				s.eventSink.Emit("speed", s.state)
			}
			if update.Volume != nil || update.Mute != nil {
				s.eventSink.Emit("volume", s.state)
			}
			if update.PlayMode != nil {
				s.eventSink.Emit("mode", s.state)
			}
			if update.Tracks != nil {
				s.eventSink.Emit("queue", s.state)
			}
			if update.Index != nil {
				s.eventSink.Emit("track", s.state)
			}
		}
	}()
	return s, nil
}

func (s *Sonos) Check() (interface{}, error) {
	state, err := s.sonos.GetState()
	if err != nil {
		return nil, err
	}
	s.sonos.ApplyUpdate(state)
	return state, nil
}

func (s *Sonos) Monitor(interval time.Duration) func() {
	return Monitor(s, interval)
}

func (s *Sonos) HandleRead(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return s.state, nil
}

func (s *Sonos) HandleSetVolume(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	vol, err := strconv.Atoi(string(data))
	if err != nil {
		return nil, err
	}
	if vol < 0 {
		return nil, httpserver.BadRequest
	}
	if vol > 100 {
		return nil, httpserver.BadRequest
	}
	s.sonos.SetVolume(vol)
	return s.Check()
}

func (s *Sonos) HandleSetPlaylist(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	playlistData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	ct := strings.Split(r.Header.Get("Content-Type"), ";")[0]
	pl := &sonos.Playlist{}
	switch ct {
	case "application/json":
		err = json.Unmarshal(playlistData, pl)
	case "audio/x-mpegurl":
		err = pl.UnmarshalM3U(playlistData)
	case "application/x-plist", "application/xml":
		err = pl.UnmarshalPlist(playlistData)
	default:
		return nil, httpserver.BadRequest
	}
	if err != nil {
		return nil, err
	}
	err = s.sonos.ReplaceQueue(pl.Items)
	if err != nil {
		return nil, err
	}
	return s.Check()
}

func (s *Sonos) HandleSetPlayback(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	pos, err := strconv.Atoi(string(data))
	if err != nil {
		return nil, httpserver.BadRequest
	}
	if pos < 0 {
		err = s.sonos.Pause()
	} else if pos == 0 {
		err = s.sonos.Play()
	} else {
		err = s.sonos.SetQueuePos(pos-1)
		if err != nil {
			return nil, err
		}
		err = s.sonos.Play()
	}
	if err != nil {
		return nil, err
	}
	return s.Check()
}

func (s *Sonos) HandlePlayTrack(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	track := &sonos.Track{}
	err := httpserver.ReadJSON(r, track)
	if err != nil {
		return nil, err
	}
	if track.URI == "" {
		return nil, httpserver.BadRequest
	}
	err = s.sonos.ReplaceQueue([]*sonos.Track{track})
	if err != nil {
		return nil, err
	}
	return s.Check()
}
