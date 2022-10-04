package api

import (
	"time"
)

type Monitorable interface {
	Check() (interface{}, error)
}

type QuitFunc func()

func Monitor(m Monitorable, interval time.Duration) QuitFunc {
	quitch := make(chan bool, 1)
	quit := false
	quitFunc := func() {
		if !quit {
			quit = true
			close(quitch)
		}
	}
	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-quitch:
				return
			case <-ticker.C:
				m.Check()
			}
		}
	}()
	return quitFunc
}
