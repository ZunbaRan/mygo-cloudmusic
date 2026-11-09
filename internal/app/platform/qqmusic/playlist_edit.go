package qqmusic

import (
	"strings"

	"netease-music-mygo/internal/netease"
	backend "netease-music-mygo/internal/platform/qqmusic"

	"github.com/egoist/mygo/ui"
)

func (h *Home) openNewPlaylist() {
	h.nameOpen, h.name, h.editing, h.pending, h.nameErr = true, "", nil, nil, ""
}

func (h *Home) requestAddToPlaylist(plID int64, tr netease.Track) {
	go func() {
		err := backend.AddToPlaylist(plID, tr.ID)
		h.Post(func() {
			if err != nil {
				h.host.Toast("Failed to add to playlist. Please try again later")
				return
			}
			// The list's cached songs are now one short of what it has.
			h.mu.Lock()
			delete(h.playlists, plID)
			h.mu.Unlock()
			if h.route.Page == PagePlaylist && h.route.ID == plID {
				h.loadPlaylist(plID)
			}
			h.host.Toast("Added to playlist")
			h.host.Invalidate()
		})
	}()
}

func (h *Home) playlistMenu(m *ui.Menu, pl netease.Playlist) {
	if m.Item("Rename…").Chosen() {
		h.nameOpen, h.name, h.editing, h.pending, h.nameErr = true, pl.Name, &pl, nil, ""
		return
	}
	if m.Item("Delete Playlist…").Chosen() {
		h.confirmRemove(pl, false)
	}
}

// confirmRemove asks before deleting a made playlist or unfavouriting a
// saved one, as the reference does.
func (h *Home) confirmRemove(pl netease.Playlist, saved bool) {
	h.removing, h.removeSaved = &pl, saved
	h.removeOpen, h.removeErr, h.removeBusy = true, "", false
}

func (h *Home) removeDialog(c *ui.Context) {
	if !h.removeOpen || h.removing == nil {
		return
	}
	t := c.Theme()
	pl := *h.removing
	title, action, desc := "Delete Playlist", "Delete", "This can't be undone. The songs in the playlist won't be deleted."
	if h.removeSaved {
		title, action, desc = "Unsave", "Unsave", "This playlist will be removed from your saved playlists."
	}
	ui.DialogBase(c.Key("qq-remove-playlist"), &h.removeOpen, func(backdrop, panel ui.Element) {
		backdrop.Background(ui.RGBA(0, 0, 0, 0.35))
		panel.Width(360).Padding(24).Radius(18).Gap(14).Background(t.Background).Border(1, t.Border).
			Shadow(0, 24, 60, -12, ui.RGBA(0, 0, 0, 0.4))
		ui.Text(c, title).FontSize(17).FontWeight(700)
		ui.Text(c, "\""+pl.Name+"\"").FontSize(13.5)
		ui.Text(c, desc).FontSize(12.5).TextColor(t.TextMuted).LineHeight(1.45)
		if h.removeBusy {
			ui.Text(c, "Working…").FontSize(12.5).TextColor(t.TextMuted)
			h.host.Bone(c, 2).Width(312).Height(4)
		}
		if h.removeErr != "" {
			ui.Text(c, h.removeErr).FontSize(12.5).TextColor(t.Accent)
		}
		ui.Row(c).Gap(10).Justify(ui.End).Children(func() {
			if h.host.PillButton(c, "qq-remove-cancel", nil, "Cancel", false).Clicked() {
				h.removeOpen = false
			}
			b := h.host.PillButton(c, "qq-remove-confirm", nil, action, true)
			if h.removeBusy {
				b.Disabled(true)
			}
			if b.Clicked() {
				h.finishRemove(pl, h.removeSaved)
			}
		})
	})
}

func (h *Home) finishRemove(pl netease.Playlist, saved bool) {
	if h.removeBusy {
		return
	}
	h.removeBusy, h.removeErr = true, ""
	go func() {
		var err error
		if saved {
			err = backend.UnfavoritePlaylist(pl.ID)
		} else {
			err = backend.DeletePlaylist(pl.ID)
		}
		h.Post(func() {
			h.removeBusy = false
			if err != nil {
				h.removeErr = "Something went wrong. Please try again later"
				h.host.Invalidate()
				return
			}
			h.removeOpen = false
			lists := h.state.Playlists()
			out := lists[:0]
			for _, item := range lists {
				if item.ID != pl.ID {
					out = append(out, item)
				}
			}
			h.state.SetPlaylists(out)
			h.mu.Lock()
			delete(h.playlists, pl.ID)
			h.mu.Unlock()
			if h.route.Page == PagePlaylist && h.route.ID == pl.ID {
				h.Open(Route{Page: PageHome})
			}
			if saved {
				h.host.Toast("Removed from collection")
			} else {
				h.host.Toast("Playlist deleted")
			}
			h.host.Invalidate()
		})
	}()
}

func (h *Home) playlistDialog(c *ui.Context) {
	if !h.nameOpen {
		return
	}
	t := c.Theme()
	ui.DialogBase(c.Key("qq-playlist-name"), &h.nameOpen, func(backdrop, panel ui.Element) {
		backdrop.Background(ui.RGBA(0, 0, 0, 0.35))
		panel.Width(340).Padding(24).Radius(18).Gap(14).Background(t.Background).Border(1, t.Border).
			Shadow(0, 24, 60, -12, ui.RGBA(0, 0, 0, 0.4))
		title := "New Playlist"
		if h.editing != nil {
			title = "Rename Playlist"
		}
		ui.Text(c, title).FontSize(17).FontWeight(700)
		in := ui.TextInput(c.Key("qq-playlist-name-input"), &h.name).Placeholder("Playlist Name").Label("Playlist Name").FontSize(14).AutoFocus()
		if h.nameErr != "" {
			ui.Text(c, h.nameErr).FontSize(12.5).TextColor(t.Accent)
		}
		ui.Row(c).Gap(10).Justify(ui.End).Children(func() {
			if h.host.PillButton(c, "qq-playlist-name-cancel", nil, "Cancel", false).Clicked() {
				h.nameOpen = false
			}
			save := h.host.PillButton(c, "qq-playlist-name-save", nil, "Save", true)
			if save.Clicked() || in.Submitted() {
				h.savePlaylistName()
			}
		})
	})
}

func (h *Home) savePlaylistName() {
	name := strings.TrimSpace(h.name)
	if name == "" {
		h.nameErr = "Enter a playlist name"
		return
	}
	h.nameErr = ""
	if h.editing != nil {
		pl := *h.editing
		go func() {
			err := backend.RenamePlaylist(pl.ID, name)
			h.Post(func() {
				if err != nil {
					h.nameErr = "Couldn't rename. Please try again later"
					h.host.Invalidate()
					return
				}
				lists := h.state.Playlists()
				for i := range lists {
					if lists[i].ID == pl.ID {
						lists[i].Name = name
					}
				}
				h.state.SetPlaylists(lists)
				h.nameOpen = false
				h.host.Invalidate()
			})
		}()
		return
	}
	pending := h.pending
	go func() {
		pl, err := backend.CreatePlaylist(name)
		if err == nil && len(pending) > 0 {
			ids := make([]int64, len(pending))
			for i := range pending {
				ids[i] = pending[i].ID
			}
			err = backend.AddToPlaylist(pl.ID, ids...)
		}
		h.Post(func() {
			if err != nil || pl == nil {
				h.nameErr = "Couldn't create the playlist. Please try again later"
				h.host.Invalidate()
				return
			}
			lists := h.state.Playlists()
			h.state.SetPlaylists(append([]netease.Playlist{playlist(*pl)}, lists...))
			h.nameOpen = false
			h.host.Invalidate()
		})
	}()
}
