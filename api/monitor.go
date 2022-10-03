package api

import (
	"time"
)

type Monitorable interface {
	Check() (interface{}, error)
}

func Monitor(m Monitorable, interval time.Duration) quitFunc func() {
	quitch := make(chan bool, 1)
	quit := false
	quitFunc = func() {
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
