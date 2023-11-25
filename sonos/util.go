package sonos

import (
	"errors"
	"fmt"
	"net"
	"strconv"
)

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

func recoverError() error {
	if r := recover(); r != nil {
		switch xr := r.(type) {
		case error:
				return xr
		case string:
			return errors.New(xr)
		case fmt.Stringer:
			return errors.New(xr.String())
		default:
			return fmt.Errorf("error communicating with sonos: %#v", r)
		}
	}
	return nil
}
