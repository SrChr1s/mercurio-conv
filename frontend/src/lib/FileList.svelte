<script lang="ts">
  import { Remove } from '../../wailsjs/go/main/App'
  import Icon from './Icon.svelte'
  import { FORMATS, app, settings, formatBytes, type Fmt, type Row } from './stores.svelte'

  function remove(id: string) {
    Remove([id])
    app.rows = app.rows.filter((r) => r.id !== id)
  }

  const saved = (src: number, out = 0) => Math.round((1 - out / src) * 100)

  const ext = (f: Fmt) => (f === 'jpeg' ? 'JPG' : f.toUpperCase())
  const effective = (r: Row): Fmt => r.target ?? settings.format

  // Per-image format menu. Fixed-positioned so the scrolling list can't clip it.
  let menu = $state<{ id: string; x: number; y: number } | null>(null)
  const MENU_W = 220
  const MENU_H = 8 + (FORMATS.length + 1) * 44

  function openMenu(e: MouseEvent, r: Row) {
    e.stopPropagation()
    if (menu?.id === r.id) return (menu = null)
    const b = (e.currentTarget as HTMLElement).getBoundingClientRect()
    const below = b.bottom + 6 + MENU_H < window.innerHeight
    menu = {
      id: r.id,
      x: Math.max(8, Math.min(b.right - MENU_W, window.innerWidth - MENU_W - 8)),
      y: below ? b.bottom + 6 : Math.max(8, b.top - 6 - MENU_H),
    }
  }

  function choose(id: string, f: Fmt | undefined) {
    const r = app.rows.find((r) => r.id === id)
    if (r) {
      r.target = f
      // A result for another format is stale now; the row is ready to be converted again.
      if (r.status === 'done' || r.status === 'error') Object.assign(r, { status: 'queued', outName: undefined, outSize: undefined, error: undefined })
    }
    menu = null
  }

  const menuRow = $derived(menu ? app.rows.find((r) => r.id === menu!.id) : undefined)
</script>

<svelte:window onclick={() => (menu = null)} onkeydown={(e) => e.key === 'Escape' && (menu = null)} onscrollcapture={() => (menu = null)} />

<ul>
  {#each app.rows as r (r.id)}
    <li class:error={r.status === 'error'}>
      <img src={r.thumb} alt="" />

      <div class="meta">
        <span class="name" title={r.name}>{r.name}</span>
        <span class="sub">
          <b>{r.format.toUpperCase()}</b>{r.width}×{r.height} · {formatBytes(r.size)}
        </span>
      </div>

      <div class="status">
        {#if r.status === 'queued'}
          <button class="badge" class:custom={!!r.target} onclick={(e) => openMenu(e, r)} disabled={app.running}
            aria-haspopup="menu" aria-expanded={menu?.id === r.id} title="Elegir formato para esta imagen">
            <Icon name="arrow" size={13} /> {ext(effective(r))} <Icon name="chevron" size={12} />
          </button>
        {:else if r.status === 'working'}
          <span class="spin" role="status" aria-label="Convirtiendo"></span>
        {:else if r.status === 'done'}
          {@const pct = saved(r.size, r.outSize)}
          <button class="done pick" onclick={(e) => openMenu(e, r)} disabled={app.running} aria-haspopup="menu" title="Cambiar formato de esta imagen">
            <span class="out">{formatBytes(r.outSize ?? 0)}
              <em class:up={pct < 0}>{pct >= 0 ? '−' : '+'}{Math.abs(pct)}%</em>
            </span>
            <span class="sub" title={r.outName}>{r.outName}</span>
          </button>
        {:else}
          <button class="done pick" onclick={(e) => openMenu(e, r)} disabled={app.running} aria-haspopup="menu" title="Cambiar formato de esta imagen">
            <span class="out fail"><Icon name="alert" size={15} /> No se pudo convertir</span>
            <span class="sub" title={r.error}>{r.error}</span>
          </button>
        {/if}
      </div>

      <button class="x" aria-label="Quitar {r.name}" onclick={() => remove(r.id)} disabled={app.running}>
        <Icon name="x" size={15} />
      </button>
    </li>
  {/each}
</ul>

{#if menu && menuRow}
  <div class="menu" role="menu" tabindex="-1" style="left:{menu.x}px;top:{menu.y}px;width:{MENU_W}px">
    <button role="menuitemradio" aria-checked={!menuRow.target} class:on={!menuRow.target} onclick={() => choose(menuRow.id, undefined)}>
      <span><b>Igual que el resto</b><small>{ext(settings.format)}</small></span>
      {#if !menuRow.target}<Icon name="check" size={15} />{/if}
    </button>
    <hr />
    {#each FORMATS as f}
      <button role="menuitemradio" aria-checked={menuRow.target === f.value} class:on={menuRow.target === f.value} onclick={() => choose(menuRow.id, f.value)}>
        <span><b>{f.label}</b><small>{f.hint}</small></span>
        {#if menuRow.target === f.value}<Icon name="check" size={15} />{/if}
      </button>
    {/each}
  </div>
{/if}

<style>
  ul { list-style: none; margin: 0; padding: 0; }
  li {
    display: grid; grid-template-columns: 52px minmax(0, 1fr) auto 32px;
    align-items: center; gap: 16px; padding: 10px 8px 10px 16px;
    border-bottom: 1px solid var(--line);
    transition: background 0.15s;
  }
  li:last-child { border-bottom: 0; }
  li:hover { background: rgb(237 237 243 / 0.03); }
  img {
    width: 52px; height: 52px; object-fit: cover; border-radius: var(--r-sm);
    background:
      conic-gradient(#2a2a3a 25%, #232331 0 50%, #2a2a3a 0 75%, #232331 0) 0 0 / 12px 12px;
    box-shadow: inset 0 0 0 1px var(--line);
  }
  .meta { display: grid; gap: 2px; min-width: 0; }
  .name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 15px; font-variation-settings: 'wght' 440; }
  .sub { font-size: 13px; color: var(--fog); letter-spacing: 0.07px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .sub b {
    font-weight: 500; font-variation-settings: 'wght' 500; font-size: 11px; letter-spacing: 0.6px;
    padding: 2px 6px; margin-right: 8px; border-radius: var(--r-xs); background: var(--night-lifted); color: var(--on-dark);
  }
  .status { display: flex; justify-content: flex-end; min-width: 0; max-width: 260px; }
  .badge {
    display: inline-flex; align-items: center; gap: 6px; height: 26px; padding: 0 10px;
    border-radius: var(--r-pill); background: var(--tint); font-size: 12px; letter-spacing: 0.6px; font-variation-settings: 'wght' 500;
  }
  .done { display: grid; gap: 2px; justify-items: end; min-width: 0; }
  .out { display: inline-flex; align-items: center; gap: 8px; font-size: 15px; font-variation-settings: 'wght' 440; }
  .out em { font-style: normal; font-size: 13px; color: var(--periwinkle); }
  .out em.up { color: var(--fog); }
  .out.fail { color: var(--on-dark); }
  li.error { background: rgb(237 237 243 / 0.025); }
  .x {
    display: grid; place-items: center; width: 32px; height: 32px; border: 0; border-radius: 9999px;
    background: transparent; color: var(--fog); opacity: 0; transition: opacity 0.15s, background 0.15s;
  }
  li:hover .x, .x:focus-visible { opacity: 1; }
  .x:hover:not(:disabled) { background: var(--tint); }
  .spin {
    width: 20px; height: 20px; border-radius: 50%;
    border: 2px solid var(--tint-hover); border-top-color: var(--periwinkle);
    animation: spin 0.8s linear infinite;
  }
  @keyframes spin { to { rotate: 360deg; } }
  .badge { border: 0; color: inherit; transition: background 0.15s; }
  .badge:hover:not(:disabled) { background: var(--tint-hover); }
  .badge.custom { background: var(--primary); color: var(--on-primary); }
  .badge.custom:hover:not(:disabled) { background: var(--primary); filter: brightness(1.12); }
  .pick { border: 0; background: transparent; color: inherit; padding: 4px 8px; margin: -4px -8px; border-radius: var(--r-sm); text-align: right; }
  .pick:hover:not(:disabled) { background: var(--tint); }

  .menu {
    position: fixed; z-index: 20; padding: 4px; border-radius: var(--r-md);
    background: var(--night-lifted); box-shadow: 0 0 0 1px var(--line-strong), 0 12px 32px rgb(0 0 0 / 0.45);
    animation: pop 0.12s ease-out;
  }
  .menu button {
    display: flex; align-items: center; justify-content: space-between; width: 100%; height: 44px; padding: 0 12px;
    border: 0; border-radius: var(--r-sm); background: transparent; text-align: left;
  }
  .menu button:hover { background: var(--tint); }
  .menu button.on { color: var(--periwinkle); }
  .menu span { display: grid; gap: 0; }
  .menu b { font-size: 14px; font-weight: 500; font-variation-settings: 'wght' 480; }
  .menu small { font-size: 11px; color: var(--fog); }
  .menu hr { border: 0; border-top: 1px solid var(--line); margin: 4px 8px; }
  @keyframes pop { from { opacity: 0; translate: 0 -4px; } }
</style>
