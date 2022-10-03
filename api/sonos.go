package api

import (
	//"context"
	//"crypto/sha256"
	//"encoding/hex"
	"encoding/json"
	"errors"
	//"fmt"
	"log"
	"net"
	"net/http"
	//"net/url"
	//"os"
	//"path"
	//"path/filepath"
	"strconv"
	"strings"
	//"sync"
	"time"

	"github.com/rclancey/go-sonos"
	"github.com/rclancey/go-sonos/ssdp"
	"github.com/rclancey/go-sonos/upnp"
	"github.com/rclancey/httpserver/v2"
	//"github.com/rclancey/synos/musicdb"
)

const (
	PlayModeShuffle = 1
	PlayModeRepeat = 2
)

type Sonos struct {
	cfg *Config
	webhooks *ThresholdWebhookList
	mgrPort int
	reactorPort int
	dev ssdp.Device
	player *sonos.Sonos
	reactor upnp.Reactor
	state *SonosState
}

func NewSonos(cfg *Config) (*Sonos, error) {
	fn, err := cfg.Abs("motion-sensor-webhooks.json")
	if err != nil {
		return nil, err
	}
	webhooks, err := NewThresholdWebhookList(fn)
	if err != nil {
		return nil, err
	}
	s := &Sonos{
		cfg: cfg,
		webhooks: webhooks,
	}
	iface := cfg.Network.GetInterface()
	var ok bool
	s.mgrPort, ok = findFreePort(11209, 11299)
	if !ok {
		return nil, errors.New("no free port")
	}
	s.reactorPort, ok = findFreePort(s.mgrPort, 11299)
	if !ok {
		return nil, errors.New("no free port")
	}
	if err != nil {
		return nil, err
	}
	s.dev, err = s.getDevice()
	if err != nil {
		return nil, err
	}
	s.reactor = sonos.MakeReactor(iface.Name, strconv.Itoa(s.reactorPort))
	go func() {
		c := s.reactor.Channel()
		for {
			msg, ok := <-c
			if !ok {
				break
			}
			msgj, err := json.Marshal(msg)
			if err != nil {
				log.Println("sonos event", msg)
			} else {
				log.Println("sonos event", string(msgj))
			}
		}
	}()
	s.player = sonos.Connect(s.dev, s.reactor, sonos.SVC_CONNECTION_MANAGER|sonos.SVC_CONTENT_DIRECTORY|sonos.SVC_RENDERING_CONTROL|sonos.SVC_AV_TRANSPORT)
	return s, nil
}

func (s *Sonos) Check() (float64, interface{}, error) {
	state, err := s.GetState()
	if err != nil {
		return -1, nil, err
	}
	s.state = state
	if state.State != nil && state.Index != nil && *state.State == "PLAYING" {
		return float64(*state.Index), state, nil
	}
	return -1, state, nil
}

func (s *Sonos) Monitor(interval time.Duration) {
	go s.webhooks.Monitor(s, interval)
}

func (s *Sonos) HandleRead(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return s.state, nil
}

func (s *Sonos) HandleListWebhooks(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return s.webhooks.webhooks, nil
}

func (s *Sonos) HandleAddWebhook(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	webhook := &ThresholdWebhook{}
	err := httpserver.ReadJSON(r, webhook)
	if err != nil {
		return nil, err
	}
	s.webhooks.RegisterWebhook(webhook)
	return s.webhooks.List(), nil
}

func (s *Sonos) HandleRemoveWebhook(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	webhook := &ThresholdWebhook{}
	err := httpserver.ReadJSON(r, webhook)
	if err != nil {
		return nil, err
	}
	s.webhooks.UnregisterWebhook(webhook)
	return s.webhooks.List(), nil
}

func (s *Sonos) Shutdown() error {
	return errors.New("not implemented")
}

func findFreePort(start, end int) (int, bool) {
	for i := start; i < end; i += 1 {
		ln, err := net.Listen("tcp", ":" + strconv.Itoa(i))
		if err == nil {
			ln.Close()
			return i, true
		}
	}
	return -1, false
}

func (s *Sonos) getDevice() (ssdp.Device, error) {
	iface := s.cfg.Network.GetInterface()
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

/*
func (s *Sonos) makeFileServer() *http.Server {
	getFileName := func(id string) (string, bool) {
		s.filesLock.Lock()
		defer s.filesLock.Unlock()
		fn, ok := s.files[id]
		return fn, ok
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		id := path.Base(r.URL.Path)
		fn, ok := getFileName(id)
		if !ok {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
			return
		}
		http.ServeFile(w, r, fn)
	}
	port, _ := findFreePort(s.reactorPort, 11299)
	s.server = &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(handler),
	}
	go func() {
		s.server.ListenAndServe()
		s.server = nil
	}()
	return s.server
}
*/

/*
func (s *Sonos) getFileUri(fn string) (string, string) {
	cfg := &httpserver.NetworkConfig{
		Network: "192.168.0.0/24",
	}
	ip := cfg.GetIP()
	hash := sha256.Sum256([]byte(fn))
	id := hex.EncodeToString(hash[:16]) + filepath.Ext(fn)
	u := &url.URL{
		Scheme: "http",
		Host: ip.String() + s.server.Addr,
		Path: "/" + id,
	}
	return id, u.String()
}
*/

/*
func (s *Sonos) serveFile(fn string) string {
	id, uri := s.getFileUri(fn)
	s.filesLock.Lock()
	defer s.filesLock.Unlock()
	s.files[id] = fn
	return uri
}
*/

/*
func (s *Sonos) playlistUri(pl *musicdb.Playlist) string {
	fn := filepath.Join("/tmp", "sonos", pl.PersistentID.String() + ".m3u")
	dn := filepath.Dir(fn)
	if _, err := os.Stat(dn); err != nil {
		os.MkdirAll(dn, 0777)
	}
	f, err := os.Create(fn)
	if err != nil {
		log.Printf("error creating file %s: %s", fn, err)
		return ""
	}
	f.Write([]byte("#EXTM3U\n"))
	s.tracksLock.Lock()
	for _, track := range pl.PlaylistItems {
		var t int
		var artist, album, title string
		if track.TotalTime != nil {
			t = int(*track.TotalTime / 1000)
		}
		if track.Artist != nil {
			artist = *track.Artist
		}
		if track.Album != nil {
			album = *track.Album
		}
		if track.Name != nil {
			title = *track.Name
		}
		f.Write([]byte(fmt.Sprintf("#EXTINF:%d,<%s><%s><%s>\n", t, artist, album, title)))
		uri := s.serveFile(track.Path())
		s.tracks[uri] = track
		f.Write([]byte(uri))
		f.Write([]byte("\n"))
	}
	s.tracksLock.Unlock()
	f.Close()
	return s.serveFile(fn)
}
*/

/*
func (s *Sonos) didlLite(track *musicdb.Track) string {
	trackId := track.PersistentID.String()
	mediaUri := s.serveFile(track.Path())
	duration := "0:00"
	if track.TotalTime != nil {
		hours := *track.TotalTime / 3600000
		mins := (*track.TotalTime % 3600000) / 60000
		secs := (*track.TotalTime % 60000) / 1000
		duration = fmt.Sprintf("%d:%02d:%02d", hours, mins, secs)
	}
	title, _ := track.GetName()
	artist, _ := track.GetArtist()
	album, _ := track.GetAlbum()
	s.tracksLock.Lock()
	s.tracks[mediaUri] = track
	s.tracksLock.Unlock()
	return fmt.Sprintf(`<DIDL-Lite xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:upnp="urn:schemas-upnp-org:metadata-1-0/upnp/" xmlns:r="urn:schemas-rinconnetworks-com:metadata-1-0/" xmlns="urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/">
  <item id="%s" parentID="%s">
	<upnp:class>object.item.audioItem.musicTrack</upnp:class>
	<res protocolInfo="http-get:*:audio/mpeg:*" duration="%s">%s</res>
	<dc:title>%s</dc:title>
	<dc:creator>%s</dc:creator>
	<upnp:album>%s</upnp:album>
  </item>
</DIDL-Lite>`, trackId, trackId, duration, mediaUri, title, artist, album)
}
*/

/*
func (s *Sonos) didlLitePl(pl *musicdb.Playlist) string {
	plId := pl.PersistentID.String()
	return fmt.Sprintf(`<DIDL-Lite xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:upnp="urn:schemas-upnp-org:metadata-1-0/upnp/" xmlns:r="urn:schemas-rinconnetworks-com:metadata-1-0/" xmlns="urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/">
	<item id="playlists:%s" parentID="playlists:%s" restricted="true">
		<dc:title>Playlists</dc:title>
		<upnp:class>object.container</upnp:class>
		<desc id="cdudn" nameSpace="urn:schemas-rinconnetworks-com:metadata-1-0/">%s</desc>
	</item>
</DIDL-Lite>`, plId, plId, pl.Name)
}
*/

/*
func (s *Sonos) PlayTrack(tr *musicdb.Track) error {
	if cfg.TestMode {
		if tr.Name != nil {
			log.Printf("TEST: play track %s", *tr.Name)
		}
		return nil
	}
	err := s.player.Stop(0)
	if err != nil {
		return err
	}
	err = s.player.RemoveAllTracksFromQueue(0)
	if err != nil {
		return err
	}
	err = s.EnqueueTracks(tr)
	if err != nil {
		return err
	}
	return s.player.Play(0, "1")
}
*/

/*
func (s *Sonos) PlayPlaylist(pl *musicdb.Playlist) error {
	if cfg.TestMode {
		log.Printf("TEST: play playlist %s", pl.Name)
		return nil
	}
	err := s.player.Stop(0)
	if err != nil {
		return err
	}
	err = s.player.RemoveAllTracksFromQueue(0)
	if err != nil {
		return err
	}
	err = s.EnqueuePlaylist(pl)
	if err != nil {
		return err
	}
	return s.player.Play(0, "1")
}
*/

/*
func (s *Sonos) PlayFile(fn string) error {
	if cfg.TestMode {
		log.Printf("TEST: play %s", fn)
		return nil
	}
	err := s.player.Stop(0)
	if err != nil {
		return err
	}
	err = s.player.RemoveAllTracksFromQueue(0)
	if err != nil {
		return err
	}
	err = s.EnqueueFile(fn)
	if err != nil {
		return err
	}
	return s.player.Play(0, "1")
}
*/

func (s *Sonos) Stop() error {
	/*
	if cfg.TestMode {
		log.Println("TEST: stop")
		return nil
	}
	*/
	return s.player.Stop(0)
}

/*
func (s *Sonos) EnqueueTracks(tracks ...*musicdb.Track) error {
	for _, track := range tracks {
		uri := s.serveFile(track.Path())
		req := &upnp.AddURIToQueueIn{
			EnqueuedURI: uri,
			EnqueuedURIMetaData: s.didlLite(track),
		}
		_, err := s.player.AddURIToQueue(0, req)
		if err != nil {
			log.Printf("error adding track %s to queue: %s", track.Path(), err)
			return err
		}
	}
	return nil
}
*/

/*
func (s *Sonos) EnqueuePlaylist(pl *musicdb.Playlist) error {
	uri := s.playlistUri(pl)
	if uri == "" {
		return errors.New("error creating playlist file")
	}
	req := &upnp.AddURIToQueueIn{
		EnqueuedURI: uri,
		EnqueuedURIMetaData: s.didlLitePl(pl),
	}
	_, err := s.player.AddURIToQueue(0, req)
	if err != nil {
		log.Printf("error adding playlist %s to queue: %s", pl.Name, err)
		return err
	}
	return nil
}
*/

/*
func (s *Sonos) EnqueueFile(fn string) error {
	if cfg.TestMode {
		//log.Printf("TEST: enqueue %s", fn)
		return nil
	}
	uri := s.serveFile(fn)
	req := &upnp.AddURIToQueueIn{
		EnqueuedURI: uri,
	}
	_, err := s.player.AddURIToQueue(0, req)
	if err != nil {
		return err
	}
	return nil
}
*/

type SonosTrack struct {
	URI string `json:"uri"`
	Title string `json:"title,omitempty"`
	Artist string `json:"artist,omitempty"`
	Album string `json:"album,omitempty"`
	Time int64 `json:"time,omitempty"`
}

type SonosState struct {
	Tracks []*SonosTrack `json:"tracks,omitempty"`
	Index *int `json:"index,omitempty"`
	Duration *int `json:"duration,omitempty"`
	Time *int `json:"time,omitempty"`
	State *string `json:"state,omitempty"`
	Speed *float64 `json:"speed,omitempty"`
	Volume *int `json:"volume,omitempty"`
	PlayMode *int `json:"mode,omitempty"`
	Error error `json:"error,omitempty"`
	LastUpdate time.Time `json:"now"`
}

var refTime = time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC)
func parseTime(timestr string, layouts ...string) (int, error) {
	var err error
	var t time.Time
	for _, l := range layouts {
		t, err = time.Parse(l, timestr)
		if err == nil {
			return int(t.Sub(refTime).Nanoseconds() / 1000000), nil
		}
	}
	return -1, err
}

func (s *Sonos) GetPlaybackStatus() (*SonosState, error) {
	info, err := s.player.GetTransportInfo(0)
	if err != nil {
		return nil, err
	}
	state := &SonosState{
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

func (s *Sonos) GetQueuePos() (*SonosState, error) {
	pos, err := s.player.GetPositionInfo(0)
	if err != nil {
		return nil, err
	}
	index := int(pos.Track) - 1
	state := &SonosState{
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

func (s *Sonos) GetState() (state *SonosState, xerr error) {
	state = nil
	xerr = nil
	defer func() {
		if r := recover(); r != nil {
			rs, isa := r.(string)
			if isa {
				xerr = errors.New(rs)
			} else {
				xerr = errors.New("panic communicating with sonos")
			}
		}
	}()
	state, xerr = s.GetPlaybackStatus()
	if xerr != nil {
		return
	}
	objs, xerr := s.player.GetQueueContents()
	if xerr != nil {
		return
	}
	tracks := make([]*SonosTrack, len(objs))
	//s.tracksLock.Lock()
	for i, item := range objs {
		tracks[i] = &SonosTrack{
			URI: item.Res(),
			Title: item.Title(),
			Artist: item.Creator(),
			Album: item.Album(),
		}
		/*
		t := s.tracks[item.Res()]
		if t != nil {
			if t.Name != nil {
				tracks[i].Title = *t.Name
			}
			if t.Artist != nil {
				tracks[i].Artist = *t.Artist
			}
			if t.Album != nil {
				tracks[i].Album = *t.Album
			}
			if t.TotalTime != nil {
				tracks[i].Time = int64(*t.TotalTime)
			}
		}
		*/
	}
	//s.tracksLock.Unlock()
	state.Tracks = tracks
	pos, xerr := s.GetQueuePos()
	if xerr != nil {
		return
	}
	state.Index = pos.Index
	state.Duration = pos.Duration
	state.Time = pos.Time
	vol, xerr := s.GetVolume()
	if xerr != nil {
		return
	}
	state.Volume = &vol
	mode, xerr := s.GetPlayMode()
	if xerr != nil {
		return
	}
	state.PlayMode = &mode
	state.LastUpdate = time.Now()
	return
}

func (s *Sonos) GetVolume() (int, error) {
	vol, err := s.player.GetVolume(0, "Master")
	if err != nil {
		return -1, err
	}
	return int(vol), nil
}

func (s *Sonos) SetVolume(vol int) error {
	if vol > 100 {
		vol = 100
	} else if vol < 0 {
		vol = 0
	}
	return s.player.SetVolume(0, "Master", uint16(vol))
}
