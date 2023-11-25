package sonos

import (
	"time"
)

type State struct {
    State      *string   `json:"state,omitempty"`
    Speed      *float64  `json:"speed,omitempty"`
    Volume     *int      `json:"volume,omitempty"`
    Mute       *bool     `json:"mute,omitempty"`
    PlayMode   *int      `json:"mode,omitempty"`
    Tracks     []*Track  `json:"tracks,omitempty"`
    Index      *int      `json:"index,omitempty"`
    Duration   *int      `json:"duration,omitempty"`
    Time       *int      `json:"time,omitempty"`
    Error      error     `json:"error,omitempty"`
    LastUpdate time.Time `json:"now"`
}

func (ss *State) Clone() *State {
    clone := *ss
    return &clone
}

func (s *State) ApplyUpdate(update *State) *State {
	out := s.Clone()
	if update.State != nil {
		out.State = update.State
	}
	if update.Speed != nil {
		out.Speed = update.Speed
	}
	if update.Volume != nil {
		out.Volume = update.Volume
	}
	if update.Mute != nil {
		out.Mute = update.Mute
	}
	if update.PlayMode != nil {
		out.PlayMode = update.PlayMode
	}
	if update.Tracks != nil {
		out.Tracks = update.Tracks
	}
	if update.Index != nil {
		out.Index = update.Index
	}
	if update.Duration != nil {
		out.Duration = update.Duration
	}
	if update.Time != nil {
		out.Time = update.Time
	}
	out.LastUpdate = update.LastUpdate
	return out
}

func (s *State) Diff(update *State) *State {
	diff := &State{LastUpdate: update.LastUpdate}
	if update.State != nil {
		if s.State == nil || *s.State != *update.State {
			diff.State = update.State
		}
	}
	if update.Speed != nil {
		if s.Speed == nil || *s.Speed != *update.Speed {
			diff.Speed = update.Speed
		}
	}
	if update.Volume != nil {
		if s.Volume == nil || *s.Volume != *update.Volume {
			diff.Volume = update.Volume
		}
	}
	if update.Mute != nil {
		if s.Mute == nil || *s.Mute != *update.Mute {
			diff.Mute = update.Mute
		}
	}
	if update.PlayMode != nil {
		if s.PlayMode == nil || *s.PlayMode != *update.PlayMode {
			diff.PlayMode = update.PlayMode
		}
	}
	if update.Index != nil {
		if s.Index == nil || *s.Index != *update.Index {
			diff.Index = update.Index
		}
	}
	if update.Duration != nil {
		if s.Duration == nil || *s.Duration != *update.Duration {
			diff.Duration = update.Duration
		}
	}
	if update.Time != nil {
		if s.Time == nil || *s.Time != *update.Time {
			diff.Time = update.Time
		}
	}
	if update.Tracks != nil {
		if len(s.Tracks) != len(update.Tracks) {
			diff.Tracks = update.Tracks
		} else {
			for i, tr := range update.Tracks {
				if s.Tracks[i].URI != tr.URI {
					diff.Tracks = update.Tracks
					break
				}
			}
		}
	}
	return diff
}
