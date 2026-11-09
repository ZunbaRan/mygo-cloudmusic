package songloft

import (
	"github.com/egoist/mygo/ui"
	"netease-music-mygo/internal/netease"
)

func (h *Home) playlistMenu(m *ui.Menu, pl netease.Playlist) {
	if m.Item("Play").Chosen() {
		if cat := h.state.Catalog(); cat != nil {
			l := cat.ListForRoute(pl.ID, "playlist", pl.Ref)
			PlayList(h.host.Player(), l, 0)
		}
	}
	if m.Item("Shuffle").Chosen() {
		if cat := h.state.Catalog(); cat != nil {
			l := cat.ListForRoute(pl.ID, "playlist", pl.Ref)
			ShuffleList(h.host.Player(), l, h.host.Post)
		}
	}
	m.Separator()
	if m.Item("Rename…").Chosen() {
		h.RequestRename(pl)
	}
	if m.Item("Delete Playlist").Chosen() {
		h.RequestDelete(pl)
	}
}
