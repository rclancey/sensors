package automator

import (
	"github.com/rclancey/events"
)

type Automator struct {
	events chan events.Event
	presence *Presence
	//sunrise *Sunrise
	//sunset *Sunset
	//motion *Motion
	//tv *TV
	//lights *Lights
}


