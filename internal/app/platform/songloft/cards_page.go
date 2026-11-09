package songloft

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	appplatform "netease-music-mygo/internal/app/platform"
	"netease-music-mygo/internal/netease"
	"netease-music-mygo/internal/platform/paging"
	backend "netease-music-mygo/internal/platform/songloft"
)

func (h *Home) cardsHeader(c *ui.Context, title string, total int) {
	t := c.Theme()
	ui.Column(c).Gap(4).Margin(0, 0, 18, 0).Children(func() {
		ui.Text(c, title).FontSize(30).FontWeight(700).LetterSpacing(-0.6)
		if total >= 0 {
			ui.Text(c, fmt.Sprintf("%d in total", total)).FontSize(13).TextColor(t.TextMuted)
		}
	})
}

func (h *Home) cardsHeaderSkeleton(c *ui.Context) {
	ui.Column(c).Gap(4).Margin(0, 0, 18, 0).Children(func() {
		h.host.TextBone(c, 30, 1.2).WidthPercent(32)
		h.host.TextBone(c, 13, 1.3).Width(52)
	})
}

func (h *Home) findCardMeta(id int64) (string, string) {
	if meta, ok := backend.LookupMeta(id); ok && meta.Ref != "" {
		return meta.Kind, meta.Ref
	}
	return "album", ""
}

func (h *Home) openCard(id int64) {
	if h == nil || h.host == nil || h.state == nil {
		return
	}
	kind, ref := h.findCardMeta(id)
	h.host.Navigate(Route{Page: PagePlaylist, ID: id, Kind: kind, Query: ref})
}

func (h *Home) playCard(id int64) {
	if h == nil || h.state == nil || h.host == nil {
		return
	}
	catalog := h.state.Catalog()
	kind, ref := h.findCardMeta(id)
	l := catalog.ListForRoute(id, kind, ref)
	if l == nil || l.Pages == nil {
		return
	}
	l.Pages.When(0, func(ok bool) {
		if ok {
			PlayList(h.host.Player(), l, 0)
		}
	})
}

func (h *Home) cardsPage(c *ui.Context) {
	catalog := h.state.Catalog()
	switch h.route.Page {
	case PageSearch:
		searchList := catalog.SearchList(h.route.Query)
		var songs *appplatform.TrackList
		if searchList != nil {
			songs = searchList.View()
		}
		h.host.RenderSearch(c, appplatform.SearchResults{
			Query:   h.route.Query,
			Songs:   songs,
			Albums:  catalog.SearchAlbums(h.route.Query),
			Artists: catalog.SearchArtists(h.route.Query),
		}, SearchView{
			Play: func(i int) {
				if searchList != nil {
					PlayList(h.host.Player(), searchList, i)
				}
			},
			PlayCard: h.playCard,
			TrackMenu: func(m *ui.Menu, tr *netease.Track, index int64) {
				if tr != nil {
					h.trackMenu(m, *tr, searchList)
				}
			},
			Liked:    h.state.Liked,
			Favorite: h.RequestFavorite,
			OpenList: h.openCard,
			OpenCards: func(kind string) {
				page := PageSearchArtists
				if kind == "albums" {
					page = PageSearchAlbums
				}
				h.host.Navigate(Route{Page: page, Query: h.route.Query})
			},
		})

	case PageSearchAlbums:
		cards := catalog.SearchAlbums(h.route.Query)
		h.host.RenderCards(c, cards, CardView{
			Title:          "Album · “" + h.route.Query + "”",
			Count:          cardsTotal(cards),
			OnOpen:         h.openCard,
			OnPlay:         h.playCard,
			Header:         func(c *ui.Context) { h.cardsHeader(c, "Album · “"+h.route.Query+"”", cardsTotal(cards)) },
			HeaderSkeleton: h.cardsHeaderSkeleton,
		})

	case PageSearchArtists:
		cards := catalog.SearchArtists(h.route.Query)
		h.host.RenderCards(c, cards, CardView{
			Title:          "Artist · “" + h.route.Query + "”",
			Count:          cardsTotal(cards),
			OnOpen:         h.openCard,
			OnPlay:         h.playCard,
			Round:          true,
			Header:         func(c *ui.Context) { h.cardsHeader(c, "Artist · “"+h.route.Query+"”", cardsTotal(cards)) },
			HeaderSkeleton: h.cardsHeaderSkeleton,
		})

	case PageAlbums:
		cards := catalog.AlbumsPager()
		h.host.RenderCards(c, cards, CardView{
			Title:   "Albums",
			Count:   cardsTotal(cards),
			OnOpen:  h.openCard,
			OnPlay:  h.playCard,
			Sort:    albumSortChoices(),
			Current: catalog.AlbumSort,
			OnSort: func(sort string) {
				catalog.SetAlbumSort(sort)
				h.host.Invalidate()
			},
			Header:         func(c *ui.Context) { h.cardsHeader(c, "Albums", cardsTotal(cards)) },
			HeaderSkeleton: h.cardsHeaderSkeleton,
		})

	case PageArtists:
		cards := catalog.ArtistsPager()
		h.host.RenderCards(c, cards, CardView{
			Title:   "Artists",
			Count:   cardsTotal(cards),
			OnOpen:  h.openCard,
			OnPlay:  h.playCard,
			Round:   true,
			Sort:    artistSortChoices(),
			Current: catalog.ArtistSort,
			OnSort: func(sort string) {
				catalog.SetArtistSort(sort)
				h.host.Invalidate()
			},
			Header:         func(c *ui.Context) { h.cardsHeader(c, "Artists", cardsTotal(cards)) },
			HeaderSkeleton: h.cardsHeaderSkeleton,
		})
	}
}

func cardsTotal(p *paging.Pager[netease.Playlist]) int {
	if p == nil {
		return 0
	}
	return p.Total()
}

func albumSortChoices() []appplatform.SortChoice {
	return []appplatform.SortChoice{
		{Label: "Most Songs", Value: "count:desc"},
		{Label: "Name A-Z", Value: "name:asc"},
		{Label: "Name Z-A", Value: "name:desc"},
		{Label: "Fewest Songs", Value: "count:asc"},
	}
}

func artistSortChoices() []appplatform.SortChoice {
	return []appplatform.SortChoice{
		{Label: "Most Songs", Value: "count:desc"},
		{Label: "Name A-Z", Value: "name:asc"},
		{Label: "Name Z-A", Value: "name:desc"},
		{Label: "Fewest Songs", Value: "count:asc"},
	}
}
