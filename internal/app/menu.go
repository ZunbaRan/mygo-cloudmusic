package app

import (
	"runtime"

	"github.com/egoist/mygo"
)

// setMenu gives the app its menu bar, in English and named as the app
// is, with the settings in the app menu and the player's controls in a
// menu of their own.
func (a *app) setMenu() {
	run := func(fn func()) func(*mygo.MenuItem, *mygo.Window) {
		return func(*mygo.MenuItem, *mygo.Window) {
			fn()
			a.redraw()
		}
	}
	name := mygo.App.Name()
	appMenu := []*mygo.MenuItem{
		{Role: mygo.RoleAbout, Label: "About " + name},
		{Label: "Checking…", Click: run(checkForUpdates)},
		mygo.Separator(),
		{Label: "Settings…", Accelerator: "CmdOrCtrl+,", Click: run(a.openSettings)},
		mygo.Separator(),
	}
	// Services and hiding are macOS's alone. Elsewhere MyGo leaves their
	// items out but keeps the separators around them, which then stack.
	if runtime.GOOS == "darwin" {
		appMenu = append(appMenu,
			&mygo.MenuItem{Role: mygo.RoleServices, Label: "Services"},
			mygo.Separator(),
			&mygo.MenuItem{Role: mygo.RoleHide, Label: "Hide " + name},
			&mygo.MenuItem{Role: mygo.RoleHideOthers, Label: "Hide Others"},
			&mygo.MenuItem{Role: mygo.RoleUnhide, Label: "Show All"},
			mygo.Separator(),
		)
	}
	appMenu = append(appMenu, &mygo.MenuItem{Role: mygo.RoleQuit, Label: "Quit " + name})
	mygo.App.SetMenu(mygo.NewMenu([]*mygo.MenuItem{
		{Label: name, Submenu: appMenu},
		{Label: "File", Submenu: []*mygo.MenuItem{
			{Label: "New Playlist…", Accelerator: "CmdOrCtrl+N", Click: run(func() {
				a.showMain()
				a.nowPlaying = false
				a.newPlaylist(nil)
			})},
		}},
		{Label: "Edit", Submenu: []*mygo.MenuItem{
			{Role: mygo.RoleUndo, Label: "Undo"},
			{Role: mygo.RoleRedo, Label: "Redo"},
			mygo.Separator(),
			{Role: mygo.RoleCut, Label: "Cut"},
			{Role: mygo.RoleCopy, Label: "Copy"},
			{Role: mygo.RolePaste, Label: "Paste"},
			{Role: mygo.RoleSelectAll, Label: "Select All"},
		}},
		{Label: "Controls", Submenu: []*mygo.MenuItem{
			{Label: "Play/Pause", Click: run(func() { a.pl.TogglePlay() })},
			{Label: "Next", Accelerator: "CmdOrCtrl+Right", Click: run(func() { a.pl.Next() })},
			{Label: "Previous", Accelerator: "CmdOrCtrl+Left", Click: run(func() { a.pl.Prev() })},
			mygo.Separator(),
			{Label: "Like/Unlike", Accelerator: "CmdOrCtrl+D", Click: run(a.likeCurrent)},
			mygo.Separator(),
			{Label: "Shuffle", Click: run(func() { a.pl.ToggleShuffle() })},
			{Label: "Cycle Repeat Mode", Click: run(func() { a.pl.CycleRepeat() })},
			mygo.Separator(),
			{Label: "Now Playing", Accelerator: "CmdOrCtrl+L", Click: run(func() {
				a.showMain()
				a.nowPlaying = !a.nowPlaying && a.pl.Current() != nil
			})},
		}},
		{Label: "Window", Submenu: []*mygo.MenuItem{
			{Role: mygo.RoleMinimize, Label: "Minimize"},
			{Role: mygo.RoleZoom, Label: "Zoom"},
			{Role: mygo.RoleClose, Label: "Close"},
			mygo.Separator(),
			{Label: name, Accelerator: "CmdOrCtrl+0", Click: run(a.showMain)},
			{Role: mygo.RoleFront, Label: "Bring All to Front"},
		}},
	}))
}
