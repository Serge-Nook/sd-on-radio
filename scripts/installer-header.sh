#!/bin/sh
set -eu

APP_NAME="SD-ON RADIO"
APP_VERSION="__APP_VERSION__"
DATA_HOME="${XDG_DATA_HOME:-$HOME/.local/share}"
APP_HOME="$DATA_HOME/sd-on-radio"
BIN_HOME="$HOME/.local/bin"
DESKTOP_HOME="$DATA_HOME/applications"
LAUNCHER="$BIN_HOME/sd-on-radio"
DESKTOP_FILE="$DESKTOP_HOME/sd-on-radio.desktop"
PAYLOAD_SHA256="__PAYLOAD_SHA256__"

print_help() {
  cat <<EOF
$APP_NAME $APP_VERSION

Использование:
  $0                 Установить или обновить приложение
  $0 --install       Установить или обновить приложение
  $0 --uninstall     Удалить приложение и пользовательские настройки
  $0 --help          Показать эту справку
EOF
}

remove_installation() {
  if [ -L "$LAUNCHER" ]; then
    rm -f "$LAUNCHER"
  fi
  rm -f "$DESKTOP_FILE"
  rm -rf "$APP_HOME"

  if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database "$DESKTOP_HOME" >/dev/null 2>&1 || true
  fi

  printf '%s\n' "$APP_NAME удалён."
}

install_application() {
  case "$(uname -m)" in
    x86_64|amd64) ;;
    *)
      printf '%s\n' "Ошибка: поддерживается только архитектура x86_64." >&2
      exit 1
      ;;
  esac

  temporary_directory="$(mktemp -d)"
  trap 'rm -rf "$temporary_directory"' EXIT HUP INT TERM
  payload_line="$(awk '/^__SD_ON_RADIO_PAYLOAD__$/ { print NR + 1; exit }' "$0")"

  if [ -z "$payload_line" ]; then
    printf '%s\n' "Ошибка: установочный архив не найден." >&2
    exit 1
  fi

  tail -n +"$payload_line" "$0" > "$temporary_directory/payload.tar.gz"
  actual_sha256="$(sha256sum "$temporary_directory/payload.tar.gz" | awk '{ print $1 }')"
  if [ "$actual_sha256" != "$PAYLOAD_SHA256" ]; then
    printf '%s\n' "Ошибка: установочный файл повреждён." >&2
    exit 1
  fi

  mkdir -p "$APP_HOME" "$BIN_HOME" "$DESKTOP_HOME"
  tar -xzf "$temporary_directory/payload.tar.gz" -C "$temporary_directory"
  rm -rf "$APP_HOME/app"
  mv "$temporary_directory/app" "$APP_HOME/app"
  chmod +x "$APP_HOME/app/sd-on-radio"
  ln -sfn "$APP_HOME/app/sd-on-radio" "$LAUNCHER"

  cat > "$DESKTOP_FILE" <<EOF
[Desktop Entry]
Type=Application
Name=SD-ON RADIO
Comment=Интернет-радио для Steam Deck
Exec=$LAUNCHER
Icon=$APP_HOME/app/resources/app/src/renderer/assets/icon.svg
Terminal=false
Categories=Audio;AudioVideo;Player;
StartupNotify=true
StartupWMClass=SD-ON RADIO
EOF
  chmod 0644 "$DESKTOP_FILE"

  if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database "$DESKTOP_HOME" >/dev/null 2>&1 || true
  fi

  printf '%s\n' "$APP_NAME $APP_VERSION установлен."
  printf '%s\n' "Запуск: $LAUNCHER"
  printf '%s\n' "Ярлык добавлен в меню приложений."
}

case "${1:---install}" in
  --install) install_application ;;
  --uninstall) remove_installation ;;
  --help|-h) print_help ;;
  *)
    printf 'Неизвестный параметр: %s\n' "$1" >&2
    print_help >&2
    exit 2
    ;;
esac

exit 0
__SD_ON_RADIO_PAYLOAD__
