import { Link, createFileRoute } from '@tanstack/solid-router'
import { For, createSignal, type JSX } from 'solid-js'
import { Layout } from '../components/Layout'
import { AppWindow } from '../components/AppWindow'
import { GitHubMark, Icon, type IconName } from '../components/Icon'
import { ThemeSwitch } from '../theme'
import { mygo, repo } from '../site'

export const Route = createFileRoute('/')({
  component: Home,
})

function Home() {
  return (
    <Layout>
      <Hero />
      <Features />
    </Layout>
  )
}

function Hero() {
  return (
    <section class="hero">
      <div class="grooves" aria-hidden="true" />
      <div class="hero-text">
        <p class="eyebrow">
          <span class="pill">Open Source</span>A third-party desktop music client
        </p>
        <h1 class="display hero-title">
          <span>Listen natively,</span>
          <span>listen quietly.</span>
        </h1>
        <p class="lede">
          Written in Go and <a href={mygo}>MyGo</a> with a native UI — no WebView. Quick to launch, light on
          resources, and stays out of your way when idle.
        </p>
        <div class="actions">
          <Link class="btn btn-primary" to="/download">
            <Icon name="download" size={17} />
            Download
          </Link>
          <a class="btn btn-ghost" href={repo}>
            <GitHubMark size={17} />
            View source
          </a>
        </div>
        <p class="fine">Free · open source under AGPL-3.0 · sign in by scanning a QR code</p>
      </div>
      <div class="hero-window">
        <AppWindow />
        <p class="hint">Try pressing play, or liking the song.</p>
      </div>
    </section>
  )
}

type Feature = { icon: IconName; title: string; body: string; wide?: boolean; extra?: () => JSX.Element }

const features: Feature[] = [
  {
    icon: 'wave',
    title: 'Stream while downloading, up to Hi-Res',
    body: 'No waiting for the whole file. Standard, higher, very high, lossless (FLAC) and Hi-Res — pick what your ears and bandwidth deserve.',
    wide: true,
    extra: () => <Quality />,
  },
  {
    icon: 'disk',
    title: 'Heard it? It stays',
    body: 'Songs you play are cached locally and replay instantly.',
  },
  {
    icon: 'history',
    title: 'Pick up where you left off',
    body: 'Reopen after quitting and resume from the previous queue and position.',
  },
  {
    icon: 'volume',
    title: 'Gentle volume',
    body: 'Volume leveling, with a soft fade-out and fade-in on pause and resume.',
    extra: () => <Fade />,
  },
  {
    icon: 'radio',
    title: 'Private FM and Heart Mode',
    body: 'Three moods — discovery, familiar and exploring — or start Heart Mode from the songs you love.',
  },
  {
    icon: 'calendar',
    title: 'Daily Recommendations',
    body: 'Refreshed at 6:00 every day, plus recommended playlists, charts, your playlists and search.',
  },
  {
    icon: 'qr',
    title: 'Scan QR to Sign In',
    body: 'Scan with the NetEase Cloud Music app. Your session lives in the system credential store.',
  },
  {
    icon: 'notch',
    title: 'Notch Player',
    body: 'A mini player in the screen’s top notch that expands when you hover over it. Enabled in settings.',
  },
  {
    icon: 'lyrics',
    title: 'Word-by-word lyrics & Groove',
    body: 'Synced per-word highlighting with translations side by side; in Groove, each word lifts, glows and settles as it’s sung. Song comments too.',
    wide: true,
  },
  {
    icon: 'sun',
    title: 'Light, dark, or follow the system',
    body: 'Just like this site — try it.',
    extra: () => <ThemeSwitch labels class="feature-theme" />,
  },
]

function Features() {
  return (
    <section id="features" class="section">
      <div class="section-head">
        <p class="eyebrow">Features</p>
        <h2 class="display">
          <span>Everything you need,</span>
          <span>built in.</span>
        </h2>
        <p class="lede">Rebuilt for the desktop from discovery to playback. Likes, the queue, shuffle and repeat — all here.</p>
      </div>
      <div class="features">
        <For each={features}>
          {(f) => (
            <article class="feature" classList={{ wide: f.wide }}>
              <div class="feature-icon">
                <Icon name={f.icon} size={20} />
              </div>
              <h3>{f.title}</h3>
              <p>{f.body}</p>
              {f.extra?.()}
            </article>
          )}
        </For>
      </div>
    </section>
  )
}

const qualities = [
  { name: 'Standard', spec: '128 kbps · MP3' },
  { name: 'Higher', spec: '192 kbps · MP3' },
  { name: 'Very High', spec: '320 kbps · MP3' },
  { name: 'Lossless', spec: 'up to 48 kHz / 16 bit · FLAC' },
  { name: 'Hi-Res', spec: 'up to 192 kHz / 24 bit · FLAC' },
]

// The settings' choice of quality, as the app shows it.
function Quality() {
  const [q, setQ] = createSignal(3)
  return (
    <div class="quality">
      <div class="segmented" role="radiogroup" aria-label="Audio Quality">
        <For each={qualities}>
          {(item, i) => (
            <button
              type="button"
              role="radio"
              aria-checked={q() === i()}
              classList={{ on: q() === i() }}
              onClick={() => setQ(i())}
            >
              <span>{item.name}</span>
            </button>
          )}
        </For>
      </div>
      <p class="quality-spec">{qualities[q()].spec}</p>
    </div>
  )
}

// A wave fading out at a pause and back in, drawn.
function Fade() {
  const bars = Array.from({ length: 44 }, (_, i) => {
    const env = i < 16 ? 1 : i < 22 ? 1 - (i - 16) / 6 : i < 26 ? 0 : Math.min((i - 26) / 6, 1)
    return Math.max(0.06, env * (0.45 + 0.55 * Math.abs(Math.sin(i * 1.7) * Math.cos(i * 0.6))))
  })
  return (
    <div class="fade" aria-hidden="true">
      <For each={bars}>{(h) => <i style={{ height: `${h * 100}%` }} />}</For>
      <span class="fade-mark">Pause</span>
    </div>
  )
}
