#!/bin/sh
# Instala Mercurio para el usuario actual (sin root) desde el .tar.gz.
#   ./install.sh              instala en ~/.local
#   ./install.sh --uninstall  lo quita
set -eu
here=$(cd "$(dirname "$0")" && pwd)
prefix="${XDG_DATA_HOME:-$HOME/.local/share}"
bin="$HOME/.local/bin"

if [ "${1:-}" = "--uninstall" ]; then
  rm -f "$bin/mercurio-conversor" \
        "$prefix/applications/mercurio-conversor.desktop" \
        "$prefix/icons/hicolor/512x512/apps/mercurio-conversor.png"
  echo "Mercurio desinstalado."
  exit 0
fi

mkdir -p "$bin" "$prefix/applications" "$prefix/icons/hicolor/512x512/apps"
install -m 755 "$here/mercurio-conversor" "$bin/mercurio-conversor"
install -m 644 "$here/mercurio-conversor.png" "$prefix/icons/hicolor/512x512/apps/mercurio-conversor.png"
sed "s|^Exec=.*|Exec=$bin/mercurio-conversor|" "$here/mercurio-conversor.desktop" > "$prefix/applications/mercurio-conversor.desktop"
echo "Mercurio instalado en $bin. Necesita webkit2gtk 4.1 (libwebkit2gtk-4.1 / webkit2gtk4.1)."
