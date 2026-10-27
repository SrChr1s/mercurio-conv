// Browser-only stand-in for the Wails bindings, so the UI can be previewed with `npm run dev`.
const w = window as any

const thumb = (a: string, b: string) =>
  'data:image/svg+xml;utf8,' +
  encodeURIComponent(
    `<svg xmlns="http://www.w3.org/2000/svg" width="96" height="96"><defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="${a}"/><stop offset="1" stop-color="${b}"/></linearGradient></defs><rect width="96" height="96" fill="url(#g)"/></svg>`,
  )

const sample = [
  ['IMG_2041.jpg', 'jpeg', 4032, 3024, 3_412_000, '#e8b98a', '#5b3a6b'],
  ['hero-banner.png', 'png', 2560, 1080, 1_880_000, '#9cb4e8', '#171721'],
  ['logo-mercurio.svg', 'svg', 512, 512, 4_300, '#5266eb', '#ededf3'],
  ['screenshot 2026-10-12 at 09.14.png', 'png', 1920, 1200, 612_000, '#70707d', '#c3c3cc'],
  ['portrait.webp', 'webp', 1600, 2000, 240_000, '#c3a6d9', '#3b3556'],
  ['scan-0003.tiff', 'tiff', 3300, 2550, 24_800_000, '#d9d6e6', '#8b86a3'],
] as const

let n = 0
const listeners: Record<string, Set<Function>> = {}
const emit = (name: string, data: unknown) => listeners[name]?.forEach((cb) => cb(data))
let cancelled = false

w.go = {
  main: {
    App: {
      PickFiles: async () =>
        sample.map(([name, format, width, height, size, a, b]) => ({
          id: `m${++n}`, name, format, width, height, size, thumb: thumb(a, b),
        })),
      AddPaths: async () => [],
      Remove: async () => {},
      PickOutputDir: async () => '/home/chris/Pictures/converted',
      Cancel: async () => { cancelled = true },
      SaveZip: async () => '/home/chris/converted-images.zip',
      Convert: async (req: any) => {
        cancelled = false
        let i = 0
        for (const id of req.ids) {
          await new Promise((r) => setTimeout(r, 700))
          if (cancelled) break
          i++
          const src = sample[(Number(id.slice(1)) - 1) % sample.length]
          const f = req.formats?.[id] ?? req.options.format
          const ext = f === 'jpeg' ? 'jpg' : f
          const base = String(src[0]).replace(/\.[^.]+$/, '')
          if (i === 4) emit('convert:progress', { id, status: 'error', error: 'formato de imagen no reconocido' })
          else emit('convert:progress', { id, status: 'done', outName: `${base}.${ext}`, outSize: Math.round(Number(src[4]) * 0.38) })
        }
        emit('convert:done', { ok: i - 1, failed: 1, cancelled })
      },
    },
  },
}

w.runtime = {
  EventsOnMultiple: (name: string, cb: Function) => {
    ;(listeners[name] ??= new Set()).add(cb)
    return () => listeners[name].delete(cb)
  },
  EventsOff: () => {},
  OnFileDrop: () => {},
  WindowMinimise: () => {},
  WindowToggleMaximise: () => {},
  WindowIsMaximised: async () => false,
  Quit: () => {},
  OnFileDropOff: () => {},
}

export {}
