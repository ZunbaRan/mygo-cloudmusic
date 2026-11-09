import { createFileRoute } from '@tanstack/solid-router'
import { For, Show, createSignal, onMount } from 'solid-js'
import { Layout } from '../components/Layout'
import { GitHubMark, Icon } from '../components/Icon'
import { release, repo, site } from '../site'

export const Route = createFileRoute('/download')({
  head: () => ({ meta: [{ title: `Download · ${site.name}` }] }),
  component: Download,
})

type OS = 'mac' | 'windows' | 'linux'

// A download of a platform: the asset of the latest release it is, found by
// its name, as `mygo build` names installers.
type Build = { label: string; kind: string; match: (name: string) => boolean; primary?: boolean }

const has = (...parts: string[]) => (name: string) => {
  const n = name.toLowerCase()
  return parts.every((p) => n.includes(p))
}

const platforms: { os: OS; name: string; note: string; builds: Build[] }[] = [
  {
    os: 'mac',
    name: 'macOS',
    note: 'macOS 13 or later',
    builds: [
      { label: 'Apple Silicon', kind: '.dmg', match: has('arm64', '.dmg'), primary: true },
      { label: 'Intel', kind: '.dmg', match: has('amd64', '.dmg') },
    ],
  },
  {
    os: 'windows',
    name: 'Windows',
    note: 'Installs for the current user, no admin rights needed',
    builds: [
      { label: 'x64', kind: 'Installer', match: has('setup', 'amd64', '.exe'), primary: true },
      { label: 'ARM64', kind: 'Installer', match: has('setup', 'arm64', '.exe') },
    ],
  },
  {
    os: 'linux',
    name: 'Linux',
    note: '.deb for Debian and Ubuntu; other distros use the archives or the command below',
    builds: [
      { label: 'x64', kind: '.deb', match: has('amd64', '.deb'), primary: true },
      { label: 'ARM64', kind: '.deb', match: has('arm64', '.deb') },
      { label: 'x64', kind: '.tar.gz', match: has('linux', 'amd64', '.tar.gz') },
      { label: 'ARM64', kind: '.tar.gz', match: has('linux', 'arm64', '.tar.gz') },
    ],
  },
]

const installSh = `curl -fsSL ${repo}/releases/latest/download/install.sh | sh`
const installPs = `irm ${repo}/releases/latest/download/install.ps1 | iex`

// A build's asset; a build the release has none of links to the release.
const asset = (b: Build) => release.assets.find((a) => b.match(a.name))

function detectOS(): OS | undefined {
  const ua = navigator.userAgent
  if (/iPhone|iPad|Android/.test(ua)) return undefined
  if (/Mac/.test(ua)) return 'mac'
  if (/Windows/.test(ua)) return 'windows'
  if (/Linux|X11/.test(ua)) return 'linux'
}

function megabytes(bytes: number) {
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

function day(iso: string) {
  const d = new Date(iso)
  return d.toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })
}

function Download() {
  // The page is rendered at build time; the system of the visitor comes
  // once it runs in the browser.
  const [os, setOS] = createSignal<OS>()
  onMount(() => setOS(detectOS()))

  return (
    <Layout>
      <section class="download">
        <div class="section-head download-head">
          <p class="eyebrow">Download</p>
          <h1 class="display">
            <span>Your music,</span>
            <span>on the desktop.</span>
          </h1>
          <p class="fine download-version">
            Version {release.tag_name.replace(/^v/, '')} · Released {day(release.published_at)}
          </p>
        </div>

        <div class="platforms">
          <For each={platforms}>
            {(p) => (
              <article class="platform">
                <div class="platform-head">
                  <h2>{p.name}</h2>
                  <Show when={os() === p.os}>
                    <span class="pill">Your system</span>
                  </Show>
                </div>
                <p class="platform-note">{p.note}</p>
                <div class="platform-builds">
                  <For each={p.builds}>
                    {(b) => (
                      <a
                        class="build"
                        classList={{ primary: b.primary }}
                        href={asset(b)?.browser_download_url ?? `${repo}/releases/tag/${release.tag_name}`}
                      >
                        <Icon name="download" size={16} />
                        <span class="build-label">{b.label}</span>
                        <span class="build-kind">{b.kind}</span>
                        <Show when={asset(b)}>{(a) => <span class="build-size">{megabytes(a().size)}</span>}</Show>
                      </a>
                    )}
                  </For>
                </div>
              </article>
            )}
          </For>
        </div>

        <div class="install">
          <h2>Install with one command</h2>
          <p>Downloads the latest version into your user directory and keeps itself updated.</p>
          <Command label="macOS & Linux" text={installSh} />
          <Command label="Windows PowerShell" text={installPs} />
        </div>

        <p class="fine download-more">
          <GitHubMark size={14} />
          <a href={`${repo}/releases`}>All releases and release notes</a>
        </p>
      </section>
    </Layout>
  )
}

function Command(props: { label: string; text: string }) {
  const [copied, setCopied] = createSignal(false)
  let timer: ReturnType<typeof setTimeout> | undefined
  const copy = () => {
    navigator.clipboard?.writeText(props.text).then(() => {
      setCopied(true)
      clearTimeout(timer)
      timer = setTimeout(() => setCopied(false), 1600)
    })
  }
  return (
    <div class="command">
      <span class="command-label">{props.label}</span>
      <div class="command-line">
        <code>{props.text}</code>
        <button type="button" class="command-copy" onClick={copy} aria-label={`Copy the ${props.label} install command`}>
          {copied() ? 'Copied' : 'Copy'}
        </button>
      </div>
    </div>
  )
}
