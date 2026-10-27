import { PickFiles } from '../../wailsjs/go/main/App'

export type Fmt = 'jpeg' | 'png' | 'webp' | 'gif' | 'bmp' | 'tiff'

export interface FileInfo {
  id: string
  name: string
  size: number
  format: string
  width: number
  height: number
  thumb: string
}

export interface Row extends FileInfo {
  status: 'queued' | 'working' | 'done' | 'error'
  target?: Fmt // per-image output format; unset = follow the global format
  outName?: string
  outSize?: number
  error?: string
}

export const FORMATS: { value: Fmt; label: string; hint: string; lossy: boolean }[] = [
  { value: 'webp', label: 'WebP', hint: 'Más liviano', lossy: true },
  { value: 'jpeg', label: 'JPEG', hint: 'Fotos', lossy: true },
  { value: 'png', label: 'PNG', hint: 'Sin pérdida', lossy: false },
  { value: 'gif', label: 'GIF', hint: '256 colores', lossy: false },
  { value: 'bmp', label: 'BMP', hint: 'Clásico', lossy: false },
  { value: 'tiff', label: 'TIFF', hint: 'Archivo', lossy: false },
]

export const SIZE_PRESETS = [1920, 1280, 800]

export const settings = $state({
  format: 'webp' as Fmt,
  quality: 80,
  maxWidth: 0,
  maxHeight: 0,
  output: 'folder' as 'folder' | 'zip',
  outDir: '',
  suffix: '',
})

export const app = $state({
  rows: [] as Row[],
  running: false,
  zipReady: false,
  message: '',
  dragging: false,
})

export async function pickFiles() {
  try {
    addFiles((await PickFiles()) ?? [])
  } catch (e) {
    app.message = String(e)
  }
}

export function addFiles(files: FileInfo[]) {
  for (const f of files) app.rows.push({ ...f, status: 'queued' })
}

export function formatBytes(n: number): string {
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / 1024 / 1024).toFixed(1)} MB`
}
