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

	bs, err := NewBrightnessSensor(cfg.ServerConfig)
	if err != nil {
		log.Fatalln("can't create brightness sensor:", err)
	}
	bs.Monitor(time.Minute)

	ms, err := NewMotionSensor(cfg.ServerConfig)
	if err != nil {
		log.Fatalln("can't create motion sensor:", err)
	}
	ms.Monitor(500 * time.Millisecond)

	ls, err := NewLightStatus(cfg.ServerConfig)
	if err != nil {
		log.Fatalln("can't create light status:", err)
	}
	ls.Monitor(time.Minute)

	atv, err := NewAppleTV(cfg)
	if err != nil {
		log.Fatalln("can't create appletv:", err)
	}
	atv.Monitor(time.Second)

	sonos, err := NewSonos(cfg)
	if err != nil {
		log.Fatalln("can't create sonos:", err)
	}
	sonos.Monitor(5 * time.Second)

	ns, err := NewNetworkStatus(cfg)
	if err != nil {
		log.Fatalln("can't create network:", err)
	}
	ns.Monitor(5 * time.Minute)

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
	}()

	srv.RegisterOnShutdown(func() {
		log.Println("cleanup globals on shutdown")
		bs.webhooks.Stop()
		ms.webhooks.Stop()
	})

	/*
	uifs, err := fs.Sub(ui, "ui")
	if err != nil {
		log.Fatal(err)
	}
	*/
	errlog.Infoln("server starting...")
	srv.GET("/brightness/status", H.HandlerFunc(bs.HandleRead))
	srv.GET("/brightness/webhook", H.HandlerFunc(bs.HandleListWebhooks))
	srv.POST("/brightness/webhook", H.HandlerFunc(bs.HandleAddWebhook))
	srv.DELETE("/brightness/webhook", H.HandlerFunc(bs.HandleRemoveWebhook))

	srv.GET("/motion/status", H.HandlerFunc(ms.HandleRead))
	srv.GET("/motion/webhook", H.HandlerFunc(ms.HandleListWebhooks))
	srv.POST("/motion/webhook", H.HandlerFunc(ms.HandleAddWebhook))
	srv.DELETE("/motion/webhook", H.HandlerFunc(ms.HandleRemoveWebhook))

	srv.GET("/lights/status", H.HandlerFunc(ls.HandleRead))
	srv.GET("/lights/webhook", H.HandlerFunc(ls.HandleListWebhooks))
	srv.POST("/lights/webhook", H.HandlerFunc(ls.HandleAddWebhook))
	srv.DELETE("/lights/webhook", H.HandlerFunc(ls.HandleRemoveWebhook))
	srv.PUT("/lights/", H.HandlerFunc(ls.HandlePut))

	srv.GET("/sonos/status", H.HandlerFunc(sonos.HandleRead))
	srv.GET("/sonos/swebhook", H.HandlerFunc(sonos.HandleListWebhooks))
	srv.POST("/sonos/webhook", H.HandlerFunc(sonos.HandleAddWebhook))
	srv.DELETE("/sonos/webhook", H.HandlerFunc(sonos.HandleRemoveWebhook))

	srv.GET("/network/status", H.HandlerFunc(ns.HandleRead))
	srv.GET("/network/swebhook", H.HandlerFunc(ns.HandleListWebhooks))
	srv.POST("/network/webhook", H.HandlerFunc(ns.HandleAddWebhook))
	srv.DELETE("/network/webhook", H.HandlerFunc(ns.HandleRemoveWebhook))

	srv.GET("/status", indexFunc(map[string]sensor{
		"brightness": bs,
		"motion": ms,
		"lights": ls,
		"appletv": atv,
		"sonos": sonos,
		"network": ns,
	}))
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
