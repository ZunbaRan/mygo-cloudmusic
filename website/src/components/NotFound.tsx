import { Link } from '@tanstack/solid-router'
import { Layout } from './Layout'
import { Icon } from './Icon'

export function NotFound() {
  return (
    <Layout>
      <section class="lost">
        <div class="lost-record" aria-hidden="true">
          <div class="vinyl" />
        </div>
        <p class="eyebrow">404</p>
        <h1 class="display">This song is gone</h1>
        <p class="lede">Maybe it was delisted, maybe the link is wrong. Head home and try another.</p>
        <Link to="/" class="btn btn-primary">
          Back to home
          <Icon name="arrow" size={16} />
        </Link>
      </section>
    </Layout>
  )
}
