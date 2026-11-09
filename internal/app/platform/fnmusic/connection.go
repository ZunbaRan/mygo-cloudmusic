package fnmusic

import (
	"context"
	appplatform "netease-music-mygo/internal/app/platform"
	backend "netease-music-mygo/internal/platform/fnmusic"
	"strings"

	"github.com/egoist/mygo/ui"
)

func (h *Home) disconnect() {
	h.connectionGen++
	gen := h.connectionGen
	h.connectionOpen, h.connectionBusy, h.connectionError = false, false, ""
	if h.runtime != nil {
		go func() {
			err := h.runtime.Disconnect(context.Background())
			h.host.Post(func() {
				if gen != h.connectionGen {
					return
				}
				if err != nil {
					h.connectionError = err.Error()
				}
				h.ResetNavigation(Route{Page: PageLibrary})
				h.host.Invalidate()
			})
		}()
	}
}

func displayAddress(cfg appplatform.ConnectionConfig) string {
	addr := strings.TrimSpace(cfg.Address)
	if cfg.Mode == backend.ModeFNConnect {
		return "FN ID " + addr
	}
	if strings.Contains(addr, "://") {
		return strings.TrimRight(addr, "/")
	}
	if cfg.HTTPS {
		return "https://" + addr
	}
	return "http://" + addr
}

func splitScheme(addr string, https bool) (string, bool) {
	addr = strings.TrimSpace(addr)
	lower := strings.ToLower(addr)
	switch {
	case strings.HasPrefix(lower, "https://"):
		return strings.TrimRight(addr[len("https://"):], "/"), true
	case strings.HasPrefix(lower, "http://"):
		return strings.TrimRight(addr[len("http://"):], "/"), false
	default:
		return strings.TrimRight(addr, "/"), https
	}
}

func (h *Home) openConnection() {
	if h == nil || h.state == nil || h.runtime == nil {
		return
	}
	if h.connectionBusy && h.runtime != nil {
		_ = h.runtime.Cancel(context.Background())
	}
	cfg := h.runtime.Config()
	h.connectionGen++
	address, https := splitScheme(cfg.Address, cfg.HTTPS)
	h.formMode = cfg.Mode
	if h.formMode == backend.ModeAddress && strings.TrimSpace(cfg.Address) == "" {
		h.formMode = backend.ModeFNConnect
	}
	h.formAddress, h.formUser, h.formCode = address, cfg.Username, cfg.AccessCode
	h.formHTTPS, h.formPassword = https, ""
	h.connectionError, h.connectionBusy, h.connectionOpen = "", false, true
	h.host.OpenSettings()
}

func (h *Home) submitConnection() {
	if h.connectionBusy {
		return
	}
	address := strings.TrimSpace(h.formAddress)
	user := strings.TrimSpace(h.formUser)
	if address == "" || user == "" {
		what := "FN ID"
		if h.formMode != backend.ModeFNConnect {
			what = "Server Address"
		}
		h.connectionError = "Enter the " + what + " and username"
		return
	}
	cfg := h.runtime.Config()
	cfg.Address, cfg.Username, cfg.Mode, cfg.HTTPS, cfg.AccessCode = address, user, h.formMode, h.formHTTPS, strings.TrimSpace(h.formCode)
	if h.formPassword != "" {
		cfg.PassHash = backend.HashPassword(h.formPassword)
	}
	if address == "" || user == "" || (h.formPassword == "" && cfg.PassHash == "") {
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
			h.connectionOpen, h.formPassword, h.formCode = false, "", ""
			h.host.Invalidate()
		})
	}()
}

func (h *Home) connectionDialog(c *ui.Context) {
	if h == nil || !h.connectionOpen {
		return
	}
	t := c.Theme()
	ui.DialogBase(c.Key("fnmusic-connection-dialog"), &h.connectionOpen, func(backdrop, panel ui.Element) {
		backdrop.Background(ui.RGBA(0, 0, 0, 0.35))
		panel.Width(380).Padding(24).Radius(18).Gap(14).Background(t.Background).Border(1, t.Border).
			Shadow(0, 24, 60, -12, ui.RGBA(0, 0, 0, 0.4)).
			Transition(ui.ElementTransition{Animation: &ui.Snappy, Enter: &ui.Motion{Scale: 0.94}, Exit: &ui.Motion{Scale: 0.96}})
		ui.Text(c, "Connect to FN Music").FontSize(17).FontWeight(700)
		sub := "Enter an FN ID to connect from home or away; the account is your FN Music account on the NAS"
		if h.formMode != backend.ModeFNConnect {
			sub = "Enter the FN Music address (IP or domain, port optional) and your FN Music account on the NAS"
		}
		ui.Text(c, sub).FontSize(12.5).TextColor(t.TextMuted).LineHeight(1.4).Margin(-8, 0, 0, 0)
		mode := 0
		if h.formMode != backend.ModeFNConnect {
			mode = 1
		}
		if h.host.Segmented(c, "fnmusic-dialog-mode", &mode, "FN ID", "Address") {
			if mode == 0 {
				h.formMode = backend.ModeFNConnect
			} else {
				h.formMode = backend.ModeAddress
			}
			h.formAddress, h.connectionError = "", ""
		}
		field := func(key, label, hint string, value *string, secret bool) {
			ui.Column(c.Key("fnmusic-dialog-" + key)).Gap(5).Children(func() {
				ui.Text(c, label).FontSize(12).FontWeight(600).TextColor(t.TextMuted)
				in := ui.TextInput(c.Key("fnmusic-dialog-"+key+"-input"), value).Placeholder(hint).Label(label).FontSize(14)
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
		if h.formMode != backend.ModeFNConnect {
			ui.Column(c.Key("fnmusic-dialog-protocol")).Gap(5).Children(func() {
				ui.Text(c, "Protocol").FontSize(12).FontWeight(600).TextColor(t.TextMuted)
				choice := 0
				if h.formHTTPS {
					choice = 1
				}
				if h.host.Segmented(c, "fnmusic-dialog-scheme", &choice, "HTTP", "HTTPS") {
					h.formHTTPS, h.connectionError = choice == 1, ""
				}
			})
			field("address", "Server Address", "e.g. 192.168.1.10:5666 or nas.example.com", &h.formAddress, false)
		} else {
			field("address", "FN ID", "e.g. mynas01", &h.formAddress, false)
		}
		field("user", "Username", "FN Music account", &h.formUser, false)
		field("password", "Password", "Password used to connect", &h.formPassword, true)
		field("code", "Access Code (optional)", "Required if the NAS has an access code", &h.formCode, true)
		if h.connectionError != "" {
			ui.Text(c, h.connectionError).FontSize(12.5).TextColor(t.Accent)
		}
		ui.Row(c).Gap(10).Justify(ui.End).Margin(4, 0, 0, 0).Children(func() {
			if h.host.PillButton(c, "fnmusic-dialog-cancel", nil, "Cancel", false).Disabled(h.connectionBusy).Clicked() {
				h.connectionOpen = false
			}
			label := "Connect"
			if h.connectionBusy {
				label = "Connecting…"
			}
			ok := h.host.PillButton(c, "fnmusic-dialog-submit", iconServer, label, true).Disabled(h.connectionBusy)
			if ok.Clicked() {
				h.submitConnection()
			}
		})
	})
}

func (h *Home) connectionPage(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Fill().Center().Gap(10).Children(func() {
		ui.Text(c, "Connect to FN Music").FontSize(20).FontWeight(700)
		ui.Text(c, "Play the library, playlists and liked songs on your FN NAS").FontSize(13).TextColor(t.TextMuted)
		ui.Text(c, "Works with FN ID or a LAN address; MP3 and FLAC play directly, other formats are transcoded by the NAS").FontSize(12).TextColor(t.TextMuted)
		if err := h.state.Error(); err != nil {
			ui.Text(c, err.Error()).FontSize(12.5).TextColor(t.Accent)
		}
		if h.host.PillButton(c, "open-fnmusic-settings", iconServer, "Open settings and connect", true).Clicked() {
			h.host.OpenSettings()
		}
	})
}
