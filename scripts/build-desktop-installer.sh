#!/bin/sh
# Wraps dist/sd-on-radio-installer.sh into a self-contained .desktop file that
# can be launched by double-click in KDE Plasma (SteamOS desktop mode).
set -eu

ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
DIST="$ROOT/dist"
INPUT="$DIST/sd-on-radio-installer.sh"
OUTPUT="$DIST/sd-on-radio-installer.desktop"

if [ ! -f "$INPUT" ]; then
  printf '%s\n' "Ошибка: сначала выполните scripts/build-installer.sh." >&2
  exit 1
fi

SCRIPT_SHA256="$(sha256sum "$INPUT" | awk '{ print $1 }')"

{
  cat <<EOF
#!/bin/sh
[Desktop Entry]
Type=Application
Version=1.0
# This key doubles as the shell entrypoint invoked by Exec.
StartupWMClass=; self="\$0"; script="\$(mktemp --suffix=-sd-on-radio-installer.sh)"; trap 'rm -f "\$script"' EXIT HUP INT TERM; sed -n '/^# __SD_ON_RADIO_INSTALLER__\$/,\$p' "\$self" | sed '1d; s/^# //' | base64 -d > "\$script" || { command -v kdialog >/dev/null 2>&1 && kdialog --error "Установщик повреждён."; exit 1; }; actual="\$(sha256sum "\$script" | awk '{ print \$1 }')"; if [ "\$actual" != "$SCRIPT_SHA256" ]; then command -v kdialog >/dev/null 2>&1 && kdialog --error "Установщик повреждён."; exit 1; fi; chmod +x "\$script"; sh "\$script"; result=\$?; rm -f "\$script"; trap - EXIT HUP INT TERM; exit "\$result"
Name=SD-ON RADIO — установка и удаление
Name[ru]=SD-ON RADIO — установка и удаление
Comment=Установить, обновить или удалить SD-ON RADIO
Comment[ru]=Установить, обновить или удалить SD-ON RADIO
Exec=/bin/sh %k
Icon=system-software-install
Terminal=false
Categories=Audio;AudioVideo;Player;
Keywords=radio;steamdeck;install;uninstall;
StartupNotify=true
# __SD_ON_RADIO_INSTALLER__
EOF
  base64 -w 76 "$INPUT" | sed 's/^/# /'
} > "$OUTPUT"
chmod 0755 "$OUTPUT"

printf '%s\n' "Готово: $OUTPUT"
printf '%s\n' "SHA-256: $(sha256sum "$OUTPUT" | awk '{ print $1 }')"
