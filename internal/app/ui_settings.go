package app

import (
	"runtime"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// openSettings shows the settings window, opening it the first time.
func (a *app) openSettings() {
	if a.settingsWin != nil {
		a.settingsWin.Show()
		a.settingsWin.Focus()
		a.measureCache()
		return
	}
	a.measureCache()
	a.settingsWin = mygo.NewWindow(mygo.WindowOptions{
		Title:           "Settings",
		Width:           560,
		Height:          720,
		MinWidth:        480,
		MinHeight:       420,
		StateKey:        "settings",
		DisableMaximize: true,
		DisableMinimize: true,
		// The app's menu has nothing for this window: on Linux and
		// Windows it stays out of sight until Alt shows it.
		AutoHideMenuBar: true,
		BackgroundColor: "light-dark(#f4f4f6, #1c1c1f)",
		Content:         ui.View(a.settingsView),
	})
	a.settingsWin.OnClosed(func() {
		a.settingsWin = nil
		a.syncDock()
	})
	a.syncDock()
}

// changed applies the settings everywhere and keeps them.
func (a *app) changed() {
	a.settings.save()
	a.settings.applyAppearance()
	if a.notch != nil {
		a.notch.setEnabled(a.settings.Notch)
	}
	a.redraw()
}

func (a *app) settingsView(c *ui.Context) {
	c.OnShortcut(ui.Cmd, ui.KeyW, func() { a.settingsWin.Close() })
	if home := a.activeHome(); home != nil {
		home.Settings(c)
		return
	}
	t := *c.Theme()
	p := paletteFor(&t)
	a.platformAccent(&p)
	t.Accent, t.AccentText = p.accent, p.onAccent
	c.SetTheme(&t)
	c.Root().Background(p.sidebar)
	s := &a.settings

	ui.Scroll(c).Fill().Padding(24, 28, 32, 28).Gap(8).Children(func() {
		group := func(title string, rows func()) {
			ui.Text(c, title).FontSize(12).FontWeight(600).TextColor(p.muted).Margin(14, 4, 4, 4)
			ui.Column(c).Radius(12).Background(p.bg).Border(1, p.hairline).Dividers(1, p.hairline).Children(rows)
		}
		row := func(label, desc string, control func()) {
			ui.Row(c).Padding(12, 16).Gap(16).AlignItems(ui.Center).MinHeight(48).Children(func() {
				ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(3).Children(func() {
					ui.Text(c, label).FontSize(13.5).FontWeight(500)
					if desc != "" {
						ui.Text(c, desc).FontSize(12).TextColor(p.muted).LineHeight(1.4)
					}
				})
				control()
			})
		}

		// Only on macOS: elsewhere the icon is always in the notification
		// area.
		if runtime.GOOS == "darwin" {
			group("General", func() {
				row("App Icon", "When shown in the menu bar, the app can be reopened from its icon after the window closes", func() {
					labels := make([]string, len(appIcons))
					chosen := appIcons[0].label
					for i, ic := range appIcons {
						labels[i] = ic.label
						if ic.id == s.AppIcon {
							chosen = ic.label
						}
					}
					if ui.Select(c.Key("app-icon"), &chosen, labels).Label("App Icon").Width(150).Changed() {
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

		// Only on macOS, the one with a notch player.
		if a.notch != nil {
			group("Notch Player", func() {
				row("Show Notch Player", "Shows a mini player at the notch on top of the screen, which expands on hover", func() {
					if ui.Switch(c.Key("notch"), &s.Notch).Label("Show Notch Player").Changed() {
						a.changed()
					}
				})
				row("Show Live Activity while playing", "Shows the cover and audio waveform beside the notch; may cover menu bar icons next to it", func() {
					if ui.Switch(c.Key("activity"), &s.NotchActivity).Label("Show Live Activity while playing").Disabled(!s.Notch).Changed() {
						a.changed()
					}
				})
				row("Pin on screens without a notch", "When off, move the pointer to the top center of the screen to reveal it without covering the menu bar", func() {
					if ui.Switch(c.Key("virtual"), &s.NotchVirtual).Label("Pin on screens without a notch").Disabled(!s.Notch).Changed() {
						a.changed()
					}
				})
			})
		}

		group("Play", func() {
			row("Audio Quality", "Lossless and Hi-Res need Vinyl VIP; without it the highest quality you can play is used. Takes effect from the next song", func() {
				labels := make([]string, len(qualityLevels))
				chosen := qualityLevels[2].label
				for i, q := range qualityLevels {
					labels[i] = q.label
					if q.level == s.Quality {
						chosen = q.label
					}
				}
				if ui.Select(c.Key("quality"), &chosen, labels).Label("Audio Quality").Width(150).Changed() {
					for _, q := range qualityLevels {
						if q.label == chosen {
							s.Quality = q.level
						}
					}
					a.changed()
				}
			})
			row("Show top comments during instrumental breaks", "During intros, interludes and outros, the lyrics give way to this song's top comments", func() {
				if ui.Switch(c.Key("hot-comments"), &s.HotComments).Label("Show top comments during instrumental breaks").Changed() {
					a.changed()
				}
			})
			row("Volume Leveling", "Adjusts each song's volume by the loudness NetEase Cloud Music measured, so songs sound equally loud", func() {
				if ui.Switch(c.Key("normalize"), &s.Normalize).Label("Volume Leveling").Changed() {
					a.pl.SetVolume(a.volume)
					a.changed()
				}
			})
		})

		a.renderHifiSettings(c, group, func(key, label, desc string, control func()) {
			row(label, desc, control)
		})

		if a.captionsShown() {
			group("Podcasts", func() { a.captionSettings(row, c) })
		}

		group("Cache", func() {
			row("Song Cache", "Songs you heard are kept on this machine, so replaying them needs no waiting and uses no data; "+a.cacheUsage(), func() {
				labels := make([]string, len(cacheLimits))
				chosen := cacheLimits[2].label
				for i, l := range cacheLimits {
					labels[i] = l.label
					if l.mb == s.CacheMB {
						chosen = l.label
					}
				}
				if ui.Select(c.Key("cache"), &chosen, labels).Label("Song Cache").Width(150).Changed() {
					for _, l := range cacheLimits {
						if l.label == chosen {
							s.CacheMB = l.mb
						}
					}
					a.changed()
					a.applyCacheLimit()
				}
			})
			row("Clear Cache", "Deletes all songs kept on this machine", func() {
				if ui.Button(c.Key("clear-cache"), "Clear").Disabled(a.cacheUsed <= 0).Clicked() {
					a.clearCache()
				}
			})
		})

		group("Appearance", func() {
			row("Appearance", "", func() {
				modes := []string{"system", "light", "dark"}
				i := 0
				for k, m := range modes {
					if m == s.Appearance {
						i = k
					}
				}
				if segmented(c.Key("appearance"), &i, "System", "Light", "Dark") {
					s.Appearance = modes[i]
					a.changed()
				}
			})
		})

		group("Experimental", func() {
			row("Smoke effect on the progress bar", "While playing, the progress bar becomes a spark trailing smoke that rises and falls with the music; uses a bit more battery", func() {
				if ui.Switch(c.Key("smoke"), &s.Smoke).Label("Smoke effect on the progress bar").Changed() {
					a.changed()
				}
			})
			row("Smoke background", "The Now Playing background becomes smoke in the cover's colors, slowly rising with the bass; still being polished, uses a bit more battery", func() {
				if ui.Switch(c.Key("fog"), &s.Fog).Label("Smoke background").Changed() {
					a.changed()
				}
			})
		})

		a.platformSettingsGroup(c, group, row)

		group("Account", func() {
			if a.profile == nil {
				row("Not signed in", "Once signed in, daily picks and playlists are tailored to your taste", func() {
					if ui.PrimaryButton(c.Key("login"), "Scan QR to Sign In").Clicked() {
						a.showMain()
						a.openLogin()
						a.win.Invalidate()
					}
				})
				return
			}
			prof := a.profile
			ui.Row(c).Padding(12, 16).Gap(12).AlignItems(ui.Center).Children(func() {
				a.avatar(c, prof.AvatarURL, 36, p.press)
				ui.Column(c).Grow(1).Gap(2).Children(func() {
					ui.Text(c, prof.Nickname).FontSize(13.5).FontWeight(600)
					ui.Text(c, "Signed in to NetEase Cloud Music").FontSize(12).TextColor(p.muted)
				})
				if ui.Button(c.Key("logout"), "Sign Out").Clicked() {
					a.signOut()
				}
			})
		})

		if a.updates {
			group("Update", func() {
				desc := "Checks once a day for new versions"
				if t := lastUpdateCheck(); !t.IsZero() {
					desc += "; last checked: " + commentTime(t, time.Now())
				}
				row("Check for updates automatically", desc, func() {
					on := updateChecks()
					if ui.Switch(c.Key("update-checks"), &on).Label("Check for updates automatically").Changed() {
						setUpdateChecks(on)
					}
				})
				row("Download and install automatically", "Downloads in the background; applies on next launch", func() {
					on := updateDownloads()
					if ui.Switch(c.Key("update-downloads"), &on).Label("Download and install automatically").Disabled(!updateChecks()).Changed() {
						setUpdateDownloads(on)
					}
				})
				row("Current version "+mygo.App.Version(), "", func() {
					if ui.Button(c.Key("check-updates"), "Check for Updates").Clicked() {
						checkForUpdates()
					}
				})
			})
		}
	})
}
