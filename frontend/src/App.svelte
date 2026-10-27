<script lang="ts">
  import { onMount } from 'svelte'
  import { EventsOn, OnFileDrop, OnFileDropOff, Quit, WindowIsMaximised, WindowMinimise, WindowToggleMaximise } from '../wailsjs/runtime/runtime'
  import { AddPaths, Convert, Cancel, Remove, SaveZip } from '../wailsjs/go/main/App'
  import DropZone from './lib/DropZone.svelte'
  import FileList from './lib/FileList.svelte'
  import Settings from './lib/Settings.svelte'
  import Planet from './lib/Planet.svelte'
  import Icon from './lib/Icon.svelte'
  import { app, settings, addFiles, pickFiles, formatBytes } from './lib/stores.svelte'

  interface Progress { id: string; status: 'done' | 'error'; outName?: string; outSize?: number; error?: string }
  interface Summary { ok: number; failed: number; cancelled: boolean }

  let maximised = $state(false)
  const syncMaximised = () => WindowIsMaximised().then((v) => (maximised = v))
  // Double-clicking the bar (but not its buttons) toggles maximise, like a native title bar.
  const onBarDblClick = (e: MouseEvent) => {
    if (!(e.target as HTMLElement).closest('button')) WindowToggleMaximise()
  }

  onMount(() => {
    syncMaximised()
    window.addEventListener('resize', syncMaximised)
    // Registers the DOM drag/drop listeners that stop WebKit from opening dropped images,
    // and receives the real file paths from the native side.
    OnFileDrop(async (_x, _y, paths) => {
      if (app.running || !paths?.length) return
      addFiles((await AddPaths(paths)) ?? [])
    }, true)

    // Wayland/Nautilus fallback: always cancel the browser's default (navigating to the
    // image) and read the file:// URIs straight from the drop when Wails' native path is missing.
    let depth = 0
    const hasFiles = (e: DragEvent) => !!e.dataTransfer && [...e.dataTransfer.types].some((t) => t === 'Files' || t === 'text/uri-list')
    const onEnter = (e: DragEvent) => {
      // Accept the drag as early as possible: a very fast drop can land before any dragover fires,
      // and WebKit then falls back to navigating to the dropped file.
      e.preventDefault()
      if (e.dataTransfer) e.dataTransfer.dropEffect = 'copy'
      if (hasFiles(e) && !app.running) { depth++; app.dragging = true }
    }
    const onLeave = (e: DragEvent) => { if (hasFiles(e) && --depth <= 0) { depth = 0; app.dragging = false } }
    const onOver = (e: DragEvent) => {
      e.preventDefault()
      if (e.dataTransfer) e.dataTransfer.dropEffect = 'copy'
    }
    const onDrop = async (e: DragEvent) => {
      e.preventDefault()
      depth = 0
      app.dragging = false
      if (app.running) return
      const paths = (e.dataTransfer?.getData('text/uri-list') ?? '')
        .split(/\r?\n/)
        .filter((l) => l.startsWith('file://'))
        .map((l) => decodeURIComponent(new URL(l).pathname))
      if (paths.length) addFiles((await AddPaths(paths)) ?? [])
    }
    window.addEventListener('dragenter', onEnter, true)
    window.addEventListener('dragleave', onLeave, true)
    window.addEventListener('dragover', onOver, true)
    window.addEventListener('drop', onDrop, true)

    const offs = [
      EventsOn('convert:progress', (p: Progress) => {
        const r = app.rows.find((r) => r.id === p.id)
        if (r) Object.assign(r, { status: p.status, outName: p.outName, outSize: p.outSize, error: p.error })
      }),
      EventsOn('convert:done', (s: Summary) => {
        app.running = false
        app.zipReady = settings.output === 'zip' && s.ok > 0
        app.message = s.cancelled ? 'Cancelado.' : ''
        for (const r of app.rows) if (r.status === 'working') r.status = 'queued'
      }),
    ]
    return () => {
      offs.forEach((off) => off())
      OnFileDropOff()
      window.removeEventListener('resize', syncMaximised)
      window.removeEventListener('dragenter', onEnter, true)
      window.removeEventListener('dragleave', onLeave, true)
      window.removeEventListener('dragover', onOver, true)
      window.removeEventListener('drop', onDrop, true)
    }
  })

  const reason = $derived(
    settings.output === 'folder' && !settings.outDir ? 'Elige una carpeta de destino para continuar.' : '',
  )
  const canConvert = $derived(app.rows.length > 0 && !app.running && !reason)

  const totalSize = $derived(app.rows.reduce((n, r) => n + r.size, 0))
  const finished = $derived(app.rows.filter((r) => r.status === 'done'))
  const failed = $derived(app.rows.filter((r) => r.status === 'error').length)
  const savedBytes = $derived(
    finished.reduce((n, r) => n + r.size, 0) - finished.reduce((n, r) => n + (r.outSize ?? 0), 0),
  )
  const progress = $derived(
    app.running && app.rows.length
      ? (app.rows.filter((r) => r.status === 'done' || r.status === 'error').length / app.rows.length) * 100
      : 0,
  )

  async function convert() {
    app.message = ''
    app.zipReady = false
    for (const r of app.rows) Object.assign(r, { status: 'working', error: undefined, outName: undefined })
    app.running = true
    try {
      await Convert({
        ids: app.rows.map((r) => r.id),
        options: {
          format: settings.format,
          quality: settings.quality,
          maxWidth: settings.maxWidth,
          maxHeight: settings.maxHeight,
        },
        formats: Object.fromEntries(app.rows.filter((r) => r.target).map((r) => [r.id, r.target])),
        output: settings.output,
        outDir: settings.outDir,
        suffix: settings.suffix,
      } as any)
    } catch (e) {
      app.running = false
      app.message = String(e)
      for (const r of app.rows) r.status = 'queued'
    }
  }

  async function saveZip() {
    try {
      const p = await SaveZip()
      if (p) app.message = `Guardado en ${p}`
    } catch (e) {
      app.message = String(e)
    }
  }

  function clear() {
    Remove(app.rows.map((r) => r.id)) // let the backend forget them, or re-adding is treated as a duplicate
    app.rows = []
    app.zipReady = false
    app.message = ''
  }
</script>

<div class="shell drop-target">
  <!-- svelte-ignore a11y_no_static_element_interactions (mouse-only shortcut; the buttons are the accessible path) -->
  <header ondblclick={onBarDblClick}>
    <div class="brand">
      <Planet mini size={26} />
      <span class="mark">Mercurio</span>
    </div>
    <span class="tag">Conversor de imágenes</span>
    <span class="spacer"></span>
    <span class="local"><Icon name="shield" size={14} /> Funciona en tu equipo</span>
    <div class="controls">
      <button aria-label="Minimizar" title="Minimizar" onclick={WindowMinimise}><Icon name="minimize" size={16} /></button>
      <button aria-label={maximised ? 'Restaurar' : 'Maximizar'} title={maximised ? 'Restaurar' : 'Maximizar'} onclick={WindowToggleMaximise}>
        <Icon name={maximised ? 'restore' : 'maximize'} size={16} />
      </button>
      <button class="close" aria-label="Cerrar" title="Cerrar" onclick={Quit}><Icon name="x" size={16} /></button>
    </div>
    <div class="progress" class:on={app.running}><i style="width:{progress}%"></i></div>
  </header>

  <main>
    {#if app.rows.length === 0}
      <DropZone />
    {:else}
      <div class="work">
        <section class="files">
          <div class="head">
            <div class="title">
              <h2>{app.rows.length} {app.rows.length === 1 ? 'imagen' : 'imágenes'}</h2>
              <span class="sub">
                {#if app.message}{app.message}
                {:else if finished.length && !app.running}
                  {finished.length} {finished.length === 1 ? 'convertida' : 'convertidas'} · {savedBytes >= 0 ? `ahorro de ${formatBytes(savedBytes)}` : `${formatBytes(-savedBytes)} más pesadas`}{failed ? ` · ${failed} ${failed === 1 ? 'falló' : 'fallaron'}` : ''}
                {:else}{formatBytes(totalSize)} en total{/if}
              </span>
            </div>
            <button class="pill tint small" onclick={pickFiles} disabled={app.running}><Icon name="plus" size={15} /> Añadir</button>
            <button class="pill tint small" onclick={clear} disabled={app.running}>Vaciar</button>
          </div>
          <div class="scroll"><FileList /></div>
        </section>

        <Settings {canConvert} {reason} count={app.rows.length}
          onconvert={convert} oncancel={() => Cancel()} onsavezip={saveZip} />
      </div>

      {#if app.dragging}
        <div class="overlay" aria-hidden="true">
          <Planet size={150} active />
          <h2>Suelta para añadir más imágenes</h2>
        </div>
      {/if}
    {/if}
  </main>
</div>

<style>
  .shell { height: 100%; display: grid; grid-template-rows: 64px minmax(0, 1fr); box-shadow: inset 0 0 0 1px var(--line-strong); }

  header {
    position: relative; display: flex; align-items: center; gap: 14px; padding: 0 16px 0 28px;
    border-bottom: 1px solid var(--line);
    --wails-draggable: drag; /* the bar moves the window; interactive children opt out below */
  }
  header button { --wails-draggable: no-drag; }
  .controls { display: flex; align-items: center; gap: 4px; margin-left: 6px; padding-left: 14px; border-left: 1px solid var(--line-strong); }
  .controls button {
    display: grid; place-items: center; width: 34px; height: 34px; border: 0; border-radius: 9999px;
    background: transparent; color: var(--fog); transition: background 0.15s, color 0.15s;
  }
  .controls button:hover { background: var(--tint); color: var(--on-dark); }
  .controls .close:hover { background: var(--primary); color: var(--on-primary); }
  .brand { display: flex; align-items: center; gap: 10px; }
  .mark { font-size: 20px; font-variation-settings: 'wght' 520; letter-spacing: 0; }
  .tag { color: var(--fog); font-size: 14px; padding-left: 14px; border-left: 1px solid var(--line-strong); }
  .spacer { flex: 1; }
  .local { display: inline-flex; align-items: center; gap: 6px; color: var(--fog); font-size: 13px; }
  .progress { position: absolute; left: 0; right: 0; bottom: -1px; height: 2px; opacity: 0; transition: opacity 0.2s; }
  .progress.on { opacity: 1; }
  .progress i { display: block; height: 100%; background: var(--primary); transition: width 0.3s; box-shadow: 0 0 12px var(--primary); }

  main { position: relative; min-height: 0; padding: 24px 28px 28px; }

  .work { height: 100%; display: grid; grid-template-columns: minmax(0, 1fr) 320px; gap: 20px; align-items: start; }
  .files {
    height: 100%; min-height: 0; display: grid; grid-template-rows: auto minmax(0, 1fr);
    background: var(--night-raised); border-radius: var(--r-md); overflow: hidden;
  }
  .head { display: flex; align-items: center; gap: 8px; padding: 16px 16px 14px 20px; border-bottom: 1px solid var(--line); }
  .title { flex: 1; display: grid; gap: 1px; min-width: 0; }
  h2 { font-size: 19px; line-height: 1.2; letter-spacing: 0; }
  .sub { font-size: 13px; color: var(--fog); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .scroll { overflow-y: auto; min-height: 0; }

  .overlay {
    position: absolute; inset: 12px; z-index: 5; pointer-events: none;
    display: grid; place-content: center; justify-items: center; gap: 12px;
    border: 1.5px solid var(--primary); border-radius: var(--r-lg);
    background: rgb(23 23 33 / 0.94);
    animation: fade 0.15s ease-out;
  }
  @keyframes fade { from { opacity: 0; } }

  @media (max-width: 900px) {
    main { overflow-y: auto; }
    .work { grid-template-columns: 1fr; height: auto; }
    .files { height: auto; max-height: 55vh; }
  }
</style>
