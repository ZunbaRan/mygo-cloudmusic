export const repo = 'https://github.com/ZunbaRan/mygo-cloudmusic'

export type Asset = { name: string; size: number; browser_download_url: string }
export type Release = { tag_name: string; published_at: string; assets: Asset[] }

declare const __RELEASE__: Release
// The latest release, as it was when the site was built (see vite.config.ts).
export const release = __RELEASE__
export const mygo = 'https://mygo.egoist.dev'

export const site = {
  name: 'MyGo Cloud Music',
  title: 'MyGo Cloud Music — a native desktop music client',
  description:
    'A third-party desktop music client in Go and MyGo. Native UI, no WebView; stream-while-downloading, lossless and Hi-Res, word-by-word lyrics with Groove, and a Notch Player.',
}
