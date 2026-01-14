package verdant

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/task"
)

// Verdant Dew charging system
type Handler struct {
	f      *int
	events event.Eventter
	tasks  task.Tasker
	log    glog.Logger

	baseDur   int // in frames
	gainBonus float64

	expiryFrame   int
	charging      bool
	count         int     // 0..3
	verdantPart   float64 // 0..450
	tickScheduled bool
}

func New(f *int, events event.Eventter, tasks task.Tasker, log glog.Logger) *Handler {
	h := &Handler{
		f:       f,
		events:  events,
		tasks:   tasks,
		log:     log,
		baseDur: int(2.5 * 60), // 2.5s
		count:   0,
	}

	// subscribe to Lunar Bloom events
	if events != nil {
		events.Subscribe(event.OnLunarBloom, h.onLunarBloom, "verdant-on-lunarbloom")
	}

	return h
}

func (h *Handler) onLunarBloom(args ...interface{}) bool {
	// args: target, *AttackEvent
	// we don't need attacker here, just start/reset charging for the party
	h.StartCharge(h.baseDur)
	return false
}

// StartCharge begins or resets the charging timer
func (h *Handler) StartCharge(dur int) {
	if h.f == nil {
		return
	}
	if dur < 0 {
		dur = 0
	}
	h.expiryFrame = *h.f + dur
	h.charging = true

	// start ticking loop once per frame while charging
	if h.tasks != nil && !h.tickScheduled {
		h.tickScheduled = true
		var tick func()
		tick = func() {
			// stop if not charging or f is nil
			if h.f == nil {
				h.tickScheduled = false
				return
			}
			// if current frame >= expiryFrame, stop charging and don't schedule further
			if *h.f >= h.expiryFrame {
				h.charging = false
				h.tickScheduled = false
				return
			}

			// accumulate verdant part per frame
			add := 6.0 * (1.0 + h.gainBonus)
			h.verdantPart += add
			if h.verdantPart > 450 {
				h.verdantPart = 450
			}
			// update dew count based on thresholds of 150
			newCount := int(h.verdantPart) / 150
			if newCount > 3 {
				newCount = 3
			}
			if newCount != h.count {
				gained := newCount - h.count
				h.count = newCount
				if h.log != nil {
					h.log.NewEvent("verdant dew granted", glog.LogDebugEvent, -1).
						Write("count", h.count).
						Write("gained", gained)
				}
				if h.events != nil && gained > 0 {
					h.events.Emit(event.OnVerdantDewGain, gained, h.count)
				}
			}

			// schedule next frame
			h.tasks.Add(tick, 6)
		}

		// schedule first tick next frame
		h.tasks.Add(tick, 1)
	}
}

// SetGainBonus sets the verdant gain bonus used when accumulating per frame.
// bonus is fractional (0.2 = +20%).
func (h *Handler) SetGainBonus(b float64) {
	h.gainBonus = b
}

func (h *Handler) GetGainBonus() float64 { return h.gainBonus }
func (h *Handler) Count() int            { return h.count }
func (h *Handler) IsCharging() bool      { return h.charging }
func (h *Handler) RemainingFrames() int {
	if !h.charging || h.f == nil {
		return 0
	}
	rem := h.expiryFrame - *h.f
	if rem < 0 {
		return 0
	}
	return rem
}

// GetPart returns current verdantPart (0..450)
func (h *Handler) GetPart() float64 { return h.verdantPart }

// Consume consumes up to n Verdant Dew, removing 150 verdantPart per dew consumed.
// returns actual consumed count
func (h *Handler) Consume(n int) int {
	if n <= 0 {
		return 0
	}
	maxDews := int(h.verdantPart) / 150
	if maxDews == 0 {
		return 0
	}
	toConsume := n
	if toConsume > maxDews {
		toConsume = maxDews
	}
	h.verdantPart -= float64(toConsume * 150)
	if h.verdantPart < 0 {
		h.verdantPart = 0
	}
	// update count
	h.count = int(h.verdantPart) / 150
	return toConsume
}

