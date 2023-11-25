package sonos

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rclancey/go-sonos"
	"github.com/rclancey/go-sonos/ssdp"
	"github.com/rclancey/go-sonos/upnp"
	"github.com/rclancey/httpserver/v2"
)

const (
	PlayModeShuffle = int(1)
	PlayModeRepeat  = int(2)

	PlayStatePlaying = "PLAYING"
	PlayStatePaused  = "PLAYBACK_PAUSED"
	PlayStateStopped = "STOPPED"
)

type Sonos struct {
	cfg *httpserver.NetworkConfig
	mgrPort int
	reactorPort int
	dev ssdp.Device
	player *sonos.Sonos
	reactor upnp.Reactor
	fileServer *FileServer
	state *State
	updates chan *State
	mutex *sync.Mutex
}

func NewSonos(cfg *httpserver.NetworkConfig, updates chan *State) (*Sonos, error) {
	fs, err := NewFileServer(cfg)
	if err != nil {
		return nil, err
	}
	s := &Sonos{
		cfg: cfg,
		fileServer: fs,
		state: &State{LastUpdate: time.Now()},
		updates: updates,
		mutex: &sync.Mutex{},
	}
	iface := cfg.GetInterface()
	var ok bool
	s.mgrPort, ok = findFreePort(11209, 11299)
	if !ok {
		return nil, errors.New("no free port")
	}
	s.reactorPort, ok = findFreePort(s.mgrPort, 11299)
	if !ok {
		return nil, errors.New("no free port")
	}
	s.dev, err = s.getDevice()
	if err != nil {
		return nil, err
	}
	s.reactor = sonos.MakeReactor(iface.Name, strconv.Itoa(s.reactorPort))
	err = s.reconnect()
	if err != nil {
		return nil, err
	}
	go func() {
		timer := time.NewTimer(10 * time.Minute)
		c := s.reactor.Channel()
		for {
			select {
			case <-timer.C:
				s.reconnect()
			case msg, ok := <-c:
				if !ok {
					break
				}
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(10 * time.Minute)
				s.HandleEvent(msg)
			}
		}
	}()
	return s, nil
}

func (s *Sonos) reconnect() (err error) {
	defer func() {
		err = recoverError()
	}()
	dev, err := s.getDevice()
	if err != nil {
		return err
	}
	s.dev = dev
	s.player = sonos.Connect(s.dev, s.reactor, sonos.SVC_CONNECTION_MANAGER|sonos.SVC_CONTENT_DIRECTORY|sonos.SVC_RENDERING_CONTROL|sonos.SVC_AV_TRANSPORT)
	err = s.PrepareQueue()
	if err != nil {
		return err
	}
	state, err := s.GetState()
	if err != nil {
		s.ApplyUpdate(state)
	}
	return
}

func (s *Sonos) ApplyUpdate(update *State) {
	s.mutex.Lock()
	diff := s.state.Diff(update)
	s.state = s.state.ApplyUpdate(update)
	s.mutex.Unlock()
	select {
	case s.updates <- diff:
	default:
		log.Println("update queue full")
	}
}

var refTime = time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC)
func parseTime(timestr string, layouts ...string) (int, error) {
	var err error
	var t time.Time
	for _, l := range layouts {
		t, err = time.Parse(l, timestr)
		if err == nil {
			return int(t.Sub(refTime).Milliseconds()), nil
		}
	}
	return -1, err
}

func (s *Sonos) GetPlaybackStatus() (*State, error) {
	info, err := s.player.GetTransportInfo(0)
	if err != nil {
		return nil, err
	}
	state := &State{
		State: &info.CurrentTransportState,
	}
	if strings.Contains(info.CurrentSpeed, "/") {
		parts := strings.SplitN(info.CurrentSpeed, "/", 2)
		num, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return nil, err
		}
		den, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return nil, err
		}
		speed := float64(0)
		if den != 0 {
			speed = num / den
		}
		state.Speed = &speed
	} else {
		s, err := strconv.ParseFloat(strings.TrimSpace(info.CurrentSpeed), 64)
		if err != nil {
			return nil, err
		}
		state.Speed = &s
	}
	return state, nil
}

func (s *Sonos) GetVolume() (int, error) {
	vol, err := s.player.GetVolume(0, "Master")
	if err != nil {
		return -1, err
	}
	return int(vol), nil
}

func (s *Sonos) GetMute() (bool, error) {
	mute, err := s.player.GetMute(0, "Master")
	if err != nil {
		return false, err
	}
	return mute, nil
}

func (s *Sonos) GetPlayMode() (int, error) {
	ts, err := s.player.GetTransportSettings(0)
	if err != nil {
		return 0, err
	}
	switch ts.PlayMode {
	case upnp.PlayMode_NORMAL:
		return 0, nil
	case upnp.PlayMode_REPEAT_ALL:
		return PlayModeRepeat, nil
	case upnp.PlayMode_SHUFFLE_NOREPEAT:
		return PlayModeShuffle, nil
	case upnp.PlayMode_SHUFFLE:
		return PlayModeShuffle | PlayModeRepeat, nil
	}
	return 0, nil
}

func (s *Sonos) GetQueue() ([]*Track, error) {
	objs, err := s.player.GetQueueContents()
	if err != nil {
		return nil, err
	}
	tracks := make([]*Track, len(objs))
	for i, item := range objs {
		tracks[i] = &Track{
			URI: item.Res(),
			Title: item.Title(),
			Artist: item.Creator(),
			Album: item.Album(),
			AlbumArtURI: item.AlbumArtURI(),
		}
	}
	return s.localizeTracks(tracks), nil
}

func (s *Sonos) GetQueuePos() (*State, error) {
	pos, err := s.player.GetPositionInfo(0)
	if err != nil {
		return nil, err
	}
	index := int(pos.Track) - 1
	state := &State{
		Index: &index,
	}
	durT, err := parseTime(pos.TrackDuration, "15:04:05", "4:05")
	if err != nil {
		durT = -1
	}
	state.Duration = &durT
	curT, err := parseTime(pos.RelTime, "15:04:05", "4:05")
	if err != nil {
		curT = -1
	}
	state.Time = &curT
	return state, nil
}

func (s *Sonos) GetState() (state *State, err error) {
	defer func() {
		err = recoverError()
	}()
	state, err = s.GetPlaybackStatus()
	if err != nil {
		return
	}
	tracks, err := s.GetQueue()
	if err != nil {
		return
	}
	state.Tracks = tracks
	pos, err := s.GetQueuePos()
	if err != nil {
		return
	}
	state.Index = pos.Index
	state.Duration = pos.Duration
	state.Time = pos.Time
	vol, err := s.GetVolume()
	if err != nil {
		return
	}
	state.Volume = &vol
	mode, err := s.GetPlayMode()
	if err != nil {
		return
	}
	state.PlayMode = &mode
	state.LastUpdate = time.Now()
	return
}

func (s *Sonos) SetVolume(vol int) error {
	if vol > 100 {
		vol = 100
	} else if vol < 0 {
		vol = 0
	}
	update := &State{LastUpdate: time.Now(), Volume: &vol}
	err := s.player.SetVolume(0, "Master", uint16(vol))
	if err == nil {
		s.ApplyUpdate(update)
	}
	return err
}

func (s *Sonos) Mute() error {
	mute := true
	update := &State{LastUpdate: time.Now(), Mute: &mute}
	err := s.player.SetMute(0, "Master", mute)
	if err == nil {
		s.ApplyUpdate(update)
	}
	return err
}

func (s *Sonos) Unmute() error {
	mute := false
	update := &State{LastUpdate: time.Now(), Mute: &mute}
	err := s.player.SetMute(0, "Master", mute)
	if err == nil {
		s.ApplyUpdate(update)
	}
	return err
}

func (s *Sonos) Play() error {
	val := PlayStatePlaying
	update := &State{LastUpdate: time.Now(), State: &val}
	err := s.player.Play(0, "1")
	if err == nil {
		s.ApplyUpdate(update)
	}
	return err
}

func (s *Sonos) Pause() error {
	val := PlayStatePaused
	update := &State{LastUpdate: time.Now(), State: &val}
	err := s.player.Pause(0)
	if err == nil {
		s.ApplyUpdate(update)
	}
	return err
}

func (s *Sonos) Stop() error {
	val := PlayStateStopped
	update := &State{LastUpdate: time.Now(), State: &val}
	err := s.player.Stop(0)
	if err == nil {
		s.ApplyUpdate(update)
	}
	return err
}

func (s *Sonos) SetPlayMode(val int) error {
	var mode string
	switch val {
	case 0:
		mode = upnp.PlayMode_NORMAL
	case PlayModeRepeat:
		mode = upnp.PlayMode_REPEAT_ALL
	case PlayModeShuffle:
		mode = upnp.PlayMode_SHUFFLE_NOREPEAT
	case PlayModeShuffle | PlayModeRepeat:
		mode = upnp.PlayMode_SHUFFLE
	default:
		return fmt.Errorf("Unknown play mode %d", val)
	}
	update := &State{LastUpdate: time.Now(), PlayMode: &val}
	err := s.player.SetPlayMode(0, mode)
	if err == nil {
		s.ApplyUpdate(update)
	}
	return err
}

func (s *Sonos) SetQueuePos(idx int) error {
	update := &State{LastUpdate: time.Now(), Index: &idx}
	err := s.player.Seek(0, "TRACK_NR", strconv.Itoa(idx+1))
	if err == nil {
		s.ApplyUpdate(update)
	}
	return err
}

func (s *Sonos) Skip(n int) error {
	cur := s.state.Index
	var idx int
	if cur == nil {
		idx = n - 1
	} else {
		idx = *cur + n
	}
	return s.SetQueuePos(idx)
}

func (s *Sonos) SkipForward() error {
	return s.Skip(1)
}

func (s *Sonos) SkipBackward() error {
	return s.Skip(-1)
}

func (s *Sonos) SeekTo(ms int) error {
	if ms < 0 {
		return s.SeekTo(0)
	}
	hr := ms / 36000000
	min := (ms % 3600000) / 60000
	sec := (ms % 60000) / 1000
	ts := fmt.Sprintf("%d:%02d:%02d", hr, min, sec)
	update := &State{LastUpdate: time.Now(), Time: &ms}
	err := s.player.Seek(0, "REL_TIME", ts)
	if err == nil {
		s.ApplyUpdate(update)
	}
	return err
}

func (s *Sonos) SeekBy(ms int) error {
	pos, err := s.GetQueuePos()
	if err != nil {
		return err
	}
	if pos.Time == nil {
		return s.SeekTo(ms)
	}
	return s.SeekTo(*pos.Time + ms)
}

func (s *Sonos) UseQueue(id string) error {
	queues, err := s.player.ListQueues()
	if err != nil {
		return err
	}
	for _, q := range queues {
		if q.ID() == id {
			u := q.Res()
			err = s.player.SetAVTransportURI(0, u, "")
			if err != nil {
				return err
			}
			_, err = s.player.ListChildren(q.ID())
			return err
		}
	}
	return fmt.Errorf("queue %s not found", id)
}

func (s *Sonos) PrepareQueue() error {
	info, err := s.player.GetMediaInfo(0)
	if err != nil {
		return err
	}
	if info.CurrentURI == "" {
		err = s.UseQueue("Q:0")
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Sonos) ClearQueue() error {
	err := s.Stop()
	if err != nil {
		return err
	}
	idx := 0
	update := &State{
		LastUpdate: time.Now(),
		Tracks: []*Track{},
		Index: &idx,
	}
	err = s.player.RemoveAllTracksFromQueue(0)
	if err == nil {
		s.ApplyUpdate(update)
	}
	return err
}

func (s *Sonos) localizeURI(uri string) string {
	fileName, ok := s.fileServer.FileForURL(uri)
	if !ok {
		return uri
	}
	u := &url.URL{
		Scheme: "file",
		Path: filepath.ToSlash(fileName),
	}
	return u.String()
}

func (s *Sonos) localizeTracks(tracks []*Track) []*Track {
	out := make([]*Track, len(tracks))
	for i, track := range tracks {
		clone := *track
		clone.URI = s.localizeURI(track.URI)
		out[i] = &clone
	}
	return out
}

func (s *Sonos) serveTracks(tracks []*Track) []*Track {
	out := make([]*Track, len(tracks))
	for i, track := range tracks {
		u, err := url.Parse(track.URI)
		if err != nil || (u.Scheme == "" || u.Scheme == "file") {
			fileName := track.URI
			if u != nil {
				fileName = filepath.FromSlash(u.Path)
			}
			s.fileServer.ServeFile(fileName)
			clone := *track
			clone.URI = s.fileServer.FileURL(fileName)
			out[i] = &clone
		} else {
			out[i] = track
		}
	}
	return out
}

func (s *Sonos) AppendToQueue(tracks []*Track) error {
	for _, track := range s.serveTracks(tracks) {
		didl, err := track.MarshalDIDLLite()
		if err != nil {
			return err
		}
		req := &upnp.AddURIToQueueIn{
			EnqueuedURI: track.URI,
			EnqueuedURIMetaData: string(didl),
		}
		_, err = s.player.AddURIToQueue(0, req)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Sonos) InsertIntoQueue(tracks []*Track, pos int) error {
	for i, track := range s.serveTracks(tracks) {
		didl, err := track.MarshalDIDLLite()
		if err != nil {
			return err
		}
		req := &upnp.AddURIToQueueIn{
			EnqueuedURI: track.URI,
			EnqueuedURIMetaData: string(didl),
			DesiredFirstTrackNumberEnqueued: uint32(pos+i+1),
		}
		_, err = s.player.AddURIToQueue(0, req)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Sonos) ReplaceQueue(tracks []*Track) error {
	err := s.ClearQueue()
	if err != nil {
		return err
	}
	if len(tracks) > 0 {
		err = s.AppendToQueue(tracks[:1])
		if err != nil {
			return err
		}
		err = s.SetQueuePos(0)
		if err != nil {
			return err
		}
		err = s.Play()
		if err != nil {
			return err
		}
		err = s.AppendToQueue(tracks[1:])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Sonos) HandleEvent(event upnp.Event) {
	update := &State{LastUpdate: time.Now()}
	current := s.state
	switch evt := event.(type) {
	case upnp.AVTransportEvent:
		change := evt.LastChange.InstanceID
		playMode := -1
		switch change.CurrentPlayMode.Val {
		case upnp.PlayMode_NORMAL:
			playMode = 0
		case upnp.PlayMode_REPEAT_ALL:
			playMode = PlayModeRepeat
		case upnp.PlayMode_SHUFFLE_NOREPEAT:
			playMode = PlayModeShuffle
		case upnp.PlayMode_SHUFFLE:
			playMode = PlayModeShuffle | PlayModeRepeat
		}
		if playMode >= 0 {
			update.PlayMode = &playMode
		}
		pos, err := strconv.Atoi(change.CurrentTrack.Val)
		if err == nil {
			pos -= 1
			update.Index = &pos
		}
		if update.Index != nil && *update.Index >= 0 {
			if *update.Index >= len(current.Tracks) || s.localizeURI(change.CurrentTrackURI.Val) != current.Tracks[*update.Index].URI {
				tracks, err := s.GetQueue()
				if err == nil {
					update.Tracks = tracks
				}
			}
		}
		s.ApplyUpdate(update)
	case upnp.RenderingControlEvent:
		change := evt.LastChange.InstanceID
		for _, vol := range change.Volume {
			if vol.Channel == "Master" || vol.Channel == "" {
				v, err := strconv.Atoi(vol.Val)
				if err == nil {
					update.Volume = &v
				}
			}
		}
		for _, vol := range change.Mute {
			if vol.Channel == "Master" || vol.Channel == "" {
				v, err := strconv.Atoi(vol.Val)
				if err == nil {
					mute := v > 0
					update.Mute = &mute
				}
			}
		}
		s.ApplyUpdate(update)
	}
}

func (s *Sonos) getDevice() (ssdp.Device, error) {
	iface := s.cfg.GetInterface()
	mgr := ssdp.MakeManager()
	defer mgr.Close()
	mgr.Discover(iface.Name, strconv.Itoa(s.mgrPort), false)
	qry := ssdp.ServiceQueryTerms{
		ssdp.ServiceKey("schemas-upnp-org-MusicServices"): -1,
	}
	result := mgr.QueryServices(qry)
	if dev_list, has := result["schemas-upnp-org-MusicServices"]; has {
		for _, dev := range dev_list {
			if dev.Product() == "Sonos" {
				return dev, nil
			}
		}
	}
	return nil, errors.New("no sonos device found on network")
}
