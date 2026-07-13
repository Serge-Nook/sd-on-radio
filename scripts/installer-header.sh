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
DIALOG_BACKEND=""
SELECTED_ACTION=""

print_help() {
  cat <<EOF
$APP_NAME $APP_VERSION

Использование:
  $0                 Открыть меню установки и удаления
  $0 --install       Установить или обновить приложение
  $0 --uninstall     Удалить приложение и пользовательские настройки
  $0 --help          Показать эту справку
EOF
}

choose_action() {
  if [ -n "${DISPLAY:-}${WAYLAND_DISPLAY:-}" ] && command -v kdialog >/dev/null 2>&1; then
    DIALOG_BACKEND="kdialog"
    SELECTED_ACTION="$(
      kdialog \
        --title "$APP_NAME $APP_VERSION" \
        --menu "Выберите действие:" \
        install "Установить или обновить" \
        uninstall "Удалить приложение и настройки" \
        exit "Выйти"
    )" || SELECTED_ACTION="exit"
    return
  fi

  if [ -n "${DISPLAY:-}${WAYLAND_DISPLAY:-}" ] && command -v zenity >/dev/null 2>&1; then
    DIALOG_BACKEND="zenity"
    selection="$(
      zenity \
        --list \
        --title="$APP_NAME $APP_VERSION" \
        --text="Выберите действие:" \
        --column="Действие" \
        "Установить или обновить" \
        "Удалить приложение и настройки" \
        "Выйти"
    )" || selection="Выйти"

    case "$selection" in
      "Установить или обновить") SELECTED_ACTION="install" ;;
      "Удалить приложение и настройки") SELECTED_ACTION="uninstall" ;;
      *) SELECTED_ACTION="exit" ;;
    esac
    return
  fi

  DIALOG_BACKEND="console"
  while :; do
    printf '\n%s %s\n' "$APP_NAME" "$APP_VERSION"
    printf '%s\n' "1) Установить или обновить"
    printf '%s\n' "2) Удалить приложение и настройки"
    printf '%s\n' "3) Выйти"
    printf '%s' "Выберите действие [1-3]: "

    if ! IFS= read -r selection; then
      SELECTED_ACTION="exit"
      return
    fi

    case "$selection" in
      1) SELECTED_ACTION="install"; return ;;
      2) SELECTED_ACTION="uninstall"; return ;;
      3) SELECTED_ACTION="exit"; return ;;
      *) printf '%s\n' "Введите 1, 2 или 3." ;;
    esac
  done
}

confirm_removal() {
  case "$DIALOG_BACKEND" in
    kdialog)
      kdialog \
        --title "$APP_NAME" \
        --yesno "Удалить приложение, пользовательские станции и настройки?"
      ;;
    zenity)
      zenity \
        --question \
        --title="$APP_NAME" \
        --text="Удалить приложение, пользовательские станции и настройки?"
      ;;
    *)
      printf '%s' "Удалить приложение и все настройки? [y/N]: "
      IFS= read -r confirmation || return 1
      case "$confirmation" in
        y|Y|yes|YES|д|Д|да|ДА) return 0 ;;
        *) return 1 ;;
      esac
      ;;
  esac
}

show_completion() {
  message="$1"
  case "$DIALOG_BACKEND" in
    kdialog) kdialog --title "$APP_NAME" --msgbox "$message" || true ;;
    zenity) zenity --info --title="$APP_NAME" --text="$message" || true ;;
  esac
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
  show_completion "$APP_NAME удалён."
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
  show_completion "$APP_NAME $APP_VERSION установлен.

Запуск: $LAUNCHER
Ярлык добавлен в меню приложений."
}

if [ "$#" -eq 0 ]; then
  choose_action
  set -- "$SELECTED_ACTION"
fi

case "$1" in
  install|--install) install_application ;;
  uninstall)
    if confirm_removal; then
      remove_installation
    fi
    ;;
  --uninstall) remove_installation ;;
  exit) exit 0 ;;
  --help|-h) print_help ;;
  *)
    printf 'Неизвестный параметр: %s\n' "$1" >&2
    print_help >&2
    exit 2
    ;;
esac

exit 0
__SD_ON_RADIO_PAYLOAD__
