import { mount } from 'svelte'
import '@fontsource-variable/inter'
import './lib/tokens.css'
import App from './App.svelte'

async function start() {
  if (import.meta.env.DEV && !(window as any).go) await import('./dev-mock')
  mount(App, { target: document.getElementById('app')! })
  if (import.meta.env.DEV) await seed()
}
start()

// Dev-only: preload a UI state with ?state=list|running|done|error|drag|empty-drag
async function seed() {
  const state = new URLSearchParams(location.search).get('state')
  if (!state) return
  const stores = await import('./lib/stores.svelte')
  const { app, settings } = stores
  if (state === 'empty-drag') { app.dragging = true; return }
  stores.addFiles(await (window as any).go.main.App.PickFiles())
  settings.outDir = '/home/chris/Pictures/converted'
  if (state === 'drag') app.dragging = true
  if (state === 'running' || state === 'done') {
    app.rows.forEach((r, i) => {
      if (state === 'running' && i > 2) return void (r.status = i === 3 ? 'working' : 'queued')
      Object.assign(r, i === 3 && state === 'done'
        ? { status: 'error', error: 'formato de imagen no reconocido' }
        : { status: 'done', outName: r.name.replace(/\.[^.]+$/, '') + '.webp', outSize: Math.round(r.size * (i === 2 ? 1.4 : 0.37)) })
    })
    app.running = state === 'running'
  }
}
