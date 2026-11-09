package app

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/egoist/mygo/ui"
	"netease-music-mygo/internal/player/mpv"
)

func (a *app) syncMPVSettings() {
	if a.pl == nil {
		return
	}
	a.pl.SetUseMPV(a.settings.UseMPV)
	a.pl.SetMPVConfig(mpv.Config{
		Exclusive:        a.settings.MPVExclusive,
		ReplayGain:       a.settings.MPVReplayGain,
		AudioDevice:      a.settings.MPVAudioDevice,
		Gapless:          a.settings.MPVGapless,
		EqualizerEnabled: a.settings.EqualizerEnabled,
		EqualizerPreamp:  a.settings.EqualizerPreamp,
		EqualizerGains:   a.settings.EqualizerGains,
	})
	a.pl.SetEqualizer(a.settings.EqualizerPreamp, a.settings.EqualizerGains, a.settings.EqualizerEnabled)
}

func (a *app) mpvReady() bool {
	return mpv.Ready()
}

func (a *app) startMPVDownload() {
	if a.mpvDownloading {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.mpvCancel = cancel
	a.mpvDownloading = true
	a.mpvDone, a.mpvTotal = 0, 0
	a.changed()

	go func() {
		err := mpv.Download(ctx, func(done, total int64) {
			a.post(func() {
				a.mpvDone = done
				a.mpvTotal = total
			})
		})
		a.post(func() {
			a.mpvDownloading = false
			a.mpvCancel = nil
			if err == nil {
				_ = mpv.Load()
				a.syncMPVSettings()
				a.toast = "Decoder component downloaded"
			} else if ctx.Err() == nil {
				a.toast = fmt.Sprintf("Download failed: %v", err)
			}
			a.changed()
		})
	}()
}

func (a *app) cancelMPVDownload() {
	if a.mpvCancel != nil {
		a.mpvCancel()
		a.mpvCancel = nil
	}
	a.mpvDownloading = false
	a.changed()
}

func (a *app) deleteMPV() {
	if a.pl != nil {
		a.pl.SetUseMPV(false)
	}
	a.settings.UseMPV = false
	_ = mpv.Delete()
	a.changed()
	a.syncMPVSettings()
	a.toast = "Decoder component removed"
}

func (a *app) scanDLNADevices() {
	if a.pl == nil {
		return
	}
	ctrl := a.pl.DLNA()
	if ctrl == nil || a.dlnaScanning {
		return
	}
	a.dlnaScanning = true
	a.redraw()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		_, _ = ctrl.RefreshDevices(ctx)
		a.post(func() {
			a.dlnaScanning = false
			a.redraw()
		})
	}()
}

// renderHifiSettings renders the "Advanced Audio & Hardware Decoding" section into any settings page (both NetEase and third-party platforms).
func (a *app) renderHifiSettings(c *ui.Context, group func(title string, rows func()), row func(key, label, desc string, control func())) {
	s := &a.settings
	group("Advanced Audio & Hardware Decoding", func() {
		row("hifi-mpv-toggle", "Advanced Decoder Engine (MPV)", "An MPV-based audiophile decoding engine for native playback of DSD (DSF/DFF), APE, ALAC, AAC, Hi-Res FLAC and more", func() {
			if ui.Switch(c.Key("use-mpv"), &s.UseMPV).Label("Advanced Decoder").Disabled(!a.mpvReady()).Changed() {
				a.changed()
				a.syncMPVSettings()
			}
		})

		var statusDesc string
		switch {
		case a.mpvDownloading:
			statusDesc = fmt.Sprintf("Downloading libmpv (%s / %s)...", megabytes(a.mpvDone), megabytes(a.mpvTotal))
		case mpv.IsDownloaded():
			statusDesc = "Component ready (standalone component cached)"
		case mpv.SystemPath() != "":
			statusDesc = "System libmpv found (" + filepath.Base(mpv.SystemPath()) + "); a standalone component can also be downloaded"
		default:
			statusDesc = "Component not downloaded — download to enable the advanced decoder"
		}

		row("hifi-mpv-status", "Decoder Component", statusDesc, func() {
			switch {
			case a.mpvDownloading:
				if ui.Button(c.Key("mpv-dl"), "Cancel").Clicked() {
					a.cancelMPVDownload()
				}
			case mpv.IsDownloaded():
				if ui.Button(c.Key("mpv-dl"), "Remove").Clicked() {
					a.deleteMPV()
				}
			default:
				if ui.Button(c.Key("mpv-dl"), "Download").Clicked() {
					a.startMPVDownload()
				}
			}
		})

		if s.UseMPV && a.mpvReady() {
			devices := a.pl.AudioDevices()
			if len(devices) > 0 {
				var devLabels []string
				chosenDevLabel := "System Default (Auto)"
				devNameToLabel := make(map[string]string)
				labelToDevName := make(map[string]string)
				labelCounts := make(map[string]int)

				for _, d := range devices {
					lbl := d.Description
					if lbl == "" {
						lbl = d.Name
					}
					if d.Name == "auto" {
						lbl = "System Default (Auto)"
					}
					labelCounts[lbl]++
				}

				for _, d := range devices {
					lbl := d.Description
					if lbl == "" {
						lbl = d.Name
					}
					if d.Name == "auto" {
						lbl = "System Default (Auto)"
					} else if labelCounts[lbl] > 1 {
						lbl = fmt.Sprintf("%s (%s)", lbl, d.Name)
					}
					devLabels = append(devLabels, lbl)
					devNameToLabel[d.Name] = lbl
					labelToDevName[lbl] = d.Name
				}

				if curLbl, ok := devNameToLabel[s.MPVAudioDevice]; ok {
					chosenDevLabel = curLbl
				} else if s.MPVAudioDevice == "" || s.MPVAudioDevice == "auto" {
					chosenDevLabel = "System Default (Auto)"
				}

				row("hifi-mpv-device", "Audio Output Device", "Route audio to a specific device (headphones, onboard sound, or an external USB DAC/amp)", func() {
					if ui.Select(c.Key("mpv-audio-device"), &chosenDevLabel, devLabels).Label("Audio Output Device").Width(240).Changed() {
						if targetName, ok := labelToDevName[chosenDevLabel]; ok {
							s.MPVAudioDevice = targetName
							a.changed()
							a.syncMPVSettings()
						}
					}
				})
			}

			row("hifi-mpv-exclusive", "Exclusive Audio Output", "Bypasses the system mixer for bit-perfect passthrough output", func() {
				if ui.Switch(c.Key("mpv-exclusive"), &s.MPVExclusive).Label("Exclusive Mode").Changed() {
					a.changed()
					a.syncMPVSettings()
				}
			})
			row("hifi-mpv-gapless", "Gapless Playback", "Keeps the audio stream continuous across track changes, removing gaps and pops", func() {
				if ui.Switch(c.Key("mpv-gapless"), &s.MPVGapless).Label("Gapless Playback").Changed() {
					a.changed()
					a.syncMPVSettings()
				}
			})

			row("hifi-mpv-eq-info", "Graphic Equalizer (10-Band EQ)", "A 10-band graphic mixer on professional audio filters, integrated into the right of the player bar for quick access", func() {
				ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Icon(c, iconEqualizer).Size(14, 14).TextColor(c.Theme().TextMuted)
					ui.Text(c, "Quick entry enabled in the player bar").FontSize(12).TextColor(c.Theme().TextMuted)
				})
			})
		}
	})
}

func (a *app) drawEqSlider(c *ui.Context, key string, val *float64, label string) {
	const knob, track = 16, 4
	ui.Column(c).AlignItems(ui.Center).Gap(3).Children(func() {
		ui.Text(c, fmt.Sprintf("%+.0f", *val)).FontSize(10).TextColor(c.Theme().TextMuted).SingleLine()
		s := ui.SliderBase(c.Key(key), val, -12, 12).Vertical().Size(knob+10, 84).PaddingY(knob / 2).FocusRing(false).Label(label)
		if s.Changed() {
			a.settings.EqualizerPreset = "custom"
			a.changed()
			a.syncMPVSettings()
		}
		v := float32((*val - (-12)) / (12 - (-12)))
		accent := a.pal.accent
		dark := c.Theme().Dark
		trackColor := ui.RGBA(128, 128, 128, 0.25)
		knobColor := ui.RGBA(255, 255, 255, 0.95)
		if dark {
			knobColor = ui.RGBA(220, 220, 220, 0.95)
		}
		pressed := s.Pressed()
		s.Draw(func(p *ui.Painter, r ui.Rect) {
			top, foot := r.Y+knob/2, r.Y+r.H-knob/2
			x := r.X + (r.W-track)/2
			y := foot - (foot-top)*v
			p.Fill(ui.Rect{X: x, Y: top - track/2, W: track, H: foot - top + track}, trackColor, track/2)
			p.Fill(ui.Rect{X: x, Y: y, W: track, H: foot - y + track/2}, accent, track/2)
			k := ui.Rect{X: r.X + (r.W-knob)/2, Y: y - knob/2, W: knob, H: knob}
			p.Shadow(k, knob/2, 0, 1, 3, 0, ui.RGBA(0, 0, 0, 0.2))
			fill := knobColor
			if pressed {
				fill = fill.Mix(ui.RGB(0, 0, 0), 0.1)
			}
			p.Fill(k, fill, knob/2)
			p.Stroke(k, ui.RGBA(0, 0, 0, 0.1), knob/2, 0.5)
		})
		ui.Text(c, label).FontSize(11).SingleLine()
	})
}
