#!/usr/bin/env bash
# Builds release packages into dist/.
#
#   scripts/package.sh [linux|windows|all]
#
# VERSION (e.g. 0.2.0) defaults to the latest git tag, then to wails.json.
# Linux needs: wails, nfpm, gtk3 + webkit2gtk-4.1 dev headers. Optional: appimagetool (or $APPIMAGETOOL).
# Windows (cross-compiled, no cgo) needs: wails, makensis, zip.
set -euo pipefail
cd "$(dirname "$0")/.."

target="${1:-all}"
case "$target" in linux|windows|all) ;; *) echo "uso: $0 [linux|windows|all]" >&2; exit 2 ;; esac

VERSION="${VERSION:-$(git describe --tags --abbrev=0 2>/dev/null || true)}"
VERSION="${VERSION#v}"
VERSION="${VERSION:-$(sed -n 's/.*"productVersion": *"\([^"]*\)".*/\1/p' wails.json)}"
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "VERSION '$VERSION' no es x.y.z (el instalador de Windows lo exige)" >&2; exit 2
fi
export VERSION

# Wails reads the product version from wails.json; stamp it for the build and restore after.
backup=$(mktemp)
cp wails.json "$backup"
trap 'cp "$backup" wails.json; rm -f "$backup"' EXIT
sed -i "s/\"productVersion\": *\"[^\"]*\"/\"productVersion\": \"$VERSION\"/" wails.json

mkdir -p dist
LDFLAGS="-s -w"

build_linux() {
  echo "==> Linux $VERSION"
  wails build -clean -tags webkit2_41 -trimpath -ldflags "$LDFLAGS"

  nfpm package -p deb -f packaging/linux/nfpm.yaml -t dist/
  nfpm package -p rpm -f packaging/linux/nfpm.yaml -t dist/

  local name="mercurio-conversor-$VERSION-linux-amd64" stage
  stage=$(mktemp -d)
  mkdir "$stage/$name"
  cp build/bin/mercurio-conversor packaging/linux/{mercurio-conversor.desktop,mercurio-conversor.png,install.sh} "$stage/$name/"
  tar -C "$stage" -czf "dist/$name.tar.gz" "$name"
  rm -rf "$stage"

  local tool="${APPIMAGETOOL:-$(command -v appimagetool || true)}"
  if [[ -n "$tool" ]]; then
    local app
    app=$(mktemp -d)
    mkdir -p "$app/usr/bin"
    cp build/bin/mercurio-conversor "$app/usr/bin/"
    cp packaging/linux/mercurio-conversor.desktop packaging/linux/mercurio-conversor.png "$app/"
    printf '#!/bin/sh\nexec "$(dirname "$(readlink -f "$0")")/usr/bin/mercurio-conversor" "$@"\n' > "$app/AppRun"
    chmod +x "$app/AppRun"
    ARCH=x86_64 "$tool" "$app" "dist/Mercurio-$VERSION-x86_64.AppImage"
    rm -rf "$app"
  else
    echo "   (appimagetool no encontrado: se omite el AppImage)"
  fi
}

build_windows() {
  echo "==> Windows $VERSION"
  wails build -clean -platform windows/amd64 -nsis -trimpath -ldflags "$LDFLAGS"

  cp build/bin/mercurio-conversor-amd64-installer.exe "dist/mercurio-conversor-$VERSION-windows-amd64-setup.exe"

  local name="mercurio-conversor-$VERSION-windows-amd64-portable" stage
  stage=$(mktemp -d)
  mkdir "$stage/Mercurio"
  cp build/bin/mercurio-conversor.exe "$stage/Mercurio/Mercurio.exe"
  (cd "$stage" && zip -qr "$OLDPWD/dist/$name.zip" Mercurio)
  rm -rf "$stage"
}

[[ "$target" == linux   || "$target" == all ]] && build_linux
[[ "$target" == windows || "$target" == all ]] && build_windows

(cd dist && sha256sum -- *"$VERSION"* > "SHA256SUMS-$VERSION.txt")
echo "==> Listo:"; ls -lh dist
