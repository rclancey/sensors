package api

import (
	"embed"
	//"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rclancey/events"
	H "github.com/rclancey/httpserver/v2"
	"github.com/rclancey/logging"
)

//go:embed ui/*
var ui embed.FS

func APIMain() {
	var errlog *logging.Logger
	var srv *H.Server
	var err error
	shutdown := false
	for !shutdown {
		sigch := make(chan os.Signal, 10)
		go func() {
			sig, ok := <-sigch
			if !ok || sig == nil {
				log.Println("no signal!")
				return
			}
			log.Println("handling signal", sig)
			switch sig {
			case syscall.SIGINT:
				log.Println("got SIGINT")
				shutdown = true
				if srv != nil {
					if errlog != nil {
						errlog.Infoln("SIGINT")
					}
					srv.Shutdown()
				}
			case syscall.SIGHUP:
				log.Println("got SIGHUP")
				if srv != nil {
					if errlog != nil {
						errlog.Infoln("SIGHUP")
					}
					srv.Shutdown()
				}
			}
		}()
		signal.Notify(sigch, syscall.SIGINT, syscall.SIGHUP)
		errlog, srv, err = startup()
		if err != nil {
			break
		}
		srv.ListenAndServe()
		errlog.Infoln("server shut down")
		close(sigch)
	}
	errlog.Infoln("server exiting")
}

func colorizeLogger(l *logging.Logger) {
	l.Colorize()
	l.SetLevelColor(logging.INFO, logging.ColorCyan, logging.ColorDefault, logging.FontDefault)
	l.SetLevelColor(logging.LOG, logging.ColorMagenta, logging.ColorDefault, logging.FontDefault)
	l.SetLevelColor(logging.NONE, logging.ColorHotPink, logging.ColorDefault, logging.FontDefault)
	l.SetTimeFormat("2006-01-02 15:04:05.000")
	l.SetTimeColor(logging.ColorDefault, logging.ColorDefault, logging.FontItalic | logging.FontLight)
	l.SetSourceFormat("%{basepath}:%{linenumber}:")
	l.SetSourceColor(logging.ColorGreen, logging.ColorDefault, logging.FontDefault)
	l.SetPrefixColor(logging.ColorOrange, logging.ColorDefault, logging.FontDefault)
	l.SetMessageColor(logging.ColorDefault, logging.ColorDefault, logging.FontDefault)
	l.MakeDefault()
}

func startup() (*logging.Logger, *H.Server, error) {
	var err error
	cfg, err := Configure()
	if err != nil {
		log.Fatalln("error configuring server:", err)
	}
	if cfg == nil {
		log.Fatalln("no configuration found")
	}
	errlog, err := cfg.Logging.ErrorLogger()
	if err != nil {
		log.Fatalln("error configuring logging:", err)
	}
	colorizeLogger(errlog)

	srv, err := H.NewServer(cfg.ServerConfig)
	if err != nil {
		log.Fatalln("can't create server:", err)
	}

	eventLogFn, err := cfg.Abs("var/log/events.log")
	if err != nil {
		log.Println("can't get event log:", err)
	}
	eventLog, err := os.OpenFile(eventLogFn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln("can't open event log:", err)
	}
	eventSink := events.NewLoggedEventSink(events.NewEventSink(24 * time.Hour), eventLog)

	bs, err := NewBrightnessSensor(cfg, eventSink)
	if err != nil {
		log.Fatalln("can't create brightness sensor:", err)
	}
	bs.Monitor(time.Minute)

	ms, err := NewMotionSensor(cfg, eventSink)
	if err != nil {
		log.Fatalln("can't create motion sensor:", err)
	}
	ms.Monitor(500 * time.Millisecond)

	ls, err := NewLightStatus(cfg, eventSink)
	if err != nil {
		log.Fatalln("can't create light status:", err)
	}
	ls.Monitor(time.Minute)

	atv, err := NewAppleTV(cfg, eventSink)
	if err != nil {
		log.Fatalln("can't create appletv:", err)
	}
	atv.Monitor(time.Second)

	sonos, err := NewSonos(cfg, eventSink)
	if err != nil {
		log.Fatalln("can't create sonos:", err)
	}
	sonos.Monitor(5 * time.Second)

	ns, err := NewNetworkStatus(cfg, eventSink)
	if err != nil {
		log.Fatalln("can't create network:", err)
	}
	ns.Monitor(5 * time.Minute)

	weather, err := NewWeather(cfg, eventSink)
	if err != nil {
		log.Fatalln("can't create weather station:", err)
	}
	weather.Monitor(15 * time.Minute)

	whl, err := NewWebhookList(cfg, eventSink)
	if err != nil {
		log.Fatalln("can't create webhook list:", err)
	}

	go func() {
		log.Println("running initial checks...")
		bs.Check()
		log.Println("brightness")
		ms.Check()
		log.Println("motion")
		ls.Check()
		log.Println("lights")
		sonos.Check()
		log.Println("sonos")
		atv.Check()
		log.Println("appletv")
		ns.Check()
		log.Println("network")
		weather.Check()
		log.Println("weather")
	}()

	srv.RegisterOnShutdown(func() {
		log.Println("cleanup globals on shutdown")
		//bs.webhooks.Stop()
		//ms.webhooks.Stop()
	})

	/*
	uifs, err := fs.Sub(ui, "ui")
	if err != nil {
		log.Fatal(err)
	}
	*/
	errlog.Infoln("server starting...")
	srv.GET("/brightness/status", H.HandlerFunc(bs.HandleRead))
	srv.GET("/motion/status", H.HandlerFunc(ms.HandleRead))
	srv.GET("/lights/status", H.HandlerFunc(ls.HandleRead))
	srv.GET("/sonos/status", H.HandlerFunc(sonos.HandleRead))
	srv.GET("/network/status", H.HandlerFunc(ns.HandleRead))
	srv.GET("/weather/status", H.HandlerFunc(weather.HandleRead))

	srv.GET("/status", indexFunc(map[string]sensor{
		"brightness": bs,
		"motion": ms,
		"lights": ls,
		"appletv": atv,
		"sonos": sonos,
		"network": ns,
		"weather": weather,
	}))

	srv.PUT("/lights/", H.HandlerFunc(ls.HandlePut))
	srv.PUT("/sonos/volume", H.HandlerFunc(sonos.HandleSetVolume))
	srv.POST("/sonos/playliist", H.HandlerFunc(sonos.HandleSetPlaylist))
	srv.PUT("/sonos/playback", H.HandlerFunc(sonos.HandleSetPlayback))

	srv.GET("/event-types", H.HandlerFunc(whl.HandleListEventTypes))
	srv.GET("/events", H.HandlerFunc(whl.HandleEventLog))
	srv.GET("/webhooks", H.HandlerFunc(whl.HandleListWebhooks))
	srv.POST("/webhooks", H.HandlerFunc(whl.HandleAddWebhook))
	srv.DELETE("/webhooks", H.HandlerFunc(whl.HandleRemoveWebhook))

	//srv.GET("/", http.FileServer(http.FS(uifs)))
	errlog.Infoln("server ready")

	return errlog, srv, nil
}

type sensor interface {
	HandleRead(http.ResponseWriter, *http.Request) (interface{}, error)
}

func indexFunc(sensors map[string]sensor) H.HandlerFunc {
	return H.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (interface{}, error) {
		resp := map[string]interface{}{}
		for k, h := range sensors {
			v, err := h.HandleRead(w, r)
			if err != nil {
				resp[k] = map[string]interface{}{"error": err.Error()}
			} else {
				resp[k] = v
			}
		}
		return resp, nil
	})
}
