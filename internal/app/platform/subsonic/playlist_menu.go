package subsonic

import (
	"github.com/egoist/mygo/ui"
	"netease-music-mygo/internal/netease"
	backend "netease-music-mygo/internal/platform/subsonic"
)

func (h *Home) playlistMenu(m *ui.Menu, pl netease.Playlist) {
	if m.Item("Rename…").Chosen() {
		h.RequestRename(pl)
	}
	m.Separator()
	if m.Item("Delete Playlist…").Chosen() {
		h.RequestDelete(pl)
	}
}

func (h *Home) trackMenu(m *ui.Menu, tr *netease.Track, pid int64) {
	if tr == nil {
		return
	}
	if m.Item("Copy Song Title").Chosen() {
		h.host.CopyText(tr.Name)
	}
	if m.Item("Copy Title and Artist").Chosen() {
		h.host.CopyText(tr.Name + " - " + tr.ArtistNames())
	}
	if h.state == nil || !h.state.Connected() {
		return
	}
	m.Separator()
	m.Submenu("Add to Playlist", func(sub *ui.Menu) {
		if sub.Item("New Playlist…").Chosen() {
			h.RequestNewPlaylist([]netease.Track{*tr})
		}
		for _, pl := range h.state.Playlists() {
			if pl.ID != pid && sub.Item(pl.Name).Chosen() {
				h.RequestAddTracks(pl, []netease.Track{*tr})
			}
		}
	})
	if pid == 0 && h.route.Page == PagePlaylist {
		pid = h.route.ID
	}
	if pid == backend.RecentID && h.route.Page == PagePlaylist {
		if m.Item("Remove from Recently Played").Chosen() {
			h.RemoveRecent(tr.ID)
		}
	}
	if pid != 0 && h.route.Page == PagePlaylist && h.route.ID == pid {
		var currentPl *netease.Playlist
		for _, pl := range h.state.Playlists() {
			if pl.ID == pid {
				currentPl = &pl
				break
			}
		}
		if currentPl != nil {
			if m.Item("Remove from Playlist").Chosen() {
				h.RequestRemoveTrack(*currentPl, *tr)
			}
		}
	}
}
