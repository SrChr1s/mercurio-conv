<script lang="ts">
  const paths: Record<string, string[]> = {
    plus: ['M12 5v14M5 12h14'],
    x: ['M6 6l12 12M18 6L6 18'],
    folder: ['M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z'],
    shield: ['M12 3l7 3v5c0 4.5-3 8-7 10-4-2-7-5.5-7-10V6z', 'M9 12l2.2 2.2L15 10'],
    alert: ['M12 3a9 9 0 1 0 0 18 9 9 0 0 0 0-18z', 'M12 8v5M12 16.2v.01'],
    check: ['M5 12.5l4.5 4.5L19 7.5'],
    archive: ['M4 7h16v12H4zM4 7l2-3h12l2 3', 'M10 11h4'],
    image: ['M4 5h16v14H4z', 'M4 16l5-5 4 4 3-3 4 4', 'M9 9.5h.01'],
    arrow: ['M5 12h14M13 6l6 6-6 6'],
    chevron: ['M6 9l6 6 6-6'],
    minimize: ['M6 12h12'],
    maximize: ['M6.5 6.5h11v11h-11z'],
    restore: ['M9 9V6.5h8.5V15H15', 'M6.5 9h8.5v8.5H6.5z'],
  }
  let { name, size = 16 }: { name: keyof typeof paths | string; size?: number } = $props()
</script>

<svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor"
  stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
  {#each paths[name] ?? [] as d}<path {d} />{/each}
</svg>
