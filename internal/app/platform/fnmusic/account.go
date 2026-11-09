package fnmusic

// accountActions builds the provider account menu. The shared host owns only
// the row and popover chrome; labels, icons and lifecycle actions stay with
// this platform implementation.
func (h *Home) accountActions() []AccountAction {
	openSettings := func() {
		if h.host != nil {
			h.host.OpenSettings()
		}
	}
	settings := AccountAction{Icon: iconSettings, Label: "Settings…", Shortcut: "Ctrl+,", Action: openSettings}
	if h.state != nil && h.state.Connecting() {
		return []AccountAction{settings}
	}
	if h.state != nil && (h.state.Connected() || h.state.Error() != nil) {
		return []AccountAction{
			settings,
			{Icon: iconReload, Label: "Reconnect…", Separator: true, Action: h.openConnection},
			{Icon: iconSignOut, Label: "Disconnect FN Music", Danger: true, Action: h.disconnect},
		}
	}
	return []AccountAction{
		settings,
		{Icon: iconReload, Label: "Connecting to FN Music…", Separator: true, Action: h.openConnection},
	}
}
