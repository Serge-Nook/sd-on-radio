# SD-ON RADIO

Автономное приложение для прослушивания интернет-радио на Steam Deck. Оно не изменяет Steam, не требует Decky Loader, плагинов или прав `sudo`.

## Возможности

- восемь встроенных радиостанций;
- добавление, редактирование и удаление собственных потоков;
- Play, Pause, Stop и отдельная громкость приложения;
- автоматическое переподключение с экспоненциальной задержкой;
- сохранение станций, громкости и последнего выбора в `~/.local/share/sd-on-radio/settings.json`;
- крупный русскоязычный интерфейс для экрана 1280×800, сенсора и геймпада;
- тёмная тема и автономный Electron runtime.

## Установка на Steam Deck

Скачайте `sd-on-radio-installer.sh` из раздела Releases, затем в Konsole выполните:

```sh
chmod +x sd-on-radio-installer.sh
./sd-on-radio-installer.sh
```

Приложение появится в меню рабочего стола. Его также можно запустить командой:

```sh
sd-on-radio
```

Установщик работает от обычного пользователя и размещает файлы только в домашней директории.

## Удаление

```sh
./sd-on-radio-installer.sh --uninstall
```

Удаление также стирает пользовательские станции и настройки.

## Разработка

Требуются Node.js 20, npm и стандартная утилита `tar`.

```sh
npm ci
npm start
```

Проверки:

```sh
npm run lint
npm test
```

Создание единого самораспаковывающегося установщика для Steam Deck (Linux x86_64):

```sh
npm run build:installer
```

Результат: `dist/sd-on-radio-installer.sh`. В него включены приложение, Electron и необходимые ресурсы; на Steam Deck Node.js и Electron не требуются.

## Данные

Настройки хранятся в:

```text
~/.local/share/sd-on-radio/settings.json
```

Пример:

```json
{
  "stations": [
    {
      "name": "Моя станция",
      "url": "https://example.com/stream.mp3"
    }
  ],
  "last_station": "Моя станция",
  "volume": 0.8
}
```

## Автор

Serge Nook — [SD-ON.RU](https://sd-on.ru/)
