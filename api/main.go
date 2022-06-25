package api

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/rclancey/httpserver/v2"
	"github.com/rclancey/logging"
)

func APIMain() {
	var errlog *logging.Logger
	var srv *httpserver.Server
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

func startup() (*logging.Logger, *httpserver.Server, error) {
	var err error
	cfg, err = httpserver.Configure()
	if err != nil {
		log.Println("error configuring server:", err)
		return safeMode(nil, errors.Wrap(err, ErrInvalidConfiguration.Error()))
	}
	if cfg == nil {
		log.Println("no configuration found")
		return safeMode(nil, ErrNoConfiguration)
	}
	errlog, err := cfg.Logging.ErrorLogger()
	if err != nil {
		log.Println("error configuring logging:", err)
		return safeMode(cfg, errors.Wrap(err, ErrLoggingError.Error()))
	}
	colorizeLogger(errlog)

	srv, err := httpserver.NewServer(cfg)
	if err != nil {
		log.Fatalln("can't create server:", err)
	}

	ls, err := NewLightSensor(cfg)
	if err != nil {
		log.Fatalln("can't create light sensor:", err)
	}

	ms, err := NewMotionSensor(cfg.ServerConfig)
	if err != nil {
		log.Fatalln("can't create motion sensor:", err)
	}

	srv.RegisterOnShutdown(func() {
		log.Println("cleanup globals on shutdown")
		ls.webhooks.Stop()
		ms.webhooks.Stop()
	})

	errlog.Infoln("server starting...")
	srv.GET("/light/status", ls.HandleRead)
	srv.POST("/light/webhook", ls.HandleAddWebhook)
	srv.DELETE("/light/webhook", ls.HandleRemoveWebhook)
	srv.GET("/motion/status", ms.HandleRead)
	srv.POST("/motion/webhook", ms.HandleAddWebhook)
	srv.DELETE("/motion/webhook", ms.HandleRemoveWebhook)
	errlog.Infoln("server ready")

	return errlog, srv, nil
}
