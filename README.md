# MyGo Music Client

A third-party desktop music client written in Go and [MyGo](https://github.com/egoist/mygo), with a native UI and no WebView.

Focused on self-hosted and private cloud music platforms: **FN Music**, **Songloft**, and **Subsonic-compatible servers such as Navidrome / Jellyfin / Gonic**. Each platform keeps its own account, playlists, favorites and playback settings, while the bottom player bar, Now Playing page, lyrics and cache share the same high-quality experience.

## Features

### FN Music

- Connect to an FN Music service on a NAS: enter the address, username and password to sign in; NAS devices with an access code also need the code
- Sign-in state and connection settings are stored in a per-platform namespace and the system secure credential store
- Browse the music library by folder, with cover-wall and list views, plus paged browsing by artist, album and genre
- Liked songs, recently played and play history, with the ability to delete history entries
- Create, rename, delete and save playlists; add songs to or remove them from playlists
- Search songs, artists, albums, podcasts and genres, with suggestions as you type
- Formats that can't be played locally are transcoded to HLS by the server; transcoding mode (auto / always / never) and bitrate are adjustable in settings
- Per-page sort selection that remembers each list's last ordering
- Stream-while-downloading and in-playlist playback, with songs cached locally
- Mini turntable: roam the library with aimless shuffle, and save a song with one click when you like it

### Songloft self-hosted music server

- Connect to a self-hosted Songloft server: enter the server address, username and password to sign in; sign-in state is stored in the system secure credential store
- Browse the full library, shuffle, and switch between multiple sort orders (date added, title, artist, album, etc.)
- Deep folder-tree browsing and recursive playback, with large folder-cover tiles, breadcrumb navigation and parent-folder jumps
- Aggregated album and artist detail pages that lead straight to track lists
- Playlist management: create, rename and delete playlists, and add or remove tracks via the track context menu
- Live heart sync: favorite/unfavorite tracks ("Liked"), kept in two-way real-time sync with the server
- Full search: multi-type lookup across songs, albums and artists with direct jump to details
- Word-by-word lyrics and translations: embedded and sidecar lyrics are loaded and parsed automatically, with character-entity unescaping

### Subsonic-compatible platforms (Navidrome / Jellyfin / Gonic, etc.)

- Connect to self-hosted Subsonic / OpenSubsonic-compatible services (fully tested on Navidrome and other mainstream servers)
- Enter the server address, username and password / API token to connect; credentials are stored in a per-platform namespace and the system secure store
- Browse the full library or descend the folder tree and play recursively
- Album and artist details with sorting by newest, most played, recently played, year, alphabetical and favorites
- Heart a track or save/unsave an album, with real-time sync of the server's Star marks and instant refresh
- Playlist creation, renaming, deletion and track management (add/remove)
- Multi-type search (tracks, albums, artists) with results opening details and continuous playback
- Server-side transcoding output with a max-bitrate option (auto / always / never and bitrate cap in settings)
- Local recent-play records isolated per platform, never leaking across platforms

### General

- Experimental advanced audio decoder: native decoding of DSD (DSF/DFF), DTS, FLAC, APE, ALAC, WAV, M4A, WMA and CUE sheets — no more server-side transcoding
- Songs you've heard are cached locally and replay instantly
- Volume leveling, fade-out/in on pause and resume, glitch-free seeking
- Reopen after quitting and resume from the previous queue and position
- Word-by-word lyrics with translations and the "Groove" view on Now Playing
- Notch Player: a mini player living in the screen's top notch that expands on hover (enabled in settings)
- Liked songs, play queue, shuffle and repeat
- Light / dark / follow-system themes

## Build

Requires Go 1.27+.

```bash
go tool mygo dev     # run in dev mode
go tool mygo build   # package into build/
go test ./...        # run the tests
```

## Acknowledgements

- [MyGo](https://github.com/egoist/mygo): native UI and app framework
- [folia-major](https://github.com/chthollyphile/folia-major): reference for the lyric animations
- [sherpa-onnx](https://github.com/k2-fsa/sherpa-onnx), [Qwen3-ASR](https://github.com/QwenLM/Qwen3-ASR) and [Silero VAD](https://github.com/snakers4/silero-vad): on-device speech recognition
- [go-subsonic](https://github.com/supersonic-app/go-subsonic): reference implementation of the Subsonic API client

## License

[AGPL-3.0](LICENSE)
