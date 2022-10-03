package api

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os/exec"
	//"path/filepath"
	"strings"
	"time"

	"github.com/rclancey/httpserver/v2"
)

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

type AppleTVStatus struct {
	Result            string  `json:"result"`
	Hash              string  `json:"hash"`
	MediaType         string  `json:"media_type"`
	DeviceState       string  `json:"device_state"`
	Title             *string `json:"title,omitempty"`
	Artist            *string `json:"artist,omitempty"`
	Album             *string `json:"album,omitempty"`
	Genre             *string `json:"genre,omitempty"`
	TotalTime         *int    `json:"total_time,omitempty"`
	Position          *int    `json:"position,omitempty"`
	Shuffle           string  `json:"shuffle,omitempty"`
	Repeat            string  `json:"repeat,omitempty"`
	SeriesName        *string `json:"series_name,omitempty"`
	SeasonNumber      *int    `json:"season_number,omitempty"`
	EpisodeNumber     *int    `json:"episode_number,omitempty"`
	ContentIdentifier *string `json:"content_identifier,omitempty"`
	App               *string `json:"app,omitempty"`
	AppID             *string `json:"app_id,omitempty"`
	PowerState        *string `json:"power_state,omitempty"`
	PushUpdates       *string `json:"push_updates,omitempty"`
	Connection        *string `json:"connection,omitempty"`
	LastUpdate        time.Time `json:"now,omitempty"`
}

func (state *AppleTVStatus) Clone() *AppleTVStattusl {
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

type AppleTV struct {
	cfg *Config
	eventSink events.EventSink
	cmdIn io.WriteCloser
	state *AppleTVStatus
	updates chan *AppleTVStatus
	cmd *exec.Cmd
}

func stringp(s string) *string {
	return &s
}

func NewAppleTV(cfg *Config, eventSink events.EventSink) (*AppleTV, error) {
	sink := events.NewPrefixedEventSink("appletv", eventSink)
	atv := &AppleTV{
		cfg: cfg,
		eventSink: sink,
		state: &AppleTVStatus{
			DeviceState: "idle",
			PowerState: stringp("off"),
			Connection: stringp("closed"),
			LastUpdate: time.Now(),
		},
		updates: make(chan *AppleTVStatus, 100),
	}
	atv.registerEventTypes()
	go atv.Connect()
	return atv, nil
}

func (atv *AppleTV) registerEventTypes() {
	now := time.Now().In(time.Local)
	atv.eventSink.RegisterEventType("power", &AppleTVStatus{
		PowerState: stringp("on"),
		LastUpdate: now,
	})
	atv.eventSink.RegisterEventType("connection", &AppleTVStatus{
		Connection: stringp("open"),
		LastUpdate: now,
	})
	atv.eventSink.RegisterEventType("content", &AppleTVStatus{
		Result: "success",
		MediaType: "video",
		DeviceState: "playing",
		Title: stringp("Stranger Things 4"),
		TotalTime: intp(3746),
		Position: intp(0),
		App: stringp("Netflix"),
		LastUpdate: now,
	})
	atv.eventSink.RegisterEventType("idle", &AppleTVStatus{
		Result: "success",
		MediaType: "unknown",
		DeviceState: "idle",
		LastUpdate: now,
	})
	atv.eventSink.RegisterEventType("playing", &AppleTVStatus{
		Result: "success",
		MediaType: "video",
		DeviceState: "playing",
		Title: stringp("Stranger Things 4"),
		TotalTime: intp(3746),
		Position: intp(1482),
		App: stringp("Netflix"),
		LastUpdate: now,
	})
	atv.eventSink.RegisterEventType("paused", &AppleTVStatus{
		Result: "success",
		MediaType: "video",
		DeviceState: "paused",
		Title: stringp("Stranger Things 4"),
		TotalTime: intp(3746),
		Position: intp(1482),
		App: stringp("Netflix"),
		LastUpdate: now,
	})
}

func (atv *AppleTV) updateState(update *AppleTVStatus) {
	prev := atv.state
	state := prev.Clone()
	state.LastUpdate = update.LastUpdate
	if update.PowerState != nil {
		state.PowerState = update.PowerState
		if update.DeviceState == "" {
			state.PowerState = "idle"
		} else {
			state.PowerState = update.DeviceState
		}
		atv.state = state
		atv.eventSink.Emit("power", update)
	} else if update.Connection != nil {
		state.Connection = update.Connection
		atv.state = state
		atv.eventSink.Emit("connection", update)
	} else {
		if state.Connection == nil || *state.Connection != "open" {
			state.Connection = stringp("open")
			atv.state = state
			atv.eventSink.Emit("connection", &AppleTVState{
				Connection: state.Connection,
				LastUpdate: state.LastUpdate,
			})
			state = state.Clone()
		}
		if state.PowerState == nil || *state.PowerState != "on" {
			state.PowerState = stringp("on")
			atv.state = state
			atv.eventSink.Emit("power", &AppleTVStatus{
				PowerState: state.PowerState,
				LastUpdate: state.LastUpdate,
			})
			state = state.Clone()
		}
		if update.DeviceState != "" && update.DeviceState != state.DeviceState {
			state.DeviceState = update.DeviceState
			atv.state = state
			atv.eventSink.Emit(update.DeviceState, update)
			state = state.Clone()
		}
		if state.ContentChanged(update) {
			state.Hash = update.Hash
			state.MediaType = update.MediaType
			state.Title = update.Title
			state.Artist = update.Artist
			state.Album = update.Album
			state.Genre = update.Genre
			state.TotalTime = update.TotalTime
			state.Position = update.Position
			state.Shuffle = update.Shuffle
			state.Repeat = update.Repeat
			state.SeriesName = update.SeriesName
			state.SeasonNumber = update.SeasonNumber
			state.EpisodeNumber = update.EpisodeNumber
			state.ContentIdentifier = update.ContentIdentifier
			state.App = update.App
			state.AppID = update.AppID
			atv.state = state
			atv.eventSink.Emit("content", update)
		}
	}
}

func (atv *AppleTV) Check() (interface{}, error) {
	return atv.state, nil
}

func (atv *AppleTV) Monitor(interval time.Duration) func() {
	quitch := make(chan bool, 1)
	quit := false
	quitFunc := func() {
		if !quit {
			quit = true
			close(quitch)
		}
	}
	updates, reconnect, disconnect := atv.Connect()
	go func() {
		for {
			timer := time.NewTimer(10 * time.Minute)
			select {
			case <-quitch:
				disconnect()
				return
			case <-timer.C:
				reconnect()
			case msg, ok := <-updates:
				if !ok {
					return
				}
				atv.updateState(msg)
			}
		}
	}()
	return quitFunc
}

func (atv *AppleTV) HandleRead(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	_, data, err := atv.Check()
	return data, err
}

func (atv *AppleTV) HandleListWebhooks(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return atv.webhooks.webhooks, nil
}

func (atv *AppleTV) HandleAddWebhook(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	webhook := &ThresholdWebhook{}
	err := httpserver.ReadJSON(r, webhook)
	if err != nil {
		return nil, err
	}
	atv.webhooks.RegisterWebhook(webhook)
	return atv.webhooks.List(), nil
}

func (atv *AppleTV) HandleRemoveWebhook(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	webhook := &ThresholdWebhook{}
	err := httpserver.ReadJSON(r, webhook)
	if err != nil {
		return nil, err
	}
	atv.webhooks.UnregisterWebhook(webhook)
	return atv.webhooks.List(), nil
}

func (atv *AppleTV) Connect() (updates chan *AppleTVStatus, reconnect func(), disconnect func()){
	quitch := make(chan bool, 1)
	quit := false
	disconnect = func() {
		if !quit {
			quit = true
			close(quitch)
		}
	}
	reconch := make(chan bool, 1)
	reconnect = func() {
		reconch <- true
	}
	updates := make(chan *AppleTVStatus, 20)
	go func() {
		kill, died := atv.connectOnce(updates)
		for {
			select {
			case <-quitch:
				kill()
				close(updates)
				return
			case <-reconch:
				kill()
				kill, died = atv.connect(updates)
			case err := <-died:
				if err != nil {
					time.Sleep(10 * time.Second)
				}
				kill, died = atv.connect(updates)
			}
		}
	}()
		if err != nil {
			log.Println("error running appletv connection:", err)
			time.Sleep(10 * time.Second)
		}
	}
}

func (atv *AppleTV) connectOnce(updates chan *AppleTVStatus) (kill func(), died chan error) {
	cmdIn := atv.cmdIn
	atv.cmdIn = nil
	if cmdIn != nil {
		cmdIn.Write([]byte{'\n'})
		cmdIn.Close()
	}
	cmd := exec.Command("atvscript", "--id", atv.cfg.AppleTV.ID, "--airplay-credentials", atv.cfg.AppleTV.Airplay, "push_updates")
	cmdIn, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	cmdOut, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	log.Println(strings.Join(cmd.Args, " "))
	err = cmd.Start()
	if err != nil {
		return err
	}
	kill = func() {
		cmdIn.Write([]byte{'\n'})
		cmdIn.Close()
		p := cmd.Process
		if p != nil {
			p.Kill()
		}
	}
	died = make(chan error, 1)
	go func() {
		bufOut := bufio.NewReader(cmdOut)
		for {
			line, err := bufOut.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					p := cmd.Process
					if p != nil {
						p.Wait()
					}
					close(died)
					return
				}
				log.Println("error reading from command:", err)
				p := cmd.Process
				if p != nil {
					p.Kill()
				}
				died <- err
				return
			}
			log.Println("appletv:", string(line))
			msg := &AppleTVStatus{}
			err = json.Unmarshal([]byte(line), msg)
			if err != nil {
				log.Println("error unmarshaling message", line, err)
			}
			updates <- msg
		}
	}()
	return
}
