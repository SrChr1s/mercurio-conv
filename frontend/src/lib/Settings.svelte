<script lang="ts">
  import { PickOutputDir } from '../../wailsjs/go/main/App'
  import Icon from './Icon.svelte'
  import { FORMATS, SIZE_PRESETS, settings, app } from './stores.svelte'

  let { canConvert, reason, count, onconvert, oncancel, onsavezip }: {
    canConvert: boolean
    reason: string
    count: number
    onconvert: () => void
    oncancel: () => void
    onsavezip: () => void
  } = $props()

  const lossy = $derived(FORMATS.find((f) => f.value === settings.format)?.lossy ?? false)
  const done = $derived(app.rows.filter((r) => r.status === 'done' || r.status === 'error').length)
  const pct = $derived(app.rows.length ? (done / app.rows.length) * 100 : 0)
  const presetOn = (n: number) => settings.maxWidth === n && settings.maxHeight === n

  async function chooseDir() {
    const d = await PickOutputDir()
    if (d) settings.outDir = d
  }

  const num = (v: string) => Math.max(0, Math.floor(Number(v) || 0))
  const dirLabel = $derived(settings.outDir ? settings.outDir.split('/').filter(Boolean).slice(-2).join(' / ') : 'Elegir carpeta…')
</script>

<aside class="panel">
  <div class="group">
    <span class="eyebrow">Formato</span>
    <div class="formats" role="radiogroup" aria-label="Formato de salida">
      {#each FORMATS as f}
        <button role="radio" aria-checked={settings.format === f.value} class="fmt"
          class:on={settings.format === f.value} onclick={() => (settings.format = f.value)} disabled={app.running}>
          <b>{f.label}</b><small>{f.hint}</small>
        </button>
      {/each}
    </div>
  </div>

  <div class="group">
    <div class="row">
      <label class="eyebrow" for="q">Calidad</label>
      <span class="val">{lossy ? settings.quality : 'Sin pérdida'}</span>
    </div>
    <input id="q" type="range" min="1" max="100" bind:value={settings.quality}
      style="--p:{settings.quality}%" disabled={!lossy || app.running} />
  </div>

  <div class="group">
    <span class="eyebrow">Tamaño máximo</span>
    <div class="chips">
      <button class="chip" class:on={!settings.maxWidth && !settings.maxHeight}
        onclick={() => { settings.maxWidth = 0; settings.maxHeight = 0 }} disabled={app.running}>Original</button>
      {#each SIZE_PRESETS as n}
        <button class="chip" class:on={presetOn(n)}
          onclick={() => { settings.maxWidth = n; settings.maxHeight = n }} disabled={app.running}>{n}</button>
      {/each}
    </div>
    <div class="dims">
      <label><input type="number" min="0" placeholder="Auto" aria-label="Ancho máximo"
        value={settings.maxWidth || ''} oninput={(e) => (settings.maxWidth = num(e.currentTarget.value))} disabled={app.running} /><i>An</i></label>
      <span>×</span>
      <label><input type="number" min="0" placeholder="Auto" aria-label="Alto máximo"
        value={settings.maxHeight || ''} oninput={(e) => (settings.maxHeight = num(e.currentTarget.value))} disabled={app.running} /><i>Al</i></label>
    </div>
  </div>

  <div class="group">
    <span class="eyebrow">Guardar en</span>
    <div class="seg" role="radiogroup" aria-label="Destino">
      <button role="radio" aria-checked={settings.output === 'folder'} class:on={settings.output === 'folder'}
        onclick={() => (settings.output = 'folder')} disabled={app.running}><Icon name="folder" size={15} /> Carpeta</button>
      <button role="radio" aria-checked={settings.output === 'zip'} class:on={settings.output === 'zip'}
        onclick={() => (settings.output = 'zip')} disabled={app.running}><Icon name="archive" size={15} /> ZIP</button>
    </div>
    {#if settings.output === 'folder'}
      <button class="dir" onclick={chooseDir} disabled={app.running} title={settings.outDir}>
        <Icon name="folder" size={15} /><span>{dirLabel}</span>
      </button>
    {/if}
  </div>

  <div class="actions">
    {#if app.running}
      <div class="bar" role="progressbar" aria-valuemin="0" aria-valuemax="100" aria-valuenow={Math.round(pct)}><i style="width:{pct}%"></i></div>
      <div class="row"><span class="count">{done} de {app.rows.length}</span>
        <button class="pill tint small" onclick={oncancel}>Cancelar</button></div>
    {:else}
      {#if app.zipReady}
        <button class="pill primary block" onclick={onsavezip}><Icon name="archive" /> Guardar ZIP</button>
        <button class="pill tint block" onclick={onconvert} disabled={!canConvert}>Convertir de nuevo</button>
      {:else}
        <button class="pill primary block" onclick={onconvert} disabled={!canConvert}>
          Convertir {count} {count === 1 ? 'imagen' : 'imágenes'}
        </button>
      {/if}
      {#if reason}<p class="why">{reason}</p>{/if}
    {/if}
    <p class="note"><Icon name="shield" size={14} /> Los datos EXIF y de ubicación siempre se eliminan.</p>
  </div>
</aside>

<style>
  .panel {
    background: var(--night-raised); border-radius: var(--r-md); padding: 24px;
    display: grid; gap: 22px; align-content: start;
  }
  .group { display: grid; gap: 10px; }
  .row { display: flex; align-items: baseline; justify-content: space-between; }
  .val { font-size: 14px; color: var(--on-dark); font-variation-settings: 'wght' 480; }

  .formats { display: grid; grid-template-columns: repeat(3, 1fr); gap: 6px; }
  .fmt {
    display: grid; gap: 1px; justify-items: start; padding: 9px 10px; text-align: left;
    border: 0; border-radius: var(--r-sm); background: var(--night-lifted);
    transition: background 0.15s, box-shadow 0.15s;
  }
  .fmt b { font-size: 14px; font-weight: 500; font-variation-settings: 'wght' 500; }
  .fmt small { font-size: 11px; color: var(--fog); }
  .fmt:hover:not(:disabled) { background: #30304a; }
  .fmt.on { background: var(--primary); }
  .fmt.on small { color: rgb(255 255 255 / 0.8); }

  input[type='range'] {
    -webkit-appearance: none; appearance: none; width: 100%; height: 20px; background: transparent; margin: 0;
  }
  input[type='range']::-webkit-slider-runnable-track {
    height: 4px; border-radius: 4px;
    background: linear-gradient(to right, var(--primary) var(--p), var(--night-lifted) var(--p));
  }
  input[type='range']::-webkit-slider-thumb {
    -webkit-appearance: none; width: 16px; height: 16px; margin-top: -6px; border-radius: 50%;
    background: var(--on-dark); box-shadow: 0 0 0 4px rgb(82 102 235 / 0.35);
  }
  input[type='range']:disabled { opacity: 0.35; }

  .chips { display: flex; flex-wrap: wrap; gap: 6px; }
  .chip, .seg button {
    height: 30px; padding: 0 12px; border: 0; border-radius: var(--r-pill);
    background: var(--night-lifted); font-size: 13px; font-variation-settings: 'wght' 440;
    transition: background 0.15s;
  }
  .chip:hover:not(:disabled), .seg button:hover:not(:disabled) { background: #30304a; }
  .chip.on, .seg button.on { background: var(--primary); color: var(--on-primary); }

  .dims { display: flex; align-items: center; gap: 8px; color: var(--fog); }
  .dims label { position: relative; flex: 1; }
  .dims input {
    width: 100%; height: 38px; padding: 0 30px 0 14px; border: 0; border-radius: var(--r-pill);
    background: var(--night); box-shadow: inset 0 0 0 1px var(--line-strong); font-size: 14px;
  }
  .dims input:focus { outline: none; box-shadow: inset 0 0 0 1px var(--periwinkle); }
  .dims i { position: absolute; right: 14px; top: 50%; translate: 0 -50%; font-style: normal; font-size: 11px; color: var(--fog); pointer-events: none; }
  input[type='number']::-webkit-inner-spin-button { display: none; }

  .seg { display: grid; grid-template-columns: 1fr 1fr; gap: 6px; }
  .seg button { display: inline-flex; align-items: center; justify-content: center; gap: 6px; height: 34px; }
  .dir {
    display: flex; align-items: center; gap: 8px; height: 38px; padding: 0 14px; border: 0;
    border-radius: var(--r-pill); background: var(--tint); font-size: 14px; text-align: left; min-width: 0;
  }
  .dir span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .dir:hover:not(:disabled) { background: var(--tint-hover); }

  .actions { display: grid; gap: 10px; padding-top: 18px; border-top: 1px solid var(--line); }
  .why { font-size: 13px; color: var(--fog); text-align: center; }
  .note { display: flex; align-items: center; gap: 6px; justify-content: center; font-size: 12px; letter-spacing: 0.2px; color: var(--fog); opacity: 0.8; }
  .count { font-size: 14px; color: var(--fog); }
  .bar { height: 4px; border-radius: 4px; background: var(--night-lifted); overflow: hidden; }
  .bar i { display: block; height: 100%; background: var(--primary); border-radius: 4px; transition: width 0.3s; }
</style>
