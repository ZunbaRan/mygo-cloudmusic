package fnmusic

import (
	"context"
	"netease-music-mygo/internal/netease"
	backend "netease-music-mygo/internal/platform/fnmusic"

	"github.com/egoist/mygo/ui"
)

func (h *Home) playlistMenu(m *ui.Menu, pl netease.Playlist) {
	if m.Item("Rename…").Chosen() {
		h.RequestRename(pl)
	}
	if m.Item("Change Cover…").Chosen() {
		h.RequestChangeCover(pl)
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
			gen := h.state.Generation()
			go func() {
				err := h.state.RemoveFromHistory(context.Background(), *tr)
				h.postCurrent(gen, func() {
					if err != nil {
						h.host.Toast("Couldn't remove. Please try again later")
						return
					}
					h.state.Catalog().Refresh(backend.RecentID)
					h.host.Invalidate()
				})
			}()
		}
	}
	if pl, ok := h.state.Playlist(pid); ok && h.route.Page == PagePlaylist && h.route.ID == pid {
		if m.Item("Remove from Playlist").Chosen() {
			h.RequestRemoveTrack(pl, *tr)
		}
	}
}

func (h *Home) listMenu(m *ui.Menu, l *List) {
	if l == nil {
		return
	}
	if pl, ok := h.state.Playlist(l.ID); ok && l.Spec.Kind == ListPlaylist {
		h.playlistMenu(m, pl)
		m.Separator()
	}
	if l.Spec.Kind == ListAlbum || l.Spec.Kind == ListArtist || l.Spec.Kind == ListGenre {
		m.Submenu("Add All to Playlist", func(sub *ui.Menu) {
			if sub.Item("New Playlist…").Chosen() {
				h.RequestNewPlaylistFrom(l)
			}
			for _, pl := range h.state.Playlists() {
				if sub.Item(pl.Name).Chosen() {
					h.RequestAddList(pl, l)
				}
			}
		})
		m.Separator()
	}
	if m.Item("Refresh").Chosen() {
		l.Refresh()
		h.host.Invalidate()
	}
}

func (h *Home) RequestNewPlaylistFrom(l *List) {
	h.nameOpen, h.name, h.rename, h.delete, h.deleteOpen, h.addTracks, h.addList = true, "", nil, nil, false, nil, l
}

func (h *Home) RequestAddList(pl netease.Playlist, l *List) {
	if h == nil || h.state == nil || l == nil || l.Pages == nil {
		return
	}
	gen := h.state.Generation()
	l.Pages.Whole(func(ok bool) {
		if h.state.Generation() != gen {
			return
		}
		if !ok {
			h.host.Toast("Couldn't load the songs. Please try again later")
			return
		}
		tracks := make([]netease.Track, 0, l.Pages.Len())
		for i := 0; i < l.Pages.Len(); i++ {
			if tr := l.Pages.Peek(i); tr != nil {
				tracks = append(tracks, *tr)
			}
		}
		h.RequestAddTracks(pl, tracks)
	})
}
