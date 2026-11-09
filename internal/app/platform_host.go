package app

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	appplatform "netease-music-mygo/internal/app/platform"
	"netease-music-mygo/internal/netease"
	"netease-music-mygo/internal/platform"
	"netease-music-mygo/internal/platform/paging"
)

// Generic host bridge for platform-owned homes.
//
// This file exposes shared application capabilities such as navigation,
// playback, common controls and list rendering. It is newly added adapter
// infrastructure, so it does not need to mirror upstream file boundaries.
// Keep it provider-neutral: concrete platform behavior belongs below the
// corresponding platform adapter directory.

// platformHost is the narrow adapter between a platform-owned home and the
// shared application. It deliberately contains no platform decisions: the
// platform package owns navigation and data, while this adapter only exposes
// existing player and renderer capabilities.
type platformHost struct{ a *app }

// FetchArtwork is the single image download seam used by the shared artwork
// cache. Platform-owned URLs can be routed through the active provider;
// ordinary URLs retain the original fetch path.
func (h platformHost) FetchArtwork(raw string) ([]byte, error) {
	if h.a != nil {
		// Route by resource ownership rather than the currently selected UI.
		// An artwork request can finish after a platform switch, and must still
		// use the adapter that understands its URL scheme.
		adapters := make([]appplatform.Adapter, 0, len(h.a.platforms))
		if active := h.a.activePlatform(); active != nil {
			adapters = append(adapters, active)
		}
		for id, adapter := range h.a.platforms {
			if h.a.platformID != id {
				adapters = append(adapters, adapter)
			}
		}
		for _, adapter := range adapters {
			if adapter == nil {
				continue
			}
			rt := adapter.Runtime()
			if rt == nil {
				continue
			}
			data, handled, err := rt.FetchArtwork(context.Background(), raw, 600)
			if handled {
				return data, err
			}
		}
	}
	return netease.Fetch(raw)
}

func (h platformHost) Post(fn func()) { h.a.post(fn) }
func (h platformHost) Invalidate()    { h.a.redraw() }
func (h platformHost) ShowMain()      { h.a.showMain() }
func (h platformHost) Toast(msg string) {
	h.a.toast = msg
	h.a.redraw()
}
func (h platformHost) Player() appplatform.PlayerPort {
	if h.a == nil || h.a.pl == nil {
		return nil
	}
	return sharedPlayer{h.a.pl}
}

func (h platformHost) Navigate(r appplatform.Route) {
	home := h.a.activeHome()
	if home == nil {
		return
	}
	home.Open(r)
	h.a.nowPlaying = false
	h.a.redraw()
}

func (h platformHost) PlatformSearchField(c *ui.Context, placeholder string, submit func(string)) {
	home := h.a.activeHome()
	if home == nil || home.SearchText() == nil {
		return
	}
	field := ui.Row(c).Height(32).PaddingX(10).Gap(7).Radius(9).AlignItems(ui.Center).Margin(0, 0, 14, 0)
	if c.Focused(h.a.searchBox) {
		field.Background(h.a.pal.bg).Shadow(0, 0, 0, 3, h.a.pal.accent.Alpha(0.25)).Border(1, h.a.pal.accent.Alpha(0.6))
	} else {
		field.Background(h.a.pal.hover).Border(1, h.a.pal.hairline)
	}
	field.Children(func() {
		ui.Icon(c, iconSearch).Size(14, 14).TextColor(h.a.pal.muted)
		// Use the same base widget as the upstream search field. A plain
		// TextInput looks close, but does not keep the combobox focus,
		// keyboard and popup state that gives the search field its behavior.
		cb := ui.ComboboxBase(c.Key("search"), home.SearchText())
		in := cb.Input.Bind(&h.a.searchBox).Placeholder(placeholder).Label("Search").FontSize(13).Grow(1).MinWidth(0)
		q := strings.TrimSpace(*home.SearchText())
		if q == "" && len(h.a.settings.SearchHistory) == 0 {
			cb.SetOpen(false)
		}
		cb.Popup(func(panel ui.Element) {
			panel.AttachTo(field, ui.AnchorBottomStart, ui.AnchorTopStart).
				Width(360).Padding(6).Margin(8, 0, 0, 0).Radius(12).
				Background(h.a.pal.bg).Border(1, h.a.pal.hairline).
				Shadow(0, 8, 24, -4, ui.RGBA(0, 0, 0, 0.2))
			option := func(key, label string) ui.Element {
				row := cb.Item(key).MinHeight(36).Padding(6, 10).Gap(10).Radius(7).AlignItems(ui.Center).Label(label)
				if row.Highlighted() {
					row.Background(h.a.pal.hover)
				}
				return row
			}
			if q == "" {
				panel.Label("Search History")
				ui.Row(c).Padding(2, 4, 2, 10).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "Search History").FontSize(12).TextColor(h.a.pal.muted).Grow(1)
					iconButton(c.Key("platform-clear-search-history"), iconTrash, 14, h.a.pal.muted, "Clear Search History").KeepFocus().OnClick(h.a.clearSearchHistory)
				})
				for _, history := range h.a.settings.SearchHistory {
					history := history
					option("history:"+history, history).Children(func() {
						ui.Icon(c, iconSearch).Size(13, 13).TextColor(h.a.pal.muted)
						ui.Text(c, history).FontSize(13).SingleLine().Shrink(1).MinWidth(0)
					})
				}
			} else {
				panel.Label("Search Suggestions")
				option("query", "Search \u201c"+q+"\u201d").Children(func() {
					ui.Icon(c, iconSearch).Size(13, 13).TextColor(h.a.pal.muted)
					ui.Text(c, "Search \u201c"+q+"\u201d").FontSize(13).FontWeight(500).SingleLine().Shrink(1).MinWidth(0)
				})
			}
		})
		if key, ok := cb.Chosen(); ok {
			cb.SetOpen(false)
			if history, found := strings.CutPrefix(key, "history:"); found {
				*home.SearchText() = history
				q = history
			}
			if key == "query" || strings.HasPrefix(key, "history:") {
				if q != "" {
					h.a.rememberSearch(q)
					submit(q)
				}
			}
		} else if in.Submitted() && q != "" {
			cb.SetOpen(false)
			h.a.rememberSearch(q)
			submit(q)
		}
	})
}

func (h platformHost) SearchHistory() []string {
	if h.a == nil {
		return nil
	}
	switch h.a.platformID {
	case platform.QQMusicID:
		return append([]string(nil), h.a.settings.QQMusic.SearchHistory...)
	case platform.FnmusicID:
		return append([]string(nil), h.a.settings.FnMusic.SearchHistory...)
	}
	return append([]string(nil), h.a.settings.SearchHistory...)
}

func (h platformHost) RememberSearch(q string) {
	if h.a == nil {
		return
	}
	var history []string
	switch h.a.platformID {
	case platform.QQMusicID:
		history = h.a.settings.QQMusic.SearchHistory
	case platform.FnmusicID:
		history = h.a.settings.FnMusic.SearchHistory
	default:
		h.a.rememberSearch(q)
		return
	}
	out := []string{q}
	for _, past := range history {
		if past != q && len(out) < 20 {
			out = append(out, past)
		}
	}
	switch h.a.platformID {
	case platform.QQMusicID:
		h.a.settings.QQMusic.SearchHistory = out
	case platform.FnmusicID:
		h.a.settings.FnMusic.SearchHistory = out
	}
	h.a.settings.save()
}

func (h platformHost) SwitchPlatform(id platform.ID) { h.a.setPlatform(id) }

func (h platformHost) CurrentPlatform() platform.ID {
	if h.a == nil {
		return platform.NeteaseID
	}
	return h.a.platformID
}

func (h platformHost) PlatformOptions() []appplatform.Info {
	if h.a == nil {
		return []appplatform.Info{neteasePlatformInfo()}
	}
	return h.a.platformInfos()
}

func (h platformHost) PlatformSwitcher(c *ui.Context) {
	if h.a != nil {
		h.a.platformSwitcher(c)
	}
}

// SharedSettings is the small common settings surface kept by the upstream
// app. Platform pages call it after their own settings instead of copying
// playback, cache, appearance or notch code into every adapter.
func (h platformHost) SharedSettings(c *ui.Context) {
	a, p, s := h.a, h.a.pal, &h.a.settings
	cap := neteasePlatformInfo()
	if active := a.activePlatform(); active != nil {
		cap = active.Info()
	}
	group := func(title string, rows func()) {
		ui.Text(c, title).FontSize(12).FontWeight(600).TextColor(p.muted).Margin(14, 4, 4, 4)
		ui.Column(c).Radius(12).Background(p.bg).Border(1, p.hairline).Dividers(1, p.hairline).Children(rows)
	}
	row := func(key, label, desc string, control func()) {
		ui.Row(c.Key(key)).Padding(12, 16).Gap(16).AlignItems(ui.Center).MinHeight(48).Children(func() {
			ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(3).Children(func() {
				ui.Text(c, label).FontSize(13.5).FontWeight(500)
				if desc != "" {
					ui.Text(c, desc).FontSize(12).TextColor(p.muted)
				}
			})
			control()
		})
	}
	if cap.Quality || cap.Normalize || cap.Comments {
		group("Play", func() {
			if cap.Quality {
				levels := a.platformQualityLevels()
				row("shared-quality", "Audio Quality", "Applies from the next song", func() {
					labels := make([]string, len(levels))
					chosen := a.qualityLevel()
					for i, q := range levels {
						labels[i] = q.label
						if q.level == a.qualityLevel() {
							chosen = q.label
						}
					}
					if ui.Select(c.Key("shared-quality-select"), &chosen, labels).Label("Audio Quality").Width(150).Changed() {
						for _, q := range levels {
							if q.label == chosen {
								a.setQualityLevel(q.level)
							}
						}
						a.changed()
					}
				})
			}
			if cap.Normalize {
				row("shared-normalize", "Volume Leveling", "Keeps different songs at a similar loudness", func() {
					on := a.normalizeEnabled()
					if ui.Switch(c.Key("shared-normalize-switch"), &on).Label("Volume Leveling").Changed() {
						a.setNormalize(on)
						a.pl.SetVolume(a.volume)
						a.changed()
					}
				})
			}
			if cap.Comments {
				row("shared-hot-comments", "Show top comments during instrumental breaks", "Shows top comments where the lyrics pause", func() {
					on := a.hotCommentsEnabled()
					if ui.Switch(c.Key("shared-hot-comments-switch"), &on).Label("Show top comments during instrumental breaks").Changed() {
						a.setHotComments(on)
						a.changed()
					}
				})
			}
		})
	}
	a.renderHifiSettings(c, group, row)
	group("Cache", func() {
		row("shared-cache", "Song Cache", a.cacheUsage(), func() {
			labels := make([]string, len(cacheLimits))
			chosen := ""
			for i, v := range cacheLimits {
				labels[i] = v.label
				if v.mb == s.CacheMB {
					chosen = v.label
				}
			}
			if ui.Select(c.Key("shared-cache-select"), &chosen, labels).Label("Song Cache").Width(150).Changed() {
				for _, v := range cacheLimits {
					if v.label == chosen {
						s.CacheMB = v.mb
					}
				}
				a.changed()
				a.applyCacheLimit()
			}
		})
		row("shared-cache-clear", "Clear Cache", "Deletes songs saved on this computer", func() {
			if ui.Button(c.Key("shared-cache-clear-button"), "Clear").Disabled(a.cacheUsed <= 0).Clicked() {
				a.clearCache()
			}
		})
	})
	group("Appearance", func() {
		row("shared-appearance", "Appearance", "Follow the system or pin a theme", func() {
			modes := []string{"system", "light", "dark"}
			i := 0
			for k, v := range modes {
				if v == s.Appearance {
					i = k
				}
			}
			if segmented(c.Key("shared-appearance-segment"), &i, "System", "Light", "Dark") {
				s.Appearance = modes[i]
				a.changed()
			}
		})
	})
	if runtime.GOOS == "darwin" {
		group("App Icon", func() {
			row("shared-app-icon", "App Icon", "The icon used when shown in the menu bar", func() {
				labels := make([]string, len(appIcons))
				chosen := appIcons[0].label
				for i, ic := range appIcons {
					labels[i] = ic.label
					if ic.id == s.AppIcon {
						chosen = ic.label
					}
				}
				if ui.Select(c.Key("shared-app-icon-select"), &chosen, labels).Label("App Icon").Width(150).Changed() {
					for _, ic := range appIcons {
						if ic.label == chosen {
							s.AppIcon = ic.id
						}
					}
					a.changed()
					a.changeAppIcon()
				}
			})
		})
	}
	if a.notch != nil {
		group("Notch Player", func() {
			row("shared-notch", "Show Notch Player", "Shows a mini player at the top of the screen", func() {
				if ui.Switch(c.Key("shared-notch-switch"), &s.Notch).Label("Show Notch Player").Changed() {
					a.changed()
				}
			})
			row("shared-notch-activity", "Show Live Activity while playing", "Shows the cover and waveform", func() {
				if ui.Switch(c.Key("shared-notch-activity-switch"), &s.NotchActivity).Label("Show Live Activity while playing").Disabled(!s.Notch).Changed() {
					a.changed()
				}
			})
			row("shared-notch-virtual", "Pin on screens without a notch", "When off, hover the pointer at the top center of the screen to reveal it", func() {
				if ui.Switch(c.Key("shared-notch-virtual-switch"), &s.NotchVirtual).Label("Pin on screens without a notch").Disabled(!s.Notch).Changed() {
					a.changed()
				}
			})
		})
	}
	if cap.Podcasts && a.captionsShown() {
		group("Podcasts", func() {
			a.captionSettings(func(label, desc string, control func()) {
				row("shared-caption-"+label, label, desc, control)
			}, c)
		})
	}
	group("Experimental", func() {
		row("shared-smoke", "Smoke effect on the progress bar", "The progress bar becomes points of light that swell with the music", func() {
			if ui.Switch(c.Key("shared-smoke-switch"), &s.Smoke).Label("Smoke effect on the progress bar").Changed() {
				a.changed()
			}
		})
		row("shared-fog", "Smoke background", "The Now Playing background rises slowly in the cover's colors", func() {
			if ui.Switch(c.Key("shared-fog-switch"), &s.Fog).Label("Smoke background").Changed() {
				a.changed()
			}
		})
	})
	if a.platformID != platform.NeteaseID {
		a.platformSettingsGroup(c, group, func(label, desc string, control func()) {
			row("platform-"+label, label, desc, control)
		})
	}
	if a.updates {
		group("Update", func() {
			row("shared-update-checks", "Check for updates automatically", "Checks once a day for new versions", func() {
				on := updateChecks()
				if ui.Switch(c.Key("shared-update-checks-switch"), &on).Label("Check for updates automatically").Changed() {
					setUpdateChecks(on)
				}
			})
			row("shared-update-downloads", "Download and install automatically", "Downloads in the background; applies on next launch", func() {
				on := updateDownloads()
				if ui.Switch(c.Key("shared-update-downloads-switch"), &on).Label("Download and install automatically").Disabled(!updateChecks()).Changed() {
					setUpdateDownloads(on)
				}
			})
			row("shared-version", "Current version "+mygo.App.Version(), "", func() {
				if ui.Button(c.Key("shared-check-updates"), "Check for Updates").Clicked() {
					checkForUpdates()
				}
			})
		})
	}
}

func (h platformHost) NavItem(c *ui.Context, icon *ui.SVG, label string, r appplatform.Route) {
	p := h.a.pal
	home := h.a.activeHome()
	selected := home != nil && home.Route() == r && !h.a.nowPlaying
	b := ui.ButtonBase(c.Key("platform-" + label)).Height(34).PaddingX(10).Gap(10).Radius(8).
		AlignItems(ui.Center).Justify(ui.Start).FocusRing(false).Label(label)
	bg := ui.Transparent
	if selected {
		bg = p.press
	} else if b.Hovered() {
		bg = p.hover
	}
	b.Background(b.AnimateColor("bg", bg, ui.Animation{Duration: 120 * time.Millisecond}))
	color := p.text
	if selected {
		color = p.accent
	}
	b.Children(func() {
		ui.Icon(c, icon).Size(17, 17).TextColor(color)
		ui.Text(c, label).FontSize(13.5).FontWeight(500).TextColor(p.text)
	})
	if b.Clicked() {
		h.Navigate(r)
	}
}

func (h platformHost) PlaylistHeader(c *ui.Context, create func()) {
	p := h.a.pal
	ui.Row(c).AlignItems(ui.Center).Margin(18, 4, 6, 10).Children(func() {
		ui.Text(c, "Playlists").FontSize(11.5).FontWeight(600).TextColor(p.faint).Grow(1)
		if create != nil && iconButton(c.Key("platform-new-playlist"), iconPlus, 13, p.faint, "New Playlist").Clicked() {
			create()
		}
	})
}

func (h platformHost) PlaylistItem(c *ui.Context, key string, pl netease.Playlist, selected bool, menu func(*ui.Menu), click func()) {
	p := h.a.pal
	b := ui.ButtonBase(c.Key(key)).Height(34).PaddingX(8).Gap(10).Radius(8).AlignItems(ui.Center).Justify(ui.Start).FocusRing(false).Label(pl.Name)
	if selected {
		b.Background(p.press)
	} else if b.Hovered() {
		b.Background(p.hover)
	}
	b.Children(func() {
		h.a.cover(c, pl.Cover(), 80, 22, 5)
		ui.Text(c, pl.Name).FontSize(13).SingleLine().Shrink(1).MinWidth(0)
	})
	if menu != nil {
		b.ContextMenu(menu)
	}
	if b.Clicked() && click != nil {
		click()
	}
}

func (h platformHost) NowPlaying(c *ui.Context, tr *netease.Track) {
	h.a.sidebarNowPlaying(c, tr)
}

func (h platformHost) Cover(c *ui.Context, url string, size, radius float32) ui.Element {
	px := int(size * 3)
	if size <= 40 {
		px = 80
	}
	return h.a.cover(c, url, px, size, radius)
}

func (h platformHost) CoverBitmap(url string) *ui.Bitmap {
	if h.a == nil || url == "" {
		return nil
	}
	return bitmap(h.a.art.get(netease.Sized(url, 240), false))
}

func (h platformHost) ArtworkBackdrop(url string) *ui.Bitmap {
	if h.a == nil || url == "" {
		return nil
	}
	if art := h.a.art.get(netease.Sized(url, 600), true); art != nil {
		return art.backdrop
	}
	return nil
}

func (h platformHost) ArtworkColors(url string) (ui.Color, ui.Color, ui.Color, bool) {
	if h.a == nil || url == "" {
		return ui.Color{}, ui.Color{}, ui.Color{}, false
	}
	art := h.a.art.get(netease.Sized(url, 600), false)
	if art == nil || art.failed {
		return ui.Color{}, ui.Color{}, ui.Color{}, false
	}
	deep, rich, glow := dailyColors(art)
	return deep, rich, glow, true
}

func (h platformHost) PillButton(c *ui.Context, key string, icon *ui.SVG, label string, primary bool) ui.Element {
	p := h.a.pal
	bg, fg := p.hover.Over(p.bg), p.text
	if primary {
		bg, fg = p.accent, p.onAccent
	}
	return pillButton(c.Key(key), icon, label, bg, fg)
}

func (h platformHost) MoreButton(c *ui.Context, key string, menu func(*ui.Menu)) ui.Element {
	b := iconButton(c.Key(key), iconMore, 18, h.a.pal.muted, "More")
	b.Menu(menu)
	return b
}
func (h platformHost) Segmented(c *ui.Context, key string, at *int, labels ...string) bool {
	return segmented(c.Key(key), at, labels...)
}

// AccountRow renders the shared account-row chrome while leaving the account
// labels and actions to the active platform. This keeps provider account
// menus out of the root app without duplicating popover geometry and focus
// behavior.
func (h platformHost) AccountRow(c *ui.Context, accountIcon *ui.SVG, avatar, username, status string, open *bool, actions []appplatform.AccountAction) {
	if h.a == nil {
		return
	}
	p := h.a.pal
	if open == nil {
		return
	}
	b := ui.ButtonBase(c.Key("platform-account")).Height(40).PaddingX(8).Gap(10).Radius(10).
		AlignItems(ui.Center).Justify(ui.Start).FocusRing(false).Label(username)
	if b.Hovered() || *open {
		b.Background(p.hover)
	}
	b.Expanded(*open).Children(func() {
		if avatar != "" {
			h.a.avatar(c, avatar, 26, p.press)
		} else {
			ui.Box(c).Size(26, 26).Radius(13).Background(p.press).Center().Children(func() {
				ui.Icon(c, accountIcon).Size(14, 14).TextColor(p.muted)
			})
		}
		ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(1).Children(func() {
			ui.Text(c, username).FontSize(13).FontWeight(600).SingleLine().Shrink(1).MinWidth(0)
			if status != "" {
				ui.Text(c, status).FontSize(11.5).TextColor(p.faint).SingleLine().Shrink(1).MinWidth(0)
			}
		})
		ui.Icon(c, iconUpDown).Size(13, 13).TextColor(p.faint)
	})
	if b.Clicked() {
		*open = !*open
	}
	pop := popoverColors(h.a.pal.dark)
	ui.PopoverBase(c.Key("platform-account-menu"), b, open, func(panel ui.Element) {
		panel.AttachTo(b, ui.AnchorTopStart, ui.AnchorBottomStart).Margin(0, 0, 6, 0).Padding(5).Gap(1).
			Radius(12).Background(pop.fill).Border(0.5, pop.edge).Shadow(0, 8, 24, -4, ui.RGBA(0, 0, 0, 0.2)).
			Role(ui.RoleMenu).FocusGroup(ui.Vertical).Label("Account").TransformOrigin(0.5, 1).
			Transition(ui.ElementTransition{Animation: &ui.Snappy, Enter: &ui.Motion{Scale: 0.96, Y: 4}, Exit: &ui.Motion{Scale: 0.98}})
		if w := b.Bounds().W; w > 0 {
			panel.Width(w)
		}
		for i, action := range actions {
			if action.Separator && i > 0 {
				ui.Box(c).Height(0.5).Margin(4, 8).Background(pop.edge)
			}
			item := h.a.menuItem(c, action.Icon, action.Label, action.Shortcut)
			if action.Danger {
				item.TextColor(p.accent)
			}
			if item.Clicked() {
				*open = false
				if action.Action != nil {
					action.Action()
				}
			}
		}
	})
}

func (h platformHost) CopyText(text string) { copyText(text) }

func (h platformHost) OpenSettings() { h.a.openSettings() }

func (h platformHost) RenderTracks(c *ui.Context, l *appplatform.TrackList, view appplatform.TrackView) {
	if l == nil || l.Len == nil || l.At == nil {
		h.a.centered(c, "This list could not be found", nil)
		return
	}
	if l.Failed != nil && l.Failed() {
		h.a.centered(c, "Failed to load. Please try again", l.Retry)
		return
	}
	// At is intentionally called while the list is built. It starts the first
	// page and leaves the pager state independent from UI rebuilds.
	if l.Total != nil && l.Total() < 0 {
		// Start the hero request while the first page is still a skeleton. The
		// first real frame should never be the one that starts downloading it.
		if view.Cover != "" {
			h.a.platformArtworkReady(view.Cover, 636)
		}
		l.At(0)
		h.a.platformTrackSkeleton(c, view)
		return
	}
	total := l.Len()
	if total == 0 {
		h.a.centered(c, "No songs here yet", nil)
		return
	}
	if !h.platformHeroReady(c, view.Cover) {
		h.a.platformTrackSkeleton(c, view)
		return
	}
	h.prefetchTrackArtwork(l, total, c)
	get := l.At
	play := view.OnPlay
	if play == nil {
		play = func(int) {}
	}

	var footer []func(*ui.Context)
	if view.Footer != nil {
		footer = append(footer, view.Footer)
	}
	h.a.trackRowsProviderOptions(c, total, get, play, true, trackRowsOptions{
		Menu:      view.Menu,
		Liked:     view.Liked,
		Favorite:  view.Favorite,
		ShowAlbum: view.ShowAlbum,
	}, func(c *ui.Context) {
		if view.Header != nil {
			view.Header(c)
		} else {
			h.a.heading(c, view.Title, view.Sub)
		}
	}, footer...)
}

// platformHeroReady gates only the detail cover. A slow thumbnail must never
// keep the page behind its skeleton, and a slow hero must not hold it forever.
func (h platformHost) platformHeroReady(c *ui.Context, cover string) bool {
	if cover == "" {
		return true
	}
	if h.a.platformHeroSince.IsZero() {
		h.a.platformHeroSince = c.Now()
	}
	if h.a.platformArtworkReady(cover, 636) {
		return true
	}
	if c.Now().Sub(h.a.platformHeroSince) >= 450*time.Millisecond {
		return true
	}
	c.After(40 * time.Millisecond)
	return false
}

func (h platformHost) prefetchTrackArtwork(l *appplatform.TrackList, total int, c *ui.Context) {
	if l == nil || l.At == nil {
		return
	}
	_, height := c.Size()
	count := max(12, int(height/54)+6)
	count = min(count, total)
	urls := make([]string, 0, count)
	for i := 0; i < count; i++ {
		if tr := l.At(i); tr != nil {
			urls = append(urls, tr.Album.PicURL)
		}
	}
	h.a.prefetchPlatformArtwork(urls, 96)
}

func (h platformHost) CardsSkeleton(c *ui.Context, view appplatform.CardView) {
	h.a.platformCardsSkeleton(c, view)
}

func (h platformHost) RenderCards(c *ui.Context, p *paging.Pager[netease.Playlist], view appplatform.CardView) {
	if p == nil {
		h.a.centered(c, "No results found", nil)
		return
	}
	p.At(0)
	header := func(c *ui.Context) {
		if view.Header != nil {
			view.Header(c)
			return
		}
		ui.Column(c).Gap(4).Margin(0, 0, 18, 0).Children(func() {
			ui.Text(c, view.Title).FontSize(30).FontWeight(700).LetterSpacing(-0.6)
			if view.Sub != "" {
				ui.Text(c, view.Sub).FontSize(12).FontWeight(600).TextColor(h.a.pal.muted).LetterSpacing(0.4)
			}
			if view.Count > 0 {
				ui.Text(c, fmt.Sprintf("%d in total", view.Count)).FontSize(13).TextColor(h.a.pal.muted)
			}
		})
	}
	top, bottom := h.a.insets(c)
	if p.FailedAt(0) {
		ui.Column(c).Grow(1).Padding(top+pageGap, pagePad, bottom+48, pagePad).Children(func() {
			header(c)
			if len(view.Sort) > 0 {
				ui.Row(c).Margin(-6, 0, 18, 0).AlignItems(ui.Center).Gap(10).Children(func() {
					ui.Text(c, "Sort").FontSize(12).TextColor(h.a.pal.muted)
					labels := make([]string, len(view.Sort))
					at := 0
					for i, ch := range view.Sort {
						labels[i] = ch.Label
						if ch.Value == view.Current {
							at = i
						}
					}
					if segmented(c.Key("platform-sort"), &at, labels...) && view.OnSort != nil {
						view.OnSort(view.Sort[at].Value)
					}
				})
			}
			ui.Column(c).Grow(1).Center().Gap(12).Children(func() {
				ui.Text(c, "Failed to load. Please try again").TextColor(h.a.pal.muted).FontSize(13)
				if ui.Button(c, "Retry").Clicked() {
					p.Retry()
				}
			})
		})
		return
	}
	if p.Total() < 0 {
		h.a.platformCardsSkeleton(c, view)
		return
	}
	h.prefetchCardArtwork(p, c)
	pages, size := p.PageCount(), p.PageSize()
	if pages == 0 {
		ui.Column(c).Grow(1).Padding(top+pageGap, pagePad, bottom+48, pagePad).Children(func() {
			header(c)
			if len(view.Sort) > 0 {
				ui.Row(c).Margin(-6, 0, 18, 0).AlignItems(ui.Center).Gap(10).Children(func() {
					ui.Text(c, "Sort").FontSize(12).TextColor(h.a.pal.muted)
					labels := make([]string, len(view.Sort))
					at := 0
					for i, ch := range view.Sort {
						labels[i] = ch.Label
						if ch.Value == view.Current {
							at = i
						}
					}
					if segmented(c.Key("platform-sort"), &at, labels...) && view.OnSort != nil {
						view.OnSort(view.Sort[at].Value)
					}
				})
			}
			ui.Column(c).Grow(1).Center().Children(func() {
				ui.Text(c, "No results found").TextColor(h.a.pal.muted).FontSize(13)
			})
		})
		return
	}
	h.a.underTop, h.a.underBottom = true, true
	renderArt := view.Art
	if renderArt == nil && view.Placeholder != nil {
		renderArt = h.a.iconCardArt(view.Placeholder)
	}
	routeKey := "platform-cards"
	if home := h.a.activeHome(); home != nil {
		r := home.Route()
		routeKey = fmt.Sprintf("platform-cards-%s-%d-%s", r.Page, r.ID, r.Query)
	}
	rows := pages + 1
	if view.Footer != nil {
		rows++
	}
	ui.List(c.Key(routeKey), &h.a.platformCardList, rows, func(row int) {
		if view.Footer != nil && row == rows-1 {
			ui.Column(c).Padding(18, 0, 32, 0).Children(func() { view.Footer(c) })
			return
		}
		if row == 0 {
			ui.Column(c).Children(func() {
				h.a.startMark(c)
				header(c)
				if len(view.Sort) > 0 {
					ui.Row(c).Margin(-6, 0, 18, 0).AlignItems(ui.Center).Gap(10).Children(func() {
						ui.Text(c, "Sort").FontSize(12).TextColor(h.a.pal.muted)
						labels := make([]string, len(view.Sort))
						at := 0
						for i, ch := range view.Sort {
							labels[i] = ch.Label
							if ch.Value == view.Current {
								at = i
							}
						}
						if segmented(c.Key("platform-sort"), &at, labels...) && view.OnSort != nil {
							view.OnSort(view.Sort[at].Value)
						}
					})
				}
			})
			return
		}
		page := row - 1
		start := page * size
		p.At(start)
		cols, card := h.a.cardWidth(c)
		ui.Grid(c).Columns(cols).GapX(20).GapY(28).Children(func() {
			for i := start; i < min(start+size, p.Len()); i++ {
				pl := p.Peek(i)
				if pl == nil {
					ui.Box(c.Key(fmt.Sprintf("platform-card-pending-%d", i))).Width(card).Children(func() {
						h.a.platformCardSkeleton(c, h.a.pageShimmer(c), i, view.Meta)
					})
					continue
				}
				ui.Box(c.Key(fmt.Sprintf("platform-card-wrap-%d", pl.ID))).Width(card).Children(func() {
					play := func() {
						if view.OnPlay != nil {
							view.OnPlay(pl.ID)
						}
					}
					key := c.Key(fmt.Sprintf("platform-card-%d", pl.ID))
					var open bool
					switch {
					case view.Round:
						open = h.roundCard(key, pl, card)
					case view.NoPlay:
						open = h.plainCard(key, pl, card)
					default:
						if pl.Cover() == "" && renderArt != nil {
							card := *pl
							open = h.a.drawnCard(key, pl.Name, pl.UpdateFreq, func(p *ui.Painter, r ui.Rect, radius float32) {
								renderArt(p, r, &card, radius)
							}, play)
						} else {
							open = h.a.coverCard(key, pl.Name, pl.Cover(), pl.PlayCount, pl.UpdateFreq, play)
						}
					}
					if open && view.OnOpen != nil {
						view.OnOpen(pl.ID)
					}
				})
			}
		})
		if row == pages {
			defer h.a.endMark(c)
		}
	}).Grow(1).Padding(top+pageGap, pagePad, bottom+48, pagePad).ScrollbarInsets(top, 0, bottom, 0)
}

// roundCard is an artist's photo, round over their name, opening their page.
func (h platformHost) roundCard(c *ui.Context, pl *netease.Playlist, size float32) bool {
	p := h.a.pal
	b := ui.ButtonBase(c).Column().Gap(10).AlignItems(ui.Stretch).FocusRing(false).Label(pl.Name)
	hovered := b.Hovered()
	b.Children(func() {
		art := ui.Box(c).AspectRatio(1).Radius(size / 2)
		if hovered {
			art.AnimatePose(ui.Pose{Y: -3}, ui.Snappy)
			art.Shadow(0, 12, 28, -6, p.cardShade)
		} else {
			art.Shadow(0, 2, 8, -2, p.cardShade)
		}
		art.Children(func() { h.Cover(c, pl.Cover(), size, size/2) })
		ui.Text(c, pl.Name).FontSize(13).FontWeight(500).SingleLine().AlignSelf(ui.Center).TextAlign(ui.Center)
	})
	return b.Clicked()
}

// plainCard is a cover over its name and line, with no play: an album grid.
func (h platformHost) plainCard(c *ui.Context, pl *netease.Playlist, size float32) bool {
	p := h.a.pal
	b := ui.ButtonBase(c).Column().Gap(10).AlignItems(ui.Stretch).FocusRing(false).Label(pl.Name)
	hovered := b.Hovered()
	b.Children(func() {
		art := ui.Box(c).AspectRatio(1).Radius(10)
		if hovered {
			art.AnimatePose(ui.Pose{Y: -3}, ui.Snappy)
			art.Shadow(0, 12, 28, -6, p.cardShade)
		} else {
			art.Shadow(0, 2, 8, -2, p.cardShade)
		}
		art.Children(func() { h.Cover(c, pl.Cover(), size, 10) })
		ui.Text(c, pl.Name).FontSize(13).FontWeight(500).MaxLines(2).LineHeight(1.35)
		if pl.UpdateFreq != "" {
			ui.Text(c, pl.UpdateFreq).FontSize(12).TextColor(p.muted).SingleLine().Margin(-6, 0, 0, 0)
		}
	})
	return b.Clicked()
}

func (h platformHost) TrackRow(c *ui.Context, i int, tr *netease.Track, view appplatform.TrackView) {
	w, _ := c.Size()
	wide := w-sidebarWidth > 820
	if tr == nil {
		h.a.pendingTrackRow(c, i, wide, true, view.ShowAlbum)
		return
	}
	play := view.OnPlay
	if play == nil {
		play = func(int) {}
	}
	h.a.providerTrackRow(c, i, tr, h.a.pl.Current(), 0, w-sidebarWidth > 820, true, trackRowsOptions{
		Menu: view.Menu, Liked: view.Liked, Favorite: view.Favorite, ShowAlbum: view.ShowAlbum,
	}, play)
}

func (h platformHost) TrackSkeleton(c *ui.Context, view appplatform.TrackView) {
	h.a.platformTrackSkeleton(c, view)
}

func (h platformHost) Bone(c *ui.Context, radius float32) ui.Element {
	return h.a.pageShimmer(c).bone(c, radius)
}

func (h platformHost) TextBone(c *ui.Context, size, lineHeight float32) ui.Element {
	return h.a.pageShimmer(c).text(c, size, lineHeight)
}

func (h platformHost) PageList(c *ui.Context, key string, state *ui.ListState, rows int, row func(int)) {
	top, bottom := h.a.insets(c)
	h.a.underTop, h.a.underBottom = true, true
	ui.List(c.Key(key), state, rows, func(i int) {
		ui.Column(c).Children(func() {
			if i == 0 {
				h.a.startMark(c)
			}
			row(i)
			if i == rows-1 {
				h.a.endMark(c)
			}
		})
	}).Grow(1).Padding(top+pageGap, pagePad-12, bottom+32, pagePad-12).ScrollbarInsets(top, 0, bottom, 0)
}

func (h platformHost) SearchSkeleton(c *ui.Context) {
	h.a.platformSearchSkeleton(c)
}

func searchPageLoading(p *paging.Pager[netease.Playlist]) bool {
	return p != nil && p.Total() < 0 && !p.FailedAt(0)
}

func searchPageEmpty(p *paging.Pager[netease.Playlist]) bool {
	return p == nil || p.FailedAt(0) || p.Total() >= 0 && p.Len() == 0
}

func (h platformHost) RenderSearch(c *ui.Context, s appplatform.SearchResults, view appplatform.SearchView) {
	if s.Songs == nil {
		h.a.centered(c, "No search yet", nil)
		return
	}
	// Search keeps the song list as the primary virtual list. Album and artist
	// cards are rendered above it from their first loaded page, so the page
	// remains useful even when a result has thousands of songs.
	albums, artists := s.Albums, s.Artists
	if albums != nil {
		albums.At(0)
	}
	if artists != nil {
		artists.At(0)
	}
	if s.Songs.Failed != nil && s.Songs.Failed() {
		h.a.centered(c, "Failed to load. Please try again", s.Songs.Retry)
		return
	}
	songsLoading := s.Songs.Total != nil && s.Songs.Total() < 0
	if songsLoading || searchPageLoading(albums) || searchPageLoading(artists) {
		h.a.platformSearchSkeleton(c)
		return
	}
	if s.Songs.Len() == 0 && searchPageEmpty(albums) && searchPageEmpty(artists) {
		h.a.centered(c, "No results found", nil)
		return
	}
	h.a.trackRowsProviderOptions(c, s.Songs.Len(), s.Songs.At, func(i int) {
		if view.Play != nil {
			view.Play(i)
		}
	}, true, trackRowsOptions{
		Menu:      view.TrackMenu,
		Liked:     view.Liked,
		Favorite:  view.Favorite,
		ShowAlbum: true,
	}, func(c *ui.Context) {
		h.a.heading(c, "“"+s.Query+"”", "Search")
		if albums != nil && albums.Len() > 0 {
			h.searchCardStrip(c, "Albums", albums, view.OpenList, view.PlayCard, func() {
				if view.OpenCards != nil {
					view.OpenCards("albums")
				}
			})
		}
		if artists != nil && artists.Len() > 0 {
			h.searchCardStrip(c, "Artists", artists, view.OpenList, view.PlayCard, func() {
				if view.OpenCards != nil {
					view.OpenCards("artists")
				}
			})
		}
		if s.Songs.Len() > 0 {
			h.a.section(c, "Songs")
		}
	})
}

func (h platformHost) searchCardStrip(c *ui.Context, title string, p *paging.Pager[netease.Playlist], open func(int64), play func(int64), all func()) {
	cols := h.a.columns(c, 168)
	ui.Row(c).AlignItems(ui.Center).Margin(34, 0, 14, 0).Children(func() {
		ui.Text(c, title).FontSize(19).FontWeight(700).LetterSpacing(-0.3).Grow(1)
		if p.Total() > cols {
			b := ui.ButtonBase(c.Key("search-all-" + title)).PaddingX(8).Height(26).Radius(7).AlignItems(ui.Center).FocusRing(false).Label("See all " + title)
			b.Background(b.AnimateColor("bg", map[bool]ui.Color{true: h.a.pal.hover, false: ui.Transparent}[b.Hovered()], ui.Animation{Duration: 120 * time.Millisecond}))
			b.Children(func() {
				ui.Text(c, fmt.Sprintf("All %d", p.Total())).FontSize(12.5).FontWeight(500).TextColor(h.a.pal.muted)
			})
			if b.Clicked() && all != nil {
				all()
			}
		}
	})
	ui.Row(c).Gap(20).Children(func() {
		for i := 0; i < min(cols, p.Len()); i++ {
			if pl := p.Peek(i); pl != nil {
				ui.Box(c.Key(fmt.Sprintf("search-%s-%d", title, pl.ID))).Width(168).Children(func() {
					clicked := h.a.coverCard(c, pl.Name, pl.Cover(), pl.PlayCount, pl.UpdateFreq, func() {
						if play != nil {
							play(pl.ID)
						}
					})
					if clicked && open != nil {
						open(pl.ID)
					}
				})
			}
		}
	})
}

func (h platformHost) prefetchCardArtwork(p *paging.Pager[netease.Playlist], c *ui.Context) {
	if p == nil {
		return
	}
	cols := h.a.columns(c, 168)
	count := min(p.Len(), cols*2)
	urls := make([]string, 0, count)
	for i := 0; i < count; i++ {
		if pl := p.Peek(i); pl != nil {
			urls = append(urls, pl.Cover())
		}
	}
	h.a.prefetchPlatformArtwork(urls, 400)
}

// sharedPlayer keeps the concrete player out of the platform package while
// preserving every existing player feature in the root application.
type sharedPlayer struct {
	p interface {
		Current() *netease.Track
		Playing() bool
		Loading() bool
		Err() string
		TogglePlay()
		Next()
		Queue() []netease.Track
		PlayOrder() ([]int, int)
		PlayQueue([]netease.Track, int)
		AppendQueue([]netease.Track)
		PlayList([]netease.Track, int, func(func([]netease.Track)))
		PlayRadio(string, []netease.Track, func(func([]netease.Track)))
		EndRadio()
		Radio() string
		ToggleShuffle()
		Shuffle() bool
	}
}

func (p sharedPlayer) Current() *netease.Track            { return p.p.Current() }
func (p sharedPlayer) Playing() bool                      { return p.p.Playing() }
func (p sharedPlayer) Loading() bool                      { return p.p.Loading() }
func (p sharedPlayer) Err() string                        { return p.p.Err() }
func (p sharedPlayer) TogglePlay()                        { p.p.TogglePlay() }
func (p sharedPlayer) Next()                              { p.p.Next() }
func (p sharedPlayer) Queue() []netease.Track             { return p.p.Queue() }
func (p sharedPlayer) PlayOrder() ([]int, int)            { return p.p.PlayOrder() }
func (p sharedPlayer) PlayQueue(t []netease.Track, i int) { p.p.PlayQueue(t, i) }
func (p sharedPlayer) AppendQueue(t []netease.Track)      { p.p.AppendQueue(t) }
func (p sharedPlayer) PlayList(t []netease.Track, i int, f func(func([]netease.Track))) {
	p.p.PlayList(t, i, f)
}
func (p sharedPlayer) PlayRadio(n string, t []netease.Track, f func(func([]netease.Track))) {
	p.p.PlayRadio(n, t, f)
}
func (p sharedPlayer) EndRadio()     { p.p.EndRadio() }
func (p sharedPlayer) Radio() string { return p.p.Radio() }
func (p sharedPlayer) SetShuffle(on bool) {
	if p.p.Shuffle() != on {
		p.p.ToggleShuffle()
	}
}

var _ appplatform.Host = platformHost{}
