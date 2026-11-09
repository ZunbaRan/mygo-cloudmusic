import { For, createSignal, onMount } from 'solid-js'
import { Icon, Transport, type IconName } from './Icon'
import { Cover } from './Cover'
import { Like } from './Like'
import { clock, playing, position, seek, toggle, track } from '../player'

const weekdays = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']

// The greeting over the home page, as the app words it by the hour.
function greeting(h: number) {
  if (h < 5) return 'Up late?'
  if (h < 11) return 'Good morning'
  if (h < 13) return 'Good afternoon'
  if (h < 18) return 'Good afternoon'
  if (h < 23) return 'Good evening'
  return 'Up late?'
}

const nav: { icon: IconName; label: string }[] = [
  { icon: 'home', label: 'For You' },
  { icon: 'calendar', label: 'Daily Recommendations' },
  { icon: 'radio', label: 'Private FM' },
  { icon: 'chart', label: 'Charts' },
]

const playlists = ['Liked Songs', 'BanG Dream! Picks', 'Anime OP Mix', 'Late-Night Anime EDs']

const picks = [
  { name: 'MyGO!!!!! Complete', hue: 230 },
  { name: 'Hype Anime OPs', hue: 20 },
  { name: 'Healing Anime EDs', hue: 160 },
  { name: 'Ave Mujica Picks', hue: 300 },
  { name: 'Classic Anime Themes', hue: 45 },
]

const charts = [
  { name: 'Rising', hue: 18, songs: ['Haruhikage', 'Idol', 'Ao no Sumika'] },
  { name: 'New Songs', hue: 205, songs: ['Mayoiuta', 'Hekitenbansou', 'Hitoshizuku'] },
  { name: 'ACG', hue: 150, songs: ['Gurenge', 'unravel', 'A Cruel Angel’s Thesis'] },
]

// The app's main window, built in HTML: the sidebar, the home page with
// its daily card and charts, and the player bar. It plays the page's song.
export function AppWindow() {
  // Rendered at build time, the date is fixed; the visitor's own fills in.
  const [now, setNow] = createSignal(new Date(2026, 9, 26, 23, 30))
  onMount(() => setNow(new Date()))
  const date = () =>
    `${now().toLocaleString('en-US', { month: 'short' })} ${now().getDate()}, ${weekdays[now().getDay()]}`

  return (
    <div class="window" classList={{ playing: playing() }}>
      <aside class="win-side">
        <div class="lights" aria-hidden="true">
          <i />
          <i />
          <i />
        </div>
        <div class="win-search">
          <Icon name="search" size={13} />
          Search songs, artists
        </div>
        <For each={nav}>
          {(n, i) => (
            <div class="win-nav" classList={{ on: i() === 0 }}>
              <Icon name={n.icon} size={16} />
              {n.label}
            </div>
          )}
        </For>
        <div class="win-heading">My Playlists</div>
        <For each={playlists}>
          {(p, i) => (
            <div class="win-list">
              <Cover hue={(i() * 70 + 340) % 360} size={22} radius={5} />
              {p}
            </div>
          )}
        </For>
        <div class="win-side-foot">
          <div class="win-now">
            <Cover hue={18} size={36} radius={8} />
            <div>
              <b>{track.title}</b>
              <span>{track.artist}</span>
            </div>
          </div>
        </div>
      </aside>

      <section class="win-main">
        <div class="win-page">
          <div class="win-date">{date()}</div>
          <div class="win-greet">{greeting(now().getHours())}</div>

          <div class="daily">
            <div class="daily-day" aria-hidden="true">
              {now().getDate()}
            </div>
            <div class="daily-text">
              <div class="daily-note">{date()} · Updated daily at 6:00</div>
              <div class="daily-title">Daily Recommendations</div>
              <div class="daily-artists">MyGO!!!!!, YOASOBI, LiSA, Ave Mujica and more</div>
              <button type="button" class="daily-play" onClick={toggle}>
                <span class="bars" aria-hidden="true">
                  <i />
                  <i />
                  <i />
                </span>
                {playing() ? 'Now Playing' : 'Play'}
              </button>
            </div>
            <div class="daily-stack" aria-hidden="true">
              <Cover hue={205} size={112} radius={14} class="sleeve s3" />
              <Cover hue={150} size={124} radius={14} class="sleeve s2" />
              <div class="daily-record">
                <div class="vinyl" />
              </div>
              <Cover hue={18} size={136} radius={14} class="sleeve s1" />
            </div>
          </div>

          <div class="win-section">Official Charts</div>
          <div class="win-charts">
            <For each={charts}>
              {(c) => (
                <div class="chart">
                  <Cover hue={c.hue} size={64} radius={10} label={c.name} />
                  <ol>
                    <For each={c.songs}>{(s) => <li>{s}</li>}</For>
                  </ol>
                </div>
              )}
            </For>
          </div>

          <div class="win-section">Recommended Playlists</div>
          <div class="win-picks">
            <For each={picks}>
              {(p) => (
                <div class="pick">
                  <Cover hue={p.hue} size={120} radius={12} class="pick-cover" />
                  <span>{p.name}</span>
                </div>
              )}
            </For>
          </div>
        </div>

        <div class="win-bar">
          <div class="bar-song">
            <Cover hue={18} size={44} radius={8} />
            <div>
              <b>{track.title}</b>
              <span>{track.artist}</span>
            </div>
            <Like />
          </div>
          <div class="bar-center">
            <div class="bar-controls">
              <span class="ctl" aria-hidden="true">
                <Icon name="shuffle" size={14} />
              </span>
              <span class="ctl" aria-hidden="true">
                <Transport name="prev" size={16} />
              </span>
              <button type="button" class="ctl play" aria-label={playing() ? 'Pause' : 'Play'} onClick={toggle}>
                <Transport name={playing() ? 'pause' : 'play'} size={20} />
              </button>
              <span class="ctl" aria-hidden="true">
                <Transport name="next" size={16} />
              </span>
              <span class="ctl" aria-hidden="true">
                <Icon name="repeat" size={14} />
              </span>
            </div>
            <div class="bar-progress">
              <span>{clock(position())}</span>
              <Progress />
              <span>{clock(track.duration)}</span>
            </div>
          </div>
          <div class="bar-end" aria-hidden="true">
            <Icon name="lyrics" size={16} />
            <Icon name="queue" size={16} />
            <Icon name="volume" size={16} />
          </div>
        </div>
      </section>
    </div>
  )
}

// The progress bar: click or drag to seek.
function Progress() {
  let el!: HTMLDivElement
  const at = (e: PointerEvent) => {
    const r = el.getBoundingClientRect()
    seek((e.clientX - r.left) / r.width)
  }
  return (
    <div
      ref={el}
      class="progress"
      role="slider"
      aria-label="Progress"
      aria-valuemin={0}
      aria-valuemax={Math.round(track.duration)}
      aria-valuenow={Math.round(position())}
      aria-valuetext={clock(position())}
      tabindex="0"
      onPointerDown={(e) => {
        el.setPointerCapture(e.pointerId)
        at(e)
      }}
      onPointerMove={(e) => el.hasPointerCapture(e.pointerId) && at(e)}
      onKeyDown={(e) => {
        const step = e.key === 'ArrowRight' ? 5 : e.key === 'ArrowLeft' ? -5 : 0
        if (step) {
          e.preventDefault()
          seek((position() + step) / track.duration)
        }
      }}
    >
      <div class="progress-fill" style={{ width: `${(position() / track.duration) * 100}%` }} />
    </div>
  )
}
