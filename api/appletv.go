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

	"github.com/rclancey/events"
	//"github.com/rclancey/httpserver/v2"
	"github.com/rclancey/sensors/types"
)

type AppleTV struct {
	cfg *Config
	eventSink events.EventSink
	cmdIn io.WriteCloser
	state *types.AppleTVStatus
	updates chan *types.AppleTVStatus
	cmd *exec.Cmd
}

func stringp(s string) *string {
	return &s
}

func intp(i int) *int {
	return &i
}

func NewAppleTV(cfg *Config, eventSink events.EventSink) (*AppleTV, error) {
	sink := events.NewPrefixedEventSource("appletv", eventSink)
	atv := &AppleTV{
		cfg: cfg,
		eventSink: sink,
		state: &types.AppleTVStatus{
			DeviceState: "idle",
			PowerState: stringp("off"),
			Connection: stringp("closed"),
			LastUpdate: time.Now(),
		},
		updates: make(chan *types.AppleTVStatus, 100),
	}
	atv.registerEventTypes()
	go atv.Connect()
	return atv, nil
}

func (atv *AppleTV) registerEventTypes() {
	now := time.Now().In(time.Local)
	atv.eventSink.RegisterEventType(events.NewEvent("power", &types.AppleTVStatus{
		PowerState: stringp("on"),
		LastUpdate: now,
	}))
	atv.eventSink.RegisterEventType(events.NewEvent("connection", &types.AppleTVStatus{
		Connection: stringp("open"),
		LastUpdate: now,
	}))
	atv.eventSink.RegisterEventType(events.NewEvent("content", &types.AppleTVStatus{
		Result: "success",
		MediaType: "video",
		DeviceState: "playing",
		Title: stringp("Stranger Things 4"),
		TotalTime: intp(3746),
		Position: intp(0),
		App: stringp("Netflix"),
		LastUpdate: now,
	}))
	atv.eventSink.RegisterEventType(events.NewEvent("idle", &types.AppleTVStatus{
		Result: "success",
		MediaType: "unknown",
		DeviceState: "idle",
		LastUpdate: now,
	}))
	atv.eventSink.RegisterEventType(events.NewEvent("playing", &types.AppleTVStatus{
		Result: "success",
		MediaType: "video",
		DeviceState: "playing",
		Title: stringp("Stranger Things 4"),
		TotalTime: intp(3746),
		Position: intp(1482),
		App: stringp("Netflix"),
		LastUpdate: now,
	}))
	atv.eventSink.RegisterEventType(events.NewEvent("paused", &types.AppleTVStatus{
		Result: "success",
		MediaType: "video",
		DeviceState: "paused",
		Title: stringp("Stranger Things 4"),
		TotalTime: intp(3746),
		Position: intp(1482),
		App: stringp("Netflix"),
		LastUpdate: now,
	}))
}

func (atv *AppleTV) updateState(update *types.AppleTVStatus) {
	prev := atv.state
	state := prev.Clone()
	state.LastUpdate = update.LastUpdate
	if update.PowerState != nil {
		state.PowerState = update.PowerState
		if update.DeviceState == "" {
			state.PowerState = stringp("idle")
		} else {
			state.PowerState = update.PowerState
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
			atv.eventSink.Emit("connection", &types.AppleTVStatus{
				Connection: state.Connection,
				LastUpdate: state.LastUpdate,
			})
			state = state.Clone()
		}
		if state.PowerState == nil || *state.PowerState != "on" {
			state.PowerState = stringp("on")
			atv.state = state
			atv.eventSink.Emit("power", &types.AppleTVStatus{
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
	if atv.state.DeviceState == "playing" {
		Measure("appletv_playing", nil, 1.0)
	} else {
		Measure("appletv_playing", nil, 0.0)
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
	return atv.Check()
}

func (atv *AppleTV) Connect() (updates chan *types.AppleTVStatus, reconnect func(), disconnect func()) {
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
	updates = make(chan *types.AppleTVStatus, 20)
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
				kill, died = atv.connectOnce(updates)
			case err := <-died:
				if err != nil {
					time.Sleep(10 * time.Second)
				}
				kill, died = atv.connectOnce(updates)
			}
		}
	}()
	return
}

func (atv *AppleTV) connectOnce(updates chan *types.AppleTVStatus) (kill func(), died chan error) {
	kill = func() {}
	died = make(chan error, 1)
	cmd := exec.Command("atvscript", "--id", atv.cfg.AppleTV.ID, "--airplay-credentials", atv.cfg.AppleTV.Airplay, "push_updates")
	cmdIn, err := cmd.StdinPipe()
	if err != nil {
		died <- err
		return kill, died
	}
	cmdOut, err := cmd.StdoutPipe()
	if err != nil {
		died <- err
		return kill, died
	}
	log.Println(strings.Join(cmd.Args, " "))
	err = cmd.Start()
	if err != nil {
		died <- err
		return kill, died
	}
	killed := false
	kill = func() {
		if !killed {
			killed = true
			cmdIn.Write([]byte{'\n'})
			cmdIn.Close()
			p := cmd.Process
			if p != nil {
				p.Kill()
			}
		}
	}
	go func() {
		err := cmd.Wait()
		if err != nil {
			died <- err
		}
		close(died)
	}()
	go func() {
		bufOut := bufio.NewReader(cmdOut)
		for {
			line, err := bufOut.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				log.Println("error reading from command:", err)
				kill()
				return
			}
			now := time.Now().In(time.UTC)
			log.Println("appletv:", string(line))
			msg := &types.AppleTVStatus{}
			err = json.Unmarshal([]byte(line), msg)
			if err != nil {
				log.Println("error unmarshaling message", line, err)
				continue
			}
			msg.LastUpdate = now
			updates <- msg
		}
	}()
	atv.eventSink.Emit("restart", &types.AppleTVStatus{LastUpdate: time.Now().In(time.UTC)})
	return
}
