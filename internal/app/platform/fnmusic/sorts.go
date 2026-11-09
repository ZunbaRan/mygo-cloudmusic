package fnmusic

import (
	backend "netease-music-mygo/internal/platform/fnmusic"

	"github.com/egoist/mygo/ui"
)

type sortField struct {
	field string
	label string
}

type sortControl struct {
	policy  backend.SortPolicy
	fields  []sortField
	current string
	onSort  func(string)
}

func (s sortControl) normalized() (sortField, string) {
	value := s.policy.Normalize(s.current)
	field, order, _ := backend.SplitSort(value)
	for _, option := range s.fields {
		if option.field == field {
			return option, order
		}
	}
	defField, defOrder, _ := backend.SplitSort(s.policy.Default)
	return sortField{field: defField, label: defField}, defOrder
}

func (s sortControl) button(c *ui.Context, key string) ui.Element {
	field, order := s.normalized()
	arrow := "↑"
	if order == backend.SortOrderDesc {
		arrow = "↓"
	}
	b := ui.ButtonBase(c.Key(key)).Height(30).PaddingX(12).Radius(999).Gap(6).AlignItems(ui.Center).FocusRing(false).Label("Sort")
	if b.Hovered() {
		b.Background(c.Theme().Text.Alpha(0.06))
	}
	b.Children(func() {
		ui.Text(c, "Sort · "+field.label+" "+arrow).FontSize(12.5).TextColor(c.Theme().TextMuted)
	})
	b.Menu(func(m *ui.Menu) {
		for _, option := range s.fields {
			if m.Item(option.label).Checked(option.field == field.field).Chosen() && option.field != field.field {
				_, defaultOrder, _ := backend.SplitSort(s.policy.Default)
				s.onSort(backend.SortValue(option.field, defaultOrder))
			}
		}
		m.Separator()
		for _, choice := range []struct {
			order string
			label string
		}{
			{backend.SortOrderAsc, "Ascending ↑"},
			{backend.SortOrderDesc, "Descending ↓"},
		} {
			if m.Item(choice.label).Checked(order == choice.order).Chosen() && order != choice.order {
				s.onSort(backend.SortValue(field.field, choice.order))
			}
		}
	})
	return b
}

var (
	librarySortFields = []sortField{
		{backend.SortFieldCreatedAt, "Date Added"},
		{backend.SortFieldTitle, "Song Title"},
	}
	favoriteSortFields = []sortField{
		{backend.SortFieldFavoriteAt, "Date Favorited"},
		{backend.SortFieldTitle, "Song Title"},
	}
	albumSortFields = []sortField{
		{backend.SortFieldAlbumAddedAt, "Date Updated"},
		{backend.SortFieldReleaseYear, "Release Year"},
		{backend.SortFieldName, "Album"},
		{backend.SortFieldArtistName, "Artist"},
		{backend.SortFieldTrackCount, "Songs"},
	}
	albumTrackSortFields = []sortField{
		{backend.SortFieldTrackNo, "Track No."},
		{backend.SortFieldTitle, "Title"},
	}
	artistSortFields = []sortField{
		{backend.SortFieldName, "Artist"},
		{backend.SortFieldTrackCount, "Songs"},
		{backend.SortFieldAlbumCount, "Albums"},
	}
	artistTrackSortFields = []sortField{
		{backend.SortFieldCreatedAt, "Date Added"},
	}
	artistAlbumSortFields = []sortField{
		{backend.SortFieldAlbumAddedAt, "Date Updated"},
		{backend.SortFieldReleaseYear, "Release Year"},
		{backend.SortFieldName, "Title"},
		{backend.SortFieldTrackCount, "Songs"},
	}
	genreSortFields = []sortField{
		{backend.SortFieldName, "Genre"},
		{backend.SortFieldTrackCount, "Songs"},
	}
	genreTrackSortFields = []sortField{
		{backend.SortFieldCreatedAt, "Date Added"},
		{backend.SortFieldArtistName, "Artists"},
	}
)

func newSortControl(policy backend.SortPolicy, fields []sortField, current string, onSort func(string)) sortControl {
	return sortControl{policy: policy, fields: fields, current: current, onSort: onSort}
}

func (h *Home) listSort(l *List) (sortControl, bool) {
	if l == nil {
		return sortControl{}, false
	}
	switch l.Spec.Kind {
	case ListLibrary:
		return h.librarySort(), true
	case ListFavorites:
		return h.favoriteSort(), true
	case ListAlbum:
		return h.albumTrackSort(), true
	case ListArtist:
		return h.artistTrackSort(), true
	case ListGenre:
		return h.genreTrackSort(), true
	default:
		return sortControl{}, false
	}
}

func (h *Home) artistTabs(c *ui.Context, id int64, ref string, active int) {
	at := active
	if !h.host.Segmented(c, "fnmusic-artist-tabs", &at, "Songs", "Albums") {
		return
	}
	page := PagePlaylist
	if at == 1 {
		page = PageArtistAlbums
	}
	h.host.Navigate(Route{Page: page, ID: id, Kind: backend.KindArtist, Query: ref})
}

func (h *Home) librarySort() sortControl {
	c := h.state.Catalog()
	return newSortControl(backend.LibrarySortPolicy, librarySortFields, c.LibrarySort, func(v string) {
		c.LibrarySort = v
		h.rememberSort(sortKeyLibrary, v)
		h.host.Invalidate()
	})
}

func (h *Home) favoriteSort() sortControl {
	c := h.state.Catalog()
	return newSortControl(backend.FavoriteSortPolicy, favoriteSortFields, c.FavoriteSort, func(v string) {
		c.FavoriteSort = v
		h.rememberSort(sortKeyFavorites, v)
		h.host.Invalidate()
	})
}

func (h *Home) albumSort() sortControl {
	c := h.state.Catalog()
	return newSortControl(backend.AlbumListSortPolicy, albumSortFields, c.AlbumSort, func(v string) {
		c.AlbumSort = v
		if c.Albums != nil {
			c.Albums.Reload()
		}
		h.rememberSort(sortKeyAlbums, v)
		h.host.Invalidate()
	})
}

func (h *Home) albumTrackSort() sortControl {
	c := h.state.Catalog()
	return newSortControl(backend.AlbumTrackSortPolicy, albumTrackSortFields, c.AlbumDetailSort, func(v string) {
		c.AlbumDetailSort = v
		h.rememberSort(sortKeyAlbumDetail, v)
		h.host.Invalidate()
	})
}

func (h *Home) artistSort() sortControl {
	c := h.state.Catalog()
	return newSortControl(backend.ArtistListSortPolicy, artistSortFields, c.ArtistSort, func(v string) {
		c.ArtistSort = v
		if c.Artists != nil {
			c.Artists.Reload()
		}
		h.rememberSort(sortKeyArtists, v)
		h.host.Invalidate()
	})
}

func (h *Home) artistTrackSort() sortControl {
	c := h.state.Catalog()
	return newSortControl(backend.ArtistTrackSortPolicy, artistTrackSortFields, c.ArtistTrackSort, func(v string) {
		c.ArtistTrackSort = v
		h.rememberSort(sortKeyArtistTracks, v)
		h.host.Invalidate()
	})
}

func (h *Home) artistAlbumSort(id int64) sortControl {
	c := h.state.Catalog()
	return newSortControl(backend.ArtistAlbumSortPolicy, artistAlbumSortFields, c.ArtistAlbumSort, func(v string) {
		c.ArtistAlbumSort = v
		if p := c.ArtistAlbums[id]; p != nil {
			p.Reload()
		}
		h.rememberSort(sortKeyArtistAlbums, v)
		h.host.Invalidate()
	})
}

func (h *Home) genreSort() sortControl {
	c := h.state.Catalog()
	return newSortControl(backend.GenreListSortPolicy, genreSortFields, c.GenreSort, func(v string) {
		c.GenreSort = v
		if c.Genres != nil {
			c.Genres.Reload()
		}
		h.rememberSort(sortKeyGenres, v)
		h.host.Invalidate()
	})
}

func (h *Home) genreTrackSort() sortControl {
	c := h.state.Catalog()
	return newSortControl(backend.GenreTrackSortPolicy, genreTrackSortFields, c.GenreTrackSort, func(v string) {
		c.GenreTrackSort = v
		h.rememberSort(sortKeyGenreTracks, v)
		h.host.Invalidate()
	})
}
