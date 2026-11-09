package fnmusic

import "netease-music-mygo/internal/netease"

func (h *Home) addFolderToPlaylist(folderPath string, pl netease.Playlist) {
	if h == nil || h.state == nil || folderPath == "" {
		return
	}
	gen := h.state.Generation()
	client := h.state.Client()
	go func() {
		err := client.AddFoldersToPlaylist(pl.Ref, []string{folderPath}, nil)
		h.postCurrent(gen, func() {
			if err != nil {
				h.host.Toast("Failed to add to playlist: " + err.Error())
				return
			}
			h.host.Toast("Added to playlist")
		})
	}()
}

func (h *Home) playFolders(paths []string, shuffle bool) {
	if h == nil || h.state == nil || len(paths) == 0 {
		return
	}
	gen := h.state.Generation()
	client := h.state.Client()
	go func() {
		tracks, err := client.PlayableFolderTracks(paths, nil)
		h.postCurrent(gen, func() {
			if err != nil {
				h.host.Toast("Failed to play the folder: " + err.Error())
				return
			}
			if len(tracks) == 0 {
				h.host.Toast("The folder has no playable songs")
				return
			}
			h.fillFolderTracks(tracks)
			if p := h.host.Player(); p != nil {
				// Shuffling turns shuffle on; playing keeps the listener's
				// choice, as a list's play button does.
				if shuffle {
					p.SetShuffle(true)
				}
				p.PlayQueue(tracks, 0)
			}
		})
	}()
}

func (h *Home) deleteFolders(paths []string, permanent bool) {
	if h == nil || h.state == nil || len(paths) == 0 {
		return
	}
	gen := h.state.Generation()
	client := h.state.Client()
	go func() {
		err := client.DeleteFolderItems(paths, nil, permanent)
		h.postCurrent(gen, func() {
			if err != nil {
				h.host.Toast("Failed to delete the folder: " + err.Error())
				return
			}
			h.refreshFolder(h.route.Query)
		})
	}()
}

func (h *Home) deleteTracks(refs []string, permanent bool) {
	if h == nil || h.state == nil || len(refs) == 0 {
		return
	}
	gen := h.state.Generation()
	client := h.state.Client()
	go func() {
		err := client.DeleteFolderItems(nil, refs, permanent)
		h.postCurrent(gen, func() {
			if err != nil {
				h.host.Toast("Failed to delete the song: " + err.Error())
				return
			}
			h.refreshFolder(h.route.Query)
		})
	}()
}

func (h *Home) recoverFolders(paths []string) {
	if h == nil || h.state == nil || len(paths) == 0 {
		return
	}
	gen := h.state.Generation()
	client := h.state.Client()
	go func() {
		err := client.RecoverFolderItems(paths, nil)
		h.postCurrent(gen, func() {
			if err != nil {
				h.host.Toast("Failed to restore the folder: " + err.Error())
				return
			}
			h.refreshFolder(h.route.Query)
		})
	}()
}

func (h *Home) recoverTracks(refs []string) {
	if h == nil || h.state == nil || len(refs) == 0 {
		return
	}
	gen := h.state.Generation()
	client := h.state.Client()
	go func() {
		err := client.RecoverFolderItems(nil, refs)
		h.postCurrent(gen, func() {
			if err != nil {
				h.host.Toast("Failed to restore the song: " + err.Error())
				return
			}
			h.refreshFolder(h.route.Query)
		})
	}()
}

// fillFolderTracks completes the songs playable-tracks gives, which have no
// length, format or album, from the folders already loaded: a song with no
// length showed 0:00 and played with no end the progress bar could reach.
func (h *Home) fillFolderTracks(tracks []netease.Track) {
	f := h.folderState()
	if f == nil || len(f.pages) == 0 {
		return
	}
	known := map[string]netease.Track{}
	for _, page := range f.pages {
		for _, item := range page.Tracks {
			known[item.Track.Ref] = item.Track
		}
	}
	for i, tr := range tracks {
		full, ok := known[tr.Ref]
		if !ok {
			continue
		}
		if tr.Duration == 0 {
			tracks[i].Duration = full.Duration
		}
		if tr.Format == "" {
			tracks[i].Format, tracks[i].Cue = full.Format, full.Cue
		}
		if tr.Album.Name == "" {
			tracks[i].Album = full.Album
		}
	}
}
