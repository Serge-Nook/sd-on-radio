#!/bin/sh
set -eu

APP_VERSION="2.0.0"
ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
DIST="$ROOT/dist"
WORK="$ROOT/.cache/installer"
PAYLOAD_ROOT="$WORK/payload"
PAYLOAD="$WORK/payload.tar.gz"
OUTPUT="$DIST/sd-on-radio-installer.sh"

if ! command -v go >/dev/null 2>&1; then
  printf '%s\n' "Ошибка: Go не найден в PATH." >&2
  exit 1
fi

case "$(uname -m)" in
  x86_64|amd64) ;;
  *)
    printf '%s\n' "Ошибка: сборка поддерживается только на x86_64." >&2
    exit 1
    ;;
esac

rm -rf "$WORK"
mkdir -p "$PAYLOAD_ROOT/app/licenses" "$DIST"

(
  cd "$ROOT"
  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build \
      -trimpath \
      -ldflags="-s -w -buildid= -X main.version=$APP_VERSION" \
      -o "$PAYLOAD_ROOT/app/sd-on-radio" \
      ./cmd/sd-on-radio
)

cp "$ROOT/assets/icon.svg" "$PAYLOAD_ROOT/app/icon.svg"
cp "$ROOT/LICENSE" "$PAYLOAD_ROOT/app/LICENSE"
cp "$ROOT/THIRD_PARTY_NOTICES.md" "$PAYLOAD_ROOT/app/THIRD_PARTY_NOTICES.md"

(
  cd "$ROOT"
  go list -m -f '{{if .Dir}}{{.Path}}|{{.Dir}}{{end}}' all
) | while IFS='|' read -r module module_dir; do
  [ -n "$module" ] || continue
  [ "$module" != "github.com/Serge-Nook/sd-on-radio" ] || continue
  license_path=""
  for candidate in LICENSE LICENSE.md LICENSE.txt COPYING COPYING.md; do
    if [ -f "$module_dir/$candidate" ]; then
      license_path="$module_dir/$candidate"
      break
    fi
  done
  [ -n "$license_path" ] || continue
  license_name="$(printf '%s' "$module" | sed 's#[/.]#_#g')"
  cp "$license_path" "$PAYLOAD_ROOT/app/licenses/$license_name.txt"
done

chmod 0755 "$PAYLOAD_ROOT/app/sd-on-radio"
find "$PAYLOAD_ROOT/app" -type f ! -name sd-on-radio -exec chmod 0644 {} +

tar \
  --sort=name \
  --mtime='UTC 2024-01-01' \
  --owner=0 \
  --group=0 \
  --numeric-owner \
  -czf "$PAYLOAD" \
  -C "$PAYLOAD_ROOT" \
  app

PAYLOAD_SHA256="$(sha256sum "$PAYLOAD" | awk '{ print $1 }')"
sed \
  -e "s/__APP_VERSION__/$APP_VERSION/g" \
  -e "s/__PAYLOAD_SHA256__/$PAYLOAD_SHA256/g" \
  "$ROOT/scripts/installer-header.sh" > "$OUTPUT"
cat "$PAYLOAD" >> "$OUTPUT"
chmod 0755 "$OUTPUT"

printf '%s\n' "Готово: $OUTPUT"
printf '%s\n' "SHA-256: $(sha256sum "$OUTPUT" | awk '{ print $1 }')"
