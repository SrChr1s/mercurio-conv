# Mercurio

Conversor de imágenes por lotes para escritorio (Windows y Linux). Arrastra las imágenes, elige el formato y conviértelas todas de una vez. Todo se procesa en tu equipo: nada se sube a ningún servidor.

Hecho con [Wails](https://wails.io) (Go + Svelte). La conversión es Go puro, sin dependencias de C ni herramientas externas.

## Qué hace

- **Entrada:** JPEG, PNG, WebP, GIF (primer fotograma), BMP, TIFF y SVG.
- **Salida:** JPEG, PNG, WebP, GIF, BMP y TIFF.
- **Formato por imagen:** el formato global se puede cambiar para imágenes concretas desde el menú de cada fila.
- **Calidad** para JPEG y WebP, y **tamaño máximo** (se ajusta dentro del rectángulo sin deformar ni ampliar; los SVG sí se dibujan al tamaño pedido).
- **Metadatos fuera:** EXIF, GPS y datos de cámara se eliminan siempre. La orientación EXIF se aplica antes, para que las fotos de móvil no salgan giradas.
- **Destino:** una carpeta o un único ZIP. Si un nombre ya existe se guarda como `nombre (1).ext`.
- Arrastrar y soltar desde el gestor de archivos, incluido Wayland.

Límites conocidos: los SVG con texto se dibujan con una fuente del sistema (Liberation Sans, DejaVu o Noto) y no con la que pida el archivo, y los SVG muy pesados con filtros pueden tardar varios segundos a tamaños grandes.

## Instalar

Descarga el paquete desde la página de [Releases](https://github.com/SrChr1s/mercurio-conv/releases) y comprueba su huella con `SHA256SUMS.txt`.

### Windows

| Archivo | Para qué |
| --- | --- |
| `…-windows-amd64-setup.exe` | Instalador con acceso directo y desinstalador. |
| `…-windows-amd64-portable.zip` | Sin instalar: descomprime y ejecuta `Mercurio.exe`. |

Necesita **WebView2**, que ya viene con Windows 10/11 actualizados. El instalador lo descarga si falta; la versión portable no.

Los ejecutables no están firmados, así que Windows SmartScreen puede avisar la primera vez ("Más información" → "Ejecutar de todas formas").

### Linux

Requiere **GTK 3 y WebKitGTK 4.1** (`libwebkit2gtk-4.1` en Debian/Ubuntu, `webkit2gtk4.1` en Fedora).

| Archivo | Cómo |
| --- | --- |
| `….deb` | `sudo apt install ./mercurio-conversor_X.Y.Z_amd64.deb` |
| `….rpm` | `sudo dnf install ./mercurio-conversor-X.Y.Z-1.x86_64.rpm` |
| `….tar.gz` | Descomprime y ejecuta `./install.sh` (instala en `~/.local`; `--uninstall` lo quita). |
| `.AppImage` | `chmod +x Mercurio-*.AppImage && ./Mercurio-*.AppImage` |

El AppImage **no incluye WebKitGTK**: usa el del sistema, igual que los demás paquetes. Los binarios se compilan en Ubuntu 24.04, así que necesitan una glibc igual o más nueva (Debian 12, por ejemplo, queda fuera).

## Desarrollo

Necesitas Go, Node.js y la [CLI de Wails v2](https://wails.io/docs/gettingstarted/installation). En Linux, además, los paquetes de desarrollo de GTK 3 y WebKitGTK 4.1.

```bash
wails dev -tags webkit2_41      # modo desarrollo con recarga
wails build -tags webkit2_41    # binario en build/bin/
go test ./...                   # pruebas de la conversión
cd frontend && npm run check    # tipos de Svelte
```

`-tags webkit2_41` solo hace falta en Linux con WebKitGTK 4.1 (la mayoría de distribuciones actuales). En Windows se omite.

Para trabajar solo el aspecto, `cd frontend && npm run dev` sirve la interfaz en el navegador con un simulador del backend (`src/dev-mock.ts`). Añadir `?state=list`, `running`, `done`, `drag` o `empty-drag` a la URL carga cada estado.

### Estructura

```
main.go, app.go        Ventana Wails y métodos expuestos al frontend
internal/convert/      Decodificación, redimensionado, codificación, lotes y ZIP
frontend/src/          Interfaz Svelte (App.svelte, lib/)
packaging/linux/       .desktop, icono, nfpm.yaml e install.sh
scripts/package.sh     Genera todos los paquetes en dist/
build/                 Icono (appicon.svg es la fuente) y plantillas de Wails
DESIGN.md              Sistema de diseño en el que se basa la interfaz
```

## Publicar una versión

Los paquetes los genera `scripts/package.sh` (`linux`, `windows` o `all`) y salen en `dist/`. Para una release:

```bash
git tag v0.2.0
git push origin v0.2.0
```

El flujo `.github/workflows/release.yml` pasa las pruebas, compila Windows y Linux, y publica la Release con los paquetes y `SHA256SUMS.txt`. La versión sale de la etiqueta (`vX.Y.Z`).

En local, `scripts/package.sh` necesita además `nfpm` (deb/rpm), `makensis` y `zip` (Windows) y, opcionalmente, `appimagetool` (si no está, se omite el AppImage). Windows se compila desde Linux, sin cgo.

El icono se dibuja en `build/appicon.svg`; `build/appicon.png` (1024 px) y `packaging/linux/mercurio-conversor.png` (512 px) se generan a partir de él.
