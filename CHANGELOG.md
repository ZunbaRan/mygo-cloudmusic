# Changelog

## 0.2.15

- Fixed false "connection failed" popups from self-hosted platforms (FN Music / Subsonic / Songloft) at startup:
  - Added cold-start retry with exponential backoff (3 automatic retries at 1.5s and 3s intervals), keeping the normal connect-loading placeholder during retries so a not-yet-awake OS network stack no longer triggers a popup
  - Relaxed the FN candidate-IP and DDNS probe timeout to 5 seconds, greatly improving connection tolerance on home networks and after wake-from-sleep
- Fixed a bug in MPV mode where a song that took a while to load could skip unexpectedly:
  - Parse the libmpv `EventEndFile` reason (`EndFileReason`) precisely, strictly distinguishing "played to the end" from "load/network error (ERROR)", so a network fault is never mistaken for the end of a song
  - MPV now defaults to `network-timeout=30` plus low-level stream reconnects (`reconnect=1,reconnect_streamed=1`) for better resistance to flaky networks
  - Enabled HTTP Keep-Alive pooling on the local proxy client, avoiding dropped requests caused by repeated handshakes during multi-part Range probing of large files
- Improved MPV instance reuse and automatic next-track prefetch (Playlist Prefetch):
  - The persistent MPV engine instance is reused when switching tracks instead of destroying and rebuilding the C engine, removing hardware-handle reinit latency
  - After playback starts, a background goroutine resolves the next track and appends it to MPV's internal playlist (`loadfile append`), truly activating `prefetch-playlist="yes"` and the 64 MB demuxer cache for instant, gapless track changes

## 0.2.14

- Added a professional 10-band graphic equalizer:
  - ISO center frequencies (31Hz, 62Hz, 125Hz, 250Hz, 500Hz, 1kHz, 2kHz, 4kHz, 8kHz, 16kHz), ±12 dB gain adjustment and a preamp control
  - Classic tuning presets: Flat, Vocal Pop, Bass Boost, Treble Boost, Rock, Classical, EDM and a custom preset
  - Quick entry in the player bar: with the advanced audio decoder enabled, an EQ mixer panel sits on the right of the bottom bar — the icon highlights while the panel is open
  - Built on hot-reloadable MPV audio filters, so tuning never interrupts the playing stream
- Added audio output device enumeration and independent hardware selection:
  - Enumerates physical output devices (Windows WASAPI, macOS CoreAudio, Linux ALSA/Pulse/PipeWire, and external USB DACs/amps)
  - A device dropdown in "Advanced Audio & Hardware Decoding" routes audio to a specific sound card or DAC and remembers the choice
  - Combined with "Exclusive Mode", an external USB DAC can take over the hardware channel for bit-perfect output
- Wireless casting (DLNA) hi-res cover relay to smart TVs/screens:
  - The local streaming proxy gained a cover-relay endpoint that serves the current track's album art to LAN playback devices in real time
  - Completed UPnP DIDL-Lite metadata injection, so smart TVs, big screens and screened speakers show hi-res covers and track info while casting
- Low-level stream buffering and identity improvements:
  - MPV now reports a system client name (`MyGo Cloud Music`) consistent with the OS mixer
  - Tuned demuxer cache (`demuxer-max-bytes` 64 MiB, `demuxer-max-back-bytes` 32 MiB) and smart prefetch for smoother Hi-Res streams and high-bitrate DSD

## 0.2.13

- Fixed noticeably low volume with the advanced audio decoder (MPV):
  - Decoupled MPV audio output from the pure-Go linear player's squared perceptual loudness curve ($volume^2$); the UI volume slider now maps losslessly into MPV's native nonlinear volume space, so 100% volume is true 0 dB bit-perfect output
  - MPV gained native ReplayGain loudness normalization (Preamp/Fallback) with clipping protection (`replaygain-clip`) for full, clean sound at even loudness
  - Improved cooperation between the player fader and MPV on track jumps and pause/resume for smooth fades
- Wireless casting (DLNA) stability and discovery improvements (still being tuned):
  - Upgraded the `go-upnpcast` core and added a persistent device cache, so devices no longer flicker or vanish during background scans
  - Expanded dual-channel SSDP service-type probing (MediaRenderer 1/2, AVTransport 1/2 and RootDevice) with decoupled timeouts, improving speaker/TV discovery across LANs
  - Completed casting metadata (DIDL-Lite) delivery and the direct audio-proxy channel

## 0.2.12

- Experimental advanced audio decoder (based on the libmpv audio core):
  - An "Advanced Audio Decoder" toggle and component manager in settings downloads a minimal decoding core on demand, keeping the main app lean
  - Native decoding of all audio formats — no more lossy server-side transcoding (no MP3 conversion):
    - **DSD / SACD sources**: `.dsf` / `.dff` (DSD64 / DSD128 / DSD256, 1-bit 2.8 MHz/5.6 MHz, restored to 352.8 kHz floating-point PCM)
    - **DTS surround**: `.dts` (DTS-ES / DTS 5.1/6.1 multichannel decoding and high-quality stereo downmix)
    - **Lossless masters**: `.flac`, `.ape` (Monkey's Audio), `.alac` (Apple Lossless)
    - **Uncompressed hi-rate audio**: `.wav`, `.aiff` (up to 24bit/32bit, 96 kHz/192 kHz masters)
    - **Mainstream formats**: `.m4a` / `.aac`, `.wma`, `.ogg` / `.opus` / `.vorbis`, `.mp3`
    - **CUE sheet splitting**: precise native-timeline playback of each track in single-file images (WAV/FLAC/APE + CUE)
- Global audio and playback UI changes:
  - The advanced decoder is now a global playback setting, decoupled from the NetEase settings — it applies on FN Music, Songloft, Navidrome / Subsonic and every other platform
  - A wireless-casting entry and device-list popover added to the top right of the title bar (still being tuned)
- Permanently fixed the "buffering forever" deadlock on large files and special formats from self-hosted libraries:
  - The local streaming proxy now preempts Range seeks: for audio whose metadata sits at the end of the file (DSF tail ID3v2 tags, non-fast-start M4A tail moov boxes, APE tail APEv2 tags), seeking interrupts and takes over the previous connection, eliminating NAS-side stream queue lockups
  - Player core lifecycle rework: with the advanced decoder on, playback resolves and starts in milliseconds — no more infinite buffering caused by the pure-Go cache force-downloading large files and gomp3 mis-scanning them
  - Fixed a slow progress bar and position mismatch during MPV playback:
    - Fixed the player ready-state check, restoring smooth per-second UI clock updates and live lyric sync
    - Prefer the track's authoritative duration, avoiding progress-bar stretching caused by decoder estimates on network streams
    - More immediate position feedback and higher-precision time conversion when scrubbing
- Windows installer over-install experience:
  - Running-process detection and a friendly close prompt in NSIS — overwriting while the client is running no longer requires manually killing it first

## 0.2.11

- Added Songloft self-hosted music server support: switch to the Songloft platform in settings, enter the server address, username and password to connect; credentials persist in the system secure store
- Library and multi-dimensional sorting: browse the full library sorted by date added, title, artist, album and more, with quick play and shuffle
- Aggregated album and artist browsing: album and artist lists lead directly to their track detail pages
- Hierarchical folder browsing: deep adaptation for Songloft (v2.12.1+) folder views with large folder-cover tiles, breadcrumb navigation, play-all / shuffle for a whole folder, and parent-folder jumps
- Playlist and track management: create, rename and delete playlists (fixed create-playlist type validation); the context menu supports copy title/artist, add to queue, remove from playlist and add to other playlists (excluding the current one)
- Lyrics parsing: the Songloft and Subsonic adapters now unescape HTML/XML character entities, fixing garbled symbols in some embedded lyrics
- Per-platform recent-play persistence: Navidrome, Jellyfin, Gonic, Subsonic and Songloft each get their own local history file, ending cross-platform history pollution

## 0.2.10

- Experimental support for Navidrome, Jellyfin, Gonic and other Subsonic-compatible platforms (fully tested on Navidrome so far)
- Connect with server address, username and password/API token; credentials and sessions persist per platform in the system credential store
- Library and folder browsing: descend the folder tree and play recursively, or scroll the full song list
- Albums and artists: sort albums by newest, most played, recently played, year, name, artist name and favorites; artist detail (top songs and albums) and album detail (full track list)
- Favorite sync: heart/unheart tracks and save/unsave albums with real-time server Star/Unstar sync and instant count refresh
- Playlist management: browse personal playlists, create, rename and delete; add songs to or remove them from playlists via the song menu
- Search: multi-type search across tracks, albums and artists; results open detail pages for smooth browsing
- Playback and transcoding: native playback or server-transcoded streams, with configurable transcode mode and max bitrate (128k / 192k / 256k / 320k)
- Polish: fixed jank on heavy default-cover loads, improved the cover pipeline and hi-res singleton thumbnails, fixed count flicker on most/recent album sorting, and friendly handling of malformed server XML errors and mount failures
- Platform order unified: NetEase Cloud Music, QQ Music, FN Music, Navidrome, Jellyfin, Gonic, Subsonic-compatible

## 0.2.9

- Added the QQ Music platform: switch in settings, sign in by scanning a QR code (QQ / WeChat), or browse and preview without signing in; accounts are independent of NetEase and FN Music
- QQ Music home: greeting and date, a big daily-recommendation card (cover-tinted gradient, blurred cover, spinning vinyl stack, hover to pull the record), "Everyone's Listening" and other recommendation shelves (reshuffle, arrow-key and trackpad paging), Today's Picks and featured playlists
- Daily recommendations page (date cover card, play all / shuffle), charts (all 29 QQ charts, official charts preview the top three songs), purchased music (albums / tracks, load more)
- Search: tracks / artists / albums tabs, suggestions and recent searches (saved per platform) while typing, 20 results per page with scroll loading and a live total; playing a result continues into following pages
- Artist page: round avatar, alias / bio, top songs / latest songs / albums tabs; albums can be saved to "My Likes"; playlists support create, rename, delete and unsave; the song menu supports copy link, view artist and view album
- My Likes: hearted songs and saved albums tabs, with paged album loading; the profile page shows avatar, nickname, membership badge, signature and created playlists
- Private FM: a big cover-tinted card with play / pause / next and favorite; automatic continuous playback with live buffering and error states
- Song comments: top and latest comments load by page with like counts; interlude hot comments can be enabled in settings (QQ doesn't support liking comments)
- Playback: standard (128k) / very high (320k) / lossless (FLAC) quality; volume leveling, interlude hot comments and search history live in QQ's own settings namespace, no longer shared with NetEase; word-by-word lyrics with translations and the tinted big-card view still work
- The UI uses QQ Music's brand green; hearts stay red; the bottom player, Now Playing, lyrics and cache remain shared across platforms
- README now covers FN Music and QQ Music, and mygo-qyin (the QQ adaptation reference) joins the acknowledgements

## 0.2.8

- More stable streaming: higher concurrent HLS segment prefetching for less waiting and fewer stalls
- Completed transcode-stream retry and on-demand renewal for expired segments, filtering empty audio segments to avoid interruptions or early track changes
- Tuned long-connection timeouts and fixed early termination caused by non-audio tags at the end of FLAC files

## 0.2.7

- FN Music gained sorting for favorites, library, albums, album detail, artists, artist songs, genres, genre songs and folders; ascending and descending with web-consistent defaults, persisted across restarts
- Added an artist album list, and pinned the artist songs/albums switcher and sort buttons so they no longer jump when switching views
- Fixed unresponsive unsave on the favorites page, newly-favorited songs not appearing, and ineffective refresh; the list now updates immediately after unsaving
- Playing an FN song now writes a recent-play record at once — new plays show up without restarting
- Cloud-drive mount folders now show the name FN set for the path instead of a random mount name
- Album cards without art show a note placeholder instead of a blank square

## 0.2.6

- Smoother loading on FN Music pages: library, playlists, albums, artists, search, roam, folders and genres now use skeleton screens matching the final layout, reserving space for covers, titles, action buttons and list structure to reduce flicker and blank waits
- Better first-paint placeholders, paged-loading placeholders and large-cover transitions on folder and genre views, so content settles in naturally

## 0.2.5

Continued FN Music adaptation. The folder view and genre view are marked experimental for now; other platforms are unchanged.

- **Experimental folder view**: browse the multiple library roots FN Music is configured with, descend folder by folder, with breadcrumbs, sorting, refresh and lazy loading — it never scans the whole library on entry
- Folder pages support play folder, shuffle, add to playlist, copy path, move to trash and restore folders or songs; folder and song writes use FN's generation check so nothing lands on a stale page after switching accounts or platforms
- **Experimental genre view**: browse songs by the genres FN Music scanned, with a genre list, genre detail, song counts, paging, genre-name/song-count sorting and song sorting
- Genre song pages support single-song play, play all, multi-select, select current list, select the whole genre, append to queue, add to playlist and favorite; play-all and select-all read the full song set only on demand
- Genres without covers use stable generative artwork instead of a blank cover; song lists keep the shared cover cache and skeleton loading
- FN song lists and folder song lists uniformly show album, favorite, duration, file size and format; the player can now append to the queue without replacing what's playing
- When the FN server lacks an optional capability or an interface errors, the related page hides or shows a retryable state without affecting the connected library or the player

## 0.2.4

This is the fork release merging FN Music support with desktop lyrics. After updating, note:

- **The config format is incompatible with older versions**. There's no automatic migration; if NetEase or FN Music sign-in, platform selection, playback or UI settings misbehave after upgrading, reconfigure them in settings. If needed, quit the app and delete the old app config before signing in again.
- Added the FN Music platform. Switch between NetEase Cloud Music and FN Music in settings; FN Music's connection, sign-in, playlists, albums, artists, search, lyrics, favorites, queue and playback settings are handled through its own API, while NetEase keeps the upstream defaults.
- FN Music supports server address, account sign-in and disconnect; the settings page shows available options per platform and hides NetEase-only options (quality, interlude, hot comments) on FN.
- FN Music's playlist, album, artist and search lists support paging, playback, shuffle, favorites and playlist operations; the queue and platform state are isolated so switching platforms never carries over another platform's queue or playback.
- A standalone desktop lyrics window: open it separately, lock and unlock, original and translated lyrics, color and display options, always-on-top, and recovery to the current screen after a display disconnects.
- Refined the transparent window, bottom-center positioning, action-bar hiding and locked-state interaction; the window position is corrected after display changes so it never lands on an invisible screen.
- Updated MyGo to v0.7.4 and kept the Windows custom-install-directory fix so reinstalling respects the chosen directory.
- This fork's app updates and download page use the `ZunbaRan/mygo-cloudmusic` release URL; the sherpa-onnx speech components still come from upstream.

## 0.2.3

- Going back to a page you just saw, a home card, or the notch shows covers instantly — no reload, no matter how long you were away
- Cover caching uses at most 64 MB of memory, dropping the least recently seen first

## 0.2.2

- Search gained a Playlists tab: found playlists show as cover cards with track count and creator, and open their page on click
- Returning to Now Playing after a while elsewhere no longer reloads and fades the cover back in

## 0.2.1

- The Now Playing cover keeps a gap from the tabs above; cover, title, progress bar and tabs are left-aligned
- Now Playing tabs sit at the same spot on every page — switching lyrics, groove or pick no longer shifts them, and changing comment counts don't move the others
- On large or maximized windows, the Now Playing lyrics page centers

## 0.2.0

- Pick gained spacing between covers and the tabs above, and the cover wall's top blur fade is shorter
- The notch greeting holds still while the pointer rests on it, and its "Enable" is clickable

## 0.1.9

- Now Playing gained "Pick": upcoming songs spread out as a cover wall you can drag, fling or scroll with a trackpad; each row sways with the music, and clicking a cover plays that song
- ⌘1 through ⌘5 switch Now Playing tabs in order (lyrics, groove, up next, pick, comments); podcast episodes switch by tab order too
- Each playlist song now shows its album cover beside it
- Covers no longer all reload and flicker when returning to a window after a while

## 0.1.8

- Security fix: the song cache only writes inside the cache folder — a malformed quality name from the API can no longer write elsewhere
- A malformed song size from the API no longer pre-allocates excessive memory

## 0.1.7

- Groove gained two lyric effects, "Partita" and "Confide", switchable next to the tab: Partita splits a line into staggered rows that spring in one after another; Confide breaks a line into big lines whose words grow as they're sung
- Groove effects gained "Sparks": each word throws sparks as it's sung, on top of any effect
- Dragging the window by the title bar no longer makes it shake

## 0.1.6

- Podcast captions: downloading the speech model after an episode finished re-transcribes it from the start instead of reporting no speech found
- Switching episodes or starting playback shows "Recognizing…" instead of the previous episode's captions, "no speech detected" or "caption generation failed"
- "Previous" now plays the previous song directly instead of returning to the start of the current one
- Loading the next song no longer makes the Now Playing view jump as the quality badge disappears

## 0.1.5

- Podcast captions: after downloading the speech model in settings, episodes transcribe locally as they play, scrolling in the "Captions" tab like lyrics — tap a line to jump to it; audio is never uploaded
- Sidebar gained "Podcasts": subscribed podcasts, NetEase recommendations and top podcasts — click a card to open a podcast, or the cover button to play the latest episode
- Like programs, view program comments and like comments (#23)
- "My Profile" shows subscribed podcasts and followed artists

## 0.1.4

- Reopening the app after an auto-update no longer leaves an extra Dock icon: the new version replaces the old one only after it exits, keeping the pinned Dock icon
- The artist page "Follow" button works again instead of reporting failure

## 0.1.3

- Search albums, artists, users and podcasts, and open their pages (#9)
- Clicking an artist's name opens the artist page (#21)
- A tray icon joins the menu bar and notification area — a record and tonearm (#19)
- The Dock icon shows while a window is open, so Command-Tab can switch to the main or settings window
- The account menu gained "My Profile"
- A program's host and podcast link to the podcast page, replacing the "View podcast" button
- Opening an artist, album or podcast from Now Playing now closes Now Playing
- After NetEase refreshes recommendations at 6:00 (Beijing time), the home page reloads so Daily Recommendations no longer stays a day behind
- Members' VIP songs play normally instead of being mislabeled as delisted (#11)
- The new-playlist button became a rounded square at the same height as the "My Playlists" heading
- With accent color in the Windows 11 title bar, the main window no longer shows a color strip on top
- The minimum supported system is now macOS 13

## 0.1.2

- The notch player and its settings are macOS-only: on Windows it would hang as a black bar at the top of the screen
