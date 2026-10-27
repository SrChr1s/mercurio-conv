<script lang="ts">
  import Planet from './Planet.svelte'
  import Icon from './Icon.svelte'
  import { app, pickFiles } from './stores.svelte'

  // Deterministic starfield so it doesn't reshuffle on every render.
  let seed = 7
  const rnd = () => ((seed = (seed * 16807) % 2147483647) / 2147483647)
  const stars = Array.from({ length: 34 }, () => ({
    x: rnd() * 100, y: rnd() * 100, r: 0.6 + rnd() * 1.4, o: 0.2 + rnd() * 0.5, d: rnd() * 6,
  }))
</script>

<section class="hero" class:active={app.dragging}>
  <div class="sky" aria-hidden="true">
    {#each stars as s, i}
      <i class:tw={i % 3 === 0} style="left:{s.x}%;top:{s.y}%;width:{s.r}px;height:{s.r}px;opacity:{s.o};animation-delay:{s.d}s"></i>
    {/each}
  </div>

  <Planet size={300} active={app.dragging} />

  <div class="copy">
    <h1>{app.dragging ? 'Suéltalas para empezar' : 'Pon tus imágenes en órbita'}</h1>
    <p>Suelta todas las que quieras. JPEG, PNG, WebP, GIF, BMP, TIFF y SVG.</p>
  </div>

  <button class="pill primary" onclick={pickFiles}><Icon name="plus" /> Añadir imágenes</button>
  <span class="hint">o arrástralas a cualquier parte de esta ventana</span>
</section>

<style>
  .hero {
    position: relative; isolation: isolate; overflow: hidden;
    height: 100%; min-height: 420px;
    display: grid; place-content: center; justify-items: center; gap: 20px;
    padding: 32px; text-align: center;
    background:
      radial-gradient(60% 55% at 50% 38%, rgb(82 102 235 / 0.14), transparent 70%),
      var(--night-raised);
    border: 1px dashed var(--line-strong);
    border-radius: var(--r-lg);
    transition: border-color 0.25s, background-color 0.25s;
  }
  .hero::after {
    content: ''; position: absolute; inset: 0; z-index: -1; border-radius: inherit; pointer-events: none;
    box-shadow: inset 0 0 64px rgb(82 102 235 / 0.28); opacity: 0; transition: opacity 0.25s;
  }
  .hero.active::after { opacity: 1; }
  .hero.active { border-color: var(--primary); border-style: solid; background-color: #20213a; box-shadow: 0 0 0 1px var(--primary); }
  .sky { position: absolute; inset: 0; z-index: -1; }
  .sky i { position: absolute; border-radius: 50%; background: var(--on-dark); }
  .sky i.tw { animation: twinkle 5s ease-in-out infinite; }
  /* Keep the hover state cheap: nothing animates while a file is held over the window. */
  .hero.active .sky i { animation: none; }
  .copy { display: grid; gap: 8px; max-width: 520px; }
  h1 { font-size: 36px; line-height: 1.15; letter-spacing: 0.4px; text-wrap: balance; }
  p { color: var(--fog); font-size: 17px; text-wrap: balance; font-variation-settings: 'wght' 360; }
  .hint { color: var(--fog); font-size: 13px; opacity: 0.75; margin-top: -8px; }
  @keyframes twinkle { 50% { opacity: 0.12; } }
</style>
