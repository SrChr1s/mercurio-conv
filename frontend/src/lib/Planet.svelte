<script lang="ts">
  // Mercury: cratered, sunlit from the upper left, with a tilted orbit ring.
  let { size = 240, mini = false, active = false }: { size?: number; mini?: boolean; active?: boolean } = $props()
  const uid = $props.id()

  const craters: [number, number, number][] = [
    [-22, -20, 14], [24, -32, 9], [32, 14, 16], [-36, 20, 10],
    [2, 36, 7], [-8, -48, 6], [48, -4, 5], [-46, -8, 8], [14, 6, 5], [-14, 48, 5],
  ]
</script>

<svg
  class="planet" class:mini class:active
  width={size} height={size}
  viewBox={mini ? '-70 -70 140 140' : '-112 -112 224 224'}
  aria-hidden="true"
>
  <defs>
    <radialGradient id="{uid}-body" cx="0.34" cy="0.3" r="0.85">
      <stop offset="0" stop-color="#ecebf5" />
      <stop offset="0.38" stop-color="#b3aecb" />
      <stop offset="0.78" stop-color="#5b5774" />
      <stop offset="1" stop-color="#2a2839" />
    </radialGradient>
    <radialGradient id="{uid}-shade" cx="0.3" cy="0.26" r="0.95">
      <stop offset="0.45" stop-color="#0e0e18" stop-opacity="0" />
      <stop offset="1" stop-color="#0e0e18" stop-opacity="0.72" />
    </radialGradient>
    <radialGradient id="{uid}-crater" cx="0.42" cy="0.42" r="0.6">
      <stop offset="0" stop-color="#1a1928" stop-opacity="0.38" />
      <stop offset="0.72" stop-color="#1a1928" stop-opacity="0.2" />
      <stop offset="0.9" stop-color="#ffffff" stop-opacity="0.2" />
      <stop offset="1" stop-color="#ffffff" stop-opacity="0" />
    </radialGradient>
    <radialGradient id="{uid}-glow">
      <stop offset="0.45" stop-color="#5266eb" stop-opacity="0.34" />
      <stop offset="1" stop-color="#5266eb" stop-opacity="0" />
    </radialGradient>
    <clipPath id="{uid}-clip"><circle r="64" /></clipPath>
  </defs>

  {#if !mini}
    <circle class="glow" r="118" fill="url(#{uid}-glow)" />
    <g transform="rotate(-18)">
      <path d="M -108 0 A 108 34 0 0 1 108 0" class="ring back" />
    </g>
  {/if}

  <circle r="64" fill="url(#{uid}-body)" />
  <g clip-path="url(#{uid}-clip)">
    {#each craters as [x, y, r]}
      <circle cx={x} cy={y} {r} fill="url(#{uid}-crater)" />
    {/each}
    <circle r="64" fill="url(#{uid}-shade)" />
  </g>

  {#if !mini}
    <g transform="rotate(-18)">
      <path d="M -108 0 A 108 34 0 0 0 108 0" class="ring front" />
    </g>
  {/if}
</svg>

<style>
  .planet { display: block; overflow: visible; transition: transform 0.4s cubic-bezier(0.2, 0.7, 0.2, 1); }
  .planet:not(.mini) { animation: float 9s ease-in-out infinite; will-change: transform; }
  .planet.active { animation-play-state: paused; }
  .planet.active { transform: scale(1.06); }
  .ring { fill: none; stroke-linecap: round; }
  .ring.back { stroke: rgb(156 180 232 / 0.16); stroke-width: 1.2; }
  .ring.front { stroke: rgb(156 180 232 / 0.5); stroke-width: 1.6; }
  .glow { opacity: 0.75; transform-box: fill-box; transform-origin: center; transition: opacity 0.3s, transform 0.4s; }
  .planet.active .glow { opacity: 1; transform: scale(1.18); }
  @keyframes float { 0%, 100% { translate: 0 0; } 50% { translate: 0 -8px; } }
</style>
