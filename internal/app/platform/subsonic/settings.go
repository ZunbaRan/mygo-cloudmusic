package subsonic

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	appplatform "netease-music-mygo/internal/app/platform"
)

func (h *Home) Settings(c *ui.Context) {
	if h == nil || h.host == nil || h.state == nil {
		return
	}
	t := *c.Theme()
	c.Root().Background(t.Background)
	prof := h.state.Profile()
	ui.Scroll(c).Fill().Padding(24, 28, 32, 28).Gap(8).Children(func() {
		ui.Text(c, prof.Name).FontSize(12).FontWeight(600).TextColor(t.TextMuted).Margin(14, 4, 4, 4)
		ui.Column(c).Radius(12).Background(t.Background).Border(1, t.Border).Dividers(1, t.Border).Children(func() {
			var cfg appplatform.ConnectionConfig
			if h.runtime != nil {
				cfg = h.runtime.Config()
			}
			ui.Row(c).Padding(12, 16).Gap(16).AlignItems(ui.Center).Children(func() {
				ui.Column(c).Grow(1).Gap(3).Children(func() {
					name := "Not connected"
					desc := "Once connected you can play the library, albums, artists and playlists"
					if h.state.Connected() {
						name = cfg.Username
						if name == "" {
							name = prof.Name
						}
						serverDetail := cfg.Address
						if info := h.state.Client().ServerInfo(); info != nil && info.ServerVersion != "" {
							serverDetail += " · " + info.Type + " " + info.ServerVersion
						}
						desc = "Connected · " + serverDetail
					}
					if err := h.state.Error(); err != nil {
						desc = "Connection failed: " + err.Error()
					}
					ui.Text(c, name).FontSize(13.5).FontWeight(500)
					ui.Text(c, desc).FontSize(12).TextColor(t.TextMuted)
				})
				label := "Connecting…"
				if h.state.Connected() {
					label = "Reconnect…"
				}
				if h.host.PillButton(c, string(prof.ID)+"-connection-open", iconServer, label, true).Clicked() {
					h.openConnection()
				}
				if h.state.Connected() && h.host.PillButton(c, string(prof.ID)+"-disconnect", iconSignOut, "Disconnect", false).Clicked() {
					h.disconnect()
				}
			})
		})
		ui.Text(c, "Playback & Bitrate").FontSize(12).FontWeight(600).TextColor(t.TextMuted).Margin(14, 4, 4, 4)
		ui.Column(c).Radius(12).Background(t.Background).Border(1, t.Border).Dividers(1, t.Border).Children(func() {
			h.transcodeSettings(c, t)
		})
		h.host.SharedSettings(c)
	})
	h.connectionDialog(c)
}

func (h *Home) transcodeSettings(c *ui.Context, t ui.Theme) {
	if h.runtime == nil {
		return
	}
	transcode := h.runtime.Transcode()
	mode, bitrate := transcode.Mode, transcode.Bitrate
	ui.Row(c).Padding(12, 16).Gap(16).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Grow(1).Gap(3).Children(func() {
			ui.Text(c, "Playback Bitrate").FontSize(13.5).FontWeight(500)
			ui.Text(c, "Caps the bitrate of streamed playback; the server transcodes to this limit").FontSize(12).TextColor(t.TextMuted)
		})
		values := h.runtime.TranscodeBitrates()
		labels := make([]string, len(values))
		chosen := "Unlimited"
		for i, v := range values {
			labels[i] = "Unlimited"
			if v > 0 {
				labels[i] = fmt.Sprintf("%d kbps", v)
			}
			if v == bitrate {
				chosen = labels[i]
			}
		}
		if ui.Select(c.Key("subsonic-transcode-bitrate"), &chosen, labels).Label("Playback Bitrate").Width(150).Disabled(mode == appplatform.TranscodeNative).Changed() {
			for i, label := range labels {
				if label == chosen {
					h.runtime.SetTranscode(appplatform.TranscodeSettings{Mode: mode, Bitrate: values[i]})
					break
				}
			}
		}
	})
	ui.Row(c).Padding(12, 16).Gap(16).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Grow(1).Gap(3).Children(func() {
			ui.Text(c, "Transcoding").FontSize(13.5).FontWeight(500)
			ui.Text(c, "Auto transcodes to MP3 when the player doesn't support the format").FontSize(12).TextColor(t.TextMuted)
		})
		labels := []string{"Auto", "Original only", "Always"}
		chosen := labels[0]
		if mode == appplatform.TranscodeNative {
			chosen = labels[1]
		} else if mode == appplatform.TranscodeAlways {
			chosen = labels[2]
		}
		if ui.Select(c.Key("subsonic-transcode-mode"), &chosen, labels).Label("Transcoding").Width(150).Changed() {
			m := appplatform.TranscodeAuto
			if chosen == labels[1] {
				m = appplatform.TranscodeNative
			} else if chosen == labels[2] {
				m = appplatform.TranscodeAlways
			}
			h.runtime.SetTranscode(appplatform.TranscodeSettings{Mode: m, Bitrate: bitrate})
		}
	})
}
