package songloft

import (
	"context"
	"strings"

	"github.com/egoist/mygo/ui"
	backend "netease-music-mygo/internal/platform/songloft"
)

func (h *Home) submitConnection() {
	if h.connectionBusy {
		return
	}
	address := strings.TrimSpace(h.formAddress)
	user := strings.TrimSpace(h.formUser)
	if address == "" || user == "" {
		h.connectionError = "Enter the server address and username"
		return
	}
	cfg := h.runtime.Config()
	cfg.Address = address
	cfg.Username = user
	if h.formPassword != "" {
		cfg.PassHash = h.formPassword
	}
	if cfg.Address == "" || cfg.Username == "" || cfg.PassHash == "" {
		h.connectionError = "Enter the address, username and password"
		return
	}
	h.connectionBusy, h.connectionError = true, ""
	if h.runtime == nil {
		h.connectionBusy = false
		h.connectionError = "Platform not initialized"
		return
	}
	h.connectionGen++
	gen := h.connectionGen
	go func() {
		err := h.runtime.Connect(context.Background(), cfg)
		h.host.Post(func() {
			if gen != h.connectionGen {
				return
			}
			h.connectionBusy = false
			if err != nil {
				h.connectionError = err.Error()
				return
			}
			h.connectionOpen, h.formPassword = false, ""
			h.host.Invalidate()
		})
	}()
}

func (h *Home) connectionDialog(c *ui.Context) {
	if h == nil || !h.connectionOpen {
		return
	}
	t := c.Theme()
	ui.DialogBase(c.Key("connection-dialog-songloft"), &h.connectionOpen, func(backdrop, panel ui.Element) {
		backdrop.Background(ui.RGBA(0, 0, 0, 0.35))
		panel.Width(380).Padding(24).Radius(18).Gap(14).Background(t.Background).Border(1, t.Border).
			Shadow(0, 24, 60, -12, ui.RGBA(0, 0, 0, 0.4))
		ui.Text(c, "Connect to a Songloft server").FontSize(17).FontWeight(700)
		ui.Text(c, backend.DefaultProfile.Description).FontSize(12.5).TextColor(t.TextMuted).LineHeight(1.4).Margin(-8, 0, 0, 0)

		field := func(key, label, hint string, value *string, secret bool) {
			ui.Column(c.Key("songloft-dialog-" + key)).Gap(5).Children(func() {
				ui.Text(c, label).FontSize(12).FontWeight(600).TextColor(t.TextMuted)
				in := ui.TextInput(c.Key("songloft-dialog-"+key+"-input"), value).Placeholder(hint).Label(label).FontSize(14)
				if secret {
					in.Password()
				}
				if key == "address" {
					in.AutoFocus()
				}
				if in.Submitted() {
					h.submitConnection()
				}
			})
		}

		field("address", "Server Address", backend.DefaultProfile.DefaultHint, &h.formAddress, false)
		field("user", "Username", "Songloft account username", &h.formUser, false)
		field("password", "Password", "Songloft account password", &h.formPassword, true)

		if h.connectionError != "" {
			ui.Text(c, h.connectionError).FontSize(12.5).TextColor(ui.Hex("#ef4444"))
		}

		ui.Row(c).Gap(10).Justify(ui.End).Margin(4, 0, 0, 0).Children(func() {
			if h.host.PillButton(c, "songloft-dialog-cancel", nil, "Cancel", false).Disabled(h.connectionBusy).Clicked() {
				h.connectionOpen = false
			}
			label := "Connect"
			if h.connectionBusy {
				label = "Connecting…"
			}
			ok := h.host.PillButton(c, "songloft-dialog-submit", iconServer, label, true).Disabled(h.connectionBusy)
			if ok.Clicked() {
				h.submitConnection()
			}
		})
	})
}

func (h *Home) connectionPage(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Fill().Center().Gap(10).Children(func() {
		ui.Text(c, "Connect to a Songloft server").FontSize(20).FontWeight(700)
		ui.Text(c, backend.DefaultProfile.Description).FontSize(13).TextColor(t.TextMuted)
		ui.Text(c, "Streams original quality (FLAC, MP3, AAC, OGG…) with lyrics and covers").FontSize(12).TextColor(t.TextMuted)
		if err := h.state.Error(); err != nil {
			ui.Text(c, err.Error()).FontSize(12.5).TextColor(ui.Hex("#ef4444"))
		}
		ui.Row(c).Gap(10).Margin(10, 0, 0, 0).Children(func() {
			if h.host.PillButton(c, "connect-songloft-now", iconServer, "Connect to Server", true).Clicked() {
				h.openConnection()
			}
			if h.host.PillButton(c, "open-songloft-settings", iconServer, "Open Settings", false).Clicked() {
				h.host.OpenSettings()
			}
		})
	})
}

func (h *Home) dialogs(c *ui.Context) {
	h.playlistEditDialog(c)
	h.playlistDeleteDialog(c)
}
