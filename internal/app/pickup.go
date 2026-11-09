package app

import (
	"fmt"
	"slices"
	"time"

	"netease-music-mygo/internal/netease"
	"netease-music-mygo/internal/player"
)

// Signed in, the listener finds what they played last, on their phone or
// wherever, waiting in the player: paused, with where they heard it under
// its title.

// recentPlays returns the listener's recent plays; tests replace it.
var recentPlays = netease.RecentPlays

// pickedUp is the song a login picked up, while it waits to play.
type pickedUp struct {
	id    int64
	heard string // where and when, as "12 min ago on phone"
}

// pickUp queues the songs the listener played last, paused, if nothing is
// in the player.
func (a *app) pickUp(uid int64) {
	go func() {
		plays, err := recentPlays(100)
		a.post(func() {
			if err != nil || a.profile == nil || a.profile.UserID != uid || a.pl.Current() != nil {
				return
			}
			songs, last, ok := lastStretch(plays, time.Now())
			if !ok {
				return
			}
			order := make([]int, len(songs))
			for i := range order {
				order[i] = i
			}
			a.pl.Restore(player.Session{Queue: songs, Order: order, At: len(songs) - 1})
			a.picked = pickedUp{id: last.Track.ID, heard: heardAt(last, time.Now())}
			where := "last time"
			if d := deviceName(last.Device); d != "" {
				where = "on " + d
			}
			a.toast = "Picked up " + last.Track.Name + ", played " + where + " — press Space to continue"
		})
	}()
}

// lastStretch returns the songs of the listener's last stretch of
// listening, in the order they played, and the last play of them: on one
// device, each song within half an hour of the next, the last in the past
// week. Songs they may not play are left out, and songs played twice show
// where they played last.
func lastStretch(plays []netease.Play, now time.Time) (songs []netease.Track, last netease.Play, ok bool) {
	if len(plays) == 0 || now.Sub(plays[0].At) > 7*24*time.Hour {
		return nil, netease.Play{}, false
	}
	seen := map[int64]bool{}
	for i, p := range plays {
		if i > 0 && (p.Device != plays[0].Device || plays[i-1].At.Sub(p.At) > 30*time.Minute) || len(songs) == 50 {
			break
		}
		if !p.Track.Playable() || seen[p.Track.ID] {
			continue
		}
		seen[p.Track.ID] = true
		if len(songs) == 0 {
			last = p
		}
		songs = append(songs, p.Track)
	}
	slices.Reverse(songs)
	return songs, last, len(songs) > 0
}

// heardAt says when and where a song was played.
func heardAt(p netease.Play, now time.Time) string {
	d := now.Sub(p.At)
	var when string
	switch {
	case d < 2*time.Minute:
		when = "just now"
	case d < time.Hour:
		when = fmt.Sprintf("%d min ago", int(d/time.Minute))
	case d < 24*time.Hour:
		when = fmt.Sprintf("%d hours ago", int(d/time.Hour))
	default:
		when = fmt.Sprintf("%d days ago", int(d/(24*time.Hour)))
	}
	if where := deviceName(p.Device); where != "" {
		return when + " on " + where
	}
	return when
}

// deviceName names the kind of device NetEase calls os, or "".
func deviceName(os string) string {
	switch os {
	case "iphone", "android", "ipad", "harmony":
		return "phone"
	case "osx", "pc", "mac", "windows", "linux":
		return "computer"
	case "web":
		return "web"
	}
	return ""
}
