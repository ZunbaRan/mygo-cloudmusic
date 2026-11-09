package app

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"netease-music-mygo/internal/netease"
)

// editTestApp is signed in, with liked songs, two playlists of their own
// and one saved, and shows their playlist 10.
func editTestApp() *app {
	a := newTestApp()
	a.settings = defaultSettings()
	a.profile = &netease.Profile{UserID: 1, Nickname: "Xiaoming"}
	a.mine = []netease.Playlist{
		{ID: 5, Name: "Xiaoming's Liked Songs", SpecialType: 5},
		{ID: 10, Name: "Commute"},
		{ID: 11, Name: "Bedtime"},
	}
	a.saved = []netease.Playlist{{ID: 20, Name: "Rising"}}
	a.playlists[10] = &netease.Playlist{ID: 10, Name: "Commute", Tracks: []netease.Track{
		{ID: 1, Name: "A Photo", Artists: []netease.Artist{{Name: "Yoga Lin"}}, Duration: 200000},
		{ID: 2, Name: "Takarabune", Artists: []netease.Artist{{Name: "DOUDOU"}}, Duration: 200000},
	}}
	a.route = route{page: pagePlaylist, id: 10}
	return a
}

// Choosing one of the listener's playlists from a song's menu adds the
// song to it, and the menu offers neither the liked songs nor the
// playlist the song is listed in.
func TestAddSongToPlaylist(t *testing.T) {
	type add struct{ pid, song int64 }
	calls := make(chan add, 1)
	addToPlaylist = func(pid int64, songs ...int64) error { calls <- add{pid, songs[0]}; return nil }
	defer func() { addToPlaylist = netease.AddToPlaylist }()

	a := editTestApp()
	tt := ui.NewTester(a.view, 1200, 800)
	tt.Frame()
	if err := tt.RightClick("Takarabune"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Add to Playlist", "Commute"); err == nil {
		t.Error("offered to add the song to the playlist it is in")
	}
	if err := tt.ChooseMenuItem("Add to Playlist", "Xiaoming's Liked Songs"); err == nil {
		t.Error("offered to add the song to the liked songs")
	}
	if err := tt.ChooseMenuItem("Add to Playlist", "Bedtime"); err != nil {
		t.Fatalf("%v; the menu shows %q", err, tt.Menu())
	}
	tt.Frame()
	if got := <-calls; got != (add{11, 2}) {
		t.Errorf("added %+v, want song 2 to playlist 11", got)
	}
}

// Removing a song from the listener's playlist takes its row away at once,
// and asks NetEase to take it out.
func TestRemoveSongFromPlaylist(t *testing.T) {
	type del struct{ pid, song int64 }
	calls := make(chan del, 1)
	removeFromPlaylist = func(pid int64, songs ...int64) error { calls <- del{pid, songs[0]}; return nil }
	defer func() { removeFromPlaylist = netease.RemoveFromPlaylist }()

	a := editTestApp()
	// Playing the id: the queue keeps the song.
	a.pl.Cue(a.playlists[10].Tracks)
	tt := ui.NewTester(a.view, 1200, 800)
	tt.Frame()
	if err := tt.RightClick("Takarabune"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Remove from Playlist"); err != nil {
		t.Fatalf("%v; the menu shows %q", err, tt.Menu())
	}
	tt.Frame()
	if got := <-calls; got != (del{10, 2}) {
		t.Errorf("removed %+v, want song 2 from playlist 10", got)
	}
	if _, ok := tt.Find("Takarabune"); ok {
		t.Error("the song's row is still listed")
	}
	if q := a.pl.Queue(); len(q) != 2 || q[1].ID != 2 {
		t.Errorf("removing the song changed the queue playing: %v", q)
	}
}

// A playlist the listener saved, not made, offers no removing.
func TestNoRemovingFromSavedPlaylist(t *testing.T) {
	a := editTestApp()
	a.playlists[20] = &netease.Playlist{ID: 20, Name: "Rising", Tracks: a.playlists[10].Tracks}
	a.route = route{page: pagePlaylist, id: 20}
	tt := ui.NewTester(a.view, 1200, 800)
	tt.Frame()
	if err := tt.RightClick("Takarabune"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Remove from Playlist"); err == nil {
		t.Error("offered to remove a song from a saved playlist")
	}
}

// The sidebar's plus asks for a name, and Enter makes the playlist so
// named.
func TestCreatePlaylist(t *testing.T) {
	names := make(chan string, 1)
	createPlaylist = func(name string) (*netease.Playlist, error) {
		names <- name
		return &netease.Playlist{ID: 12, Name: name}, nil
	}
	defer func() { createPlaylist = netease.CreatePlaylist }()

	a := editTestApp()
	tt := ui.NewTester(a.view, 1200, 800)
	tt.Frame()
	if err := tt.Click("New Playlist"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(300 * time.Millisecond)
	tt.Frame()
	tt.Type("  周末  ")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if got := <-names; got != "周末" {
		t.Errorf("made a playlist named %q, want 周末", got)
	}
}

// Choosing New Playlist… from a song's menu makes the playlist, then adds the
// song to it.
func TestCreatePlaylistForSong(t *testing.T) {
	a := editTestApp()
	tt := ui.NewTester(a.view, 1200, 800)
	tt.Frame()
	if err := tt.RightClick("Takarabune"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Add to Playlist", "New Playlist…"); err != nil {
		t.Fatalf("%v; the menu shows %q", err, tt.Menu())
	}
	tt.Frame()
	if !a.naming.open || a.naming.add == nil || a.naming.add.ID != 2 {
		t.Fatalf("the dialog is %+v, want it open to add song 2", a.naming)
	}
}

// Renaming a playlist from the sidebar's menu shows its new name at once,
// and asks NetEase to rename it.
func TestRenamePlaylist(t *testing.T) {
	type rename struct {
		id   int64
		name string
	}
	calls := make(chan rename, 1)
	renamePlaylist = func(id int64, name string) error { calls <- rename{id, name}; return nil }
	defer func() { renamePlaylist = netease.RenamePlaylist }()

	a := editTestApp()
	a.route = route{page: pageHome}
	a.homeLoading = true
	tt := ui.NewTester(a.view, 1200, 800)
	tt.Frame()
	if err := tt.RightClick("Bedtime"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Rename…"); err != nil {
		t.Fatalf("%v; the menu shows %q", err, tt.Menu())
	}
	tt.Advance(300 * time.Millisecond)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("Sleep")
	if err := tt.Click("Save"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if got := <-calls; got != (rename{11, "Sleep"}) {
		t.Errorf("renamed %+v, want playlist 11 to Sleep", got)
	}
	if a.mine[2].Name != "Sleep" {
		t.Errorf("the sidebar names it %q", a.mine[2].Name)
	}
	if _, ok := tt.Find("Sleep"); !ok {
		t.Error("the new name does not show")
	}
}

// The liked songs can't be renamed: they offer no menu.
func TestNoRenamingLikedSongs(t *testing.T) {
	a := editTestApp()
	a.route = route{page: pageHome}
	a.homeLoading = true
	tt := ui.NewTester(a.view, 1200, 800)
	tt.Frame()
	tt.RightClick("Xiaoming's Liked Songs")
	if err := tt.ChooseMenuItem("Rename…"); err == nil {
		t.Error("offered to rename the liked songs")
	}
}

// Deleting the playlist shown, once confirmed, takes it from the sidebar
// and goes back to the page before, with no way forward to it.
func TestDeletePlaylist(t *testing.T) {
	ids := make(chan int64, 1)
	deletePlaylist = func(id int64) error { ids <- id; return nil }
	defer func() { deletePlaylist = netease.DeletePlaylist }()

	a := editTestApp()
	a.homeLoading = true
	a.route = route{page: pageCharts}
	a.open(route{page: pagePlaylist, id: 10})
	tt := ui.NewTester(a.view, 1200, 800)
	tt.Frame()
	if err := tt.Click("More"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Delete Playlist…"); err != nil {
		t.Fatalf("%v; the menu shows %q", err, tt.Menu())
	}
	tt.Advance(300 * time.Millisecond)
	tt.Frame()
	select {
	case <-ids:
		t.Fatal("deleted before it was confirmed")
	default:
	}
	if err := tt.Click("Delete"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if got := <-ids; got != 10 {
		t.Errorf("deleted playlist %d, want 10", got)
	}
	if a.editable(10) != nil {
		t.Error("the sidebar still lists it")
	}
	if a.route.page != pageCharts || len(a.history) != 0 || len(a.forward) != 0 {
		t.Errorf("on %+v, history %v, forward %v; want the charts with nowhere to go", a.route, a.history, a.forward)
	}
}

// Cancelling the confirmation deletes nothing.
func TestCancelDeletingPlaylist(t *testing.T) {
	deletePlaylist = func(int64) error { t.Error("deleted the playlist"); return nil }
	defer func() { deletePlaylist = netease.DeletePlaylist }()

	a := editTestApp()
	a.route = route{page: pageHome}
	a.homeLoading = true
	tt := ui.NewTester(a.view, 1200, 800)
	tt.Frame()
	if err := tt.RightClick("Bedtime"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Delete Playlist…"); err != nil {
		t.Fatalf("%v; the menu shows %q", err, tt.Menu())
	}
	tt.Advance(300 * time.Millisecond)
	tt.Frame()
	if err := tt.Click("Cancel"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(300 * time.Millisecond)
	tt.Frame()
	if a.editable(11) == nil {
		t.Error("the playlist left the sidebar")
	}
}

// With NM_FRAMES set, saves the listener's playlist with its sidebar, and
// the dialogs that name and delete playlists, to look at.
func TestPlaylistEditFrames(t *testing.T) {
	dir := os.Getenv("NM_FRAMES")
	if dir == "" {
		t.Skip()
	}
	a := editTestApp()
	tt := ui.NewTester(a.view, 1200, 800)
	save := func(name string) {
		tt.Advance(400 * time.Millisecond)
		tt.Frame()
		f, _ := os.Create(filepath.Join(dir, name+".png"))
		defer f.Close()
		png.Encode(f, tt.Image())
	}
	for _, dark := range []bool{false, true} {
		suffix := "-light"
		if dark {
			suffix = "-dark"
		}
		tt.SetDark(dark)
		save("playlist" + suffix)
		a.newPlaylist(nil)
		save("create" + suffix)
		a.naming.open = false
		a.startRename(a.editable(11))
		save("rename" + suffix)
		a.naming.open = false
		a.confirmDelete(a.editable(11))
		save("delete" + suffix)
		a.deleting.open = false
	}
}
