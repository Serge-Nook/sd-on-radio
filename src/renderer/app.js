const {
  BUILT_IN_STATIONS,
  RadioPlayer,
  isValidStreamUrl,
} = window.SdOnRadio;

const elements = {
  aboutButton: document.querySelector("#about-button"),
  aboutDialog: document.querySelector("#about-dialog"),
  aboutVersion: document.querySelector("#about-version"),
  addStation: document.querySelector("#add-station"),
  appVersion: document.querySelector("#app-version"),
  audio: document.querySelector("#radio-audio"),
  currentStation: document.querySelector("#current-station"),
  pauseButton: document.querySelector("#pause-button"),
  playButton: document.querySelector("#play-button"),
  playerStatus: document.querySelector("#player-status"),
  stationCount: document.querySelector("#station-count"),
  stationDialog: document.querySelector("#station-dialog"),
  stationDialogTitle: document.querySelector("#station-dialog-title"),
  stationForm: document.querySelector("#station-form"),
  stationFormError: document.querySelector("#station-form-error"),
  stationIndex: document.querySelector("#station-index"),
  stationList: document.querySelector("#station-list"),
  stationName: document.querySelector("#station-name"),
  stationUrl: document.querySelector("#station-url"),
  statusDot: document.querySelector("#status-dot"),
  stopButton: document.querySelector("#stop-button"),
  toast: document.querySelector("#toast"),
  volume: document.querySelector("#volume"),
  volumeValue: document.querySelector("#volume-value"),
  websiteButton: document.querySelector("#website-button"),
};

let settings = {
  stations: [],
  last_station: "",
  volume: 0.8,
};
let selectedStation = null;
let volumeSaveTimer;
let toastTimer;

const player = new RadioPlayer(elements.audio, updatePlayerState);

function stationInitials(name) {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0].toUpperCase())
    .join("");
}

function createStationCard(station, type, index) {
  const card = document.createElement("article");
  card.className = "station-card";
  card.dataset.stationType = type;
  card.dataset.stationIndex = String(index);
  card.setAttribute("role", "listitem");

  const playButton = document.createElement("button");
  playButton.className = "station-select";
  playButton.type = "button";
  playButton.dataset.focusable = "true";
  playButton.innerHTML = `
    <span class="station-avatar">${stationInitials(station.name)}</span>
    <span class="station-copy">
      <strong></strong>
      <small>${type === "built-in" ? "Встроенная станция" : "Пользовательская станция"}</small>
    </span>
    <span class="station-play" aria-hidden="true">▶</span>
  `;
  playButton.querySelector("strong").textContent = station.name;
  playButton.addEventListener("click", () => selectStation(station, card));
  card.append(playButton);

  if (type === "custom") {
    const actions = document.createElement("div");
    actions.className = "station-actions";

    const editButton = document.createElement("button");
    editButton.className = "icon-button small-icon-button";
    editButton.type = "button";
    editButton.dataset.focusable = "true";
    editButton.setAttribute("aria-label", `Редактировать ${station.name}`);
    editButton.textContent = "✎";
    editButton.addEventListener("click", () => openStationDialog(index));

    const deleteButton = document.createElement("button");
    deleteButton.className = "icon-button small-icon-button danger-button";
    deleteButton.type = "button";
    deleteButton.dataset.focusable = "true";
    deleteButton.setAttribute("aria-label", `Удалить ${station.name}`);
    deleteButton.textContent = "×";
    deleteButton.addEventListener("click", () => deleteStation(index));

    actions.append(editButton, deleteButton);
    card.append(actions);
  }

  return card;
}

function renderStations() {
  elements.stationList.replaceChildren();

  BUILT_IN_STATIONS.forEach((station, index) => {
    elements.stationList.append(createStationCard(station, "built-in", index));
  });
  settings.stations.forEach((station, index) => {
    elements.stationList.append(createStationCard(station, "custom", index));
  });

  elements.stationCount.textContent = `${BUILT_IN_STATIONS.length + settings.stations.length} станций`;
  markSelectedStation();
}

function markSelectedStation(selectedCard) {
  document.querySelectorAll(".station-card").forEach((card) => {
    const cardName = card.querySelector(".station-select strong").textContent;
    card.classList.toggle(
      "is-active",
      selectedCard ? card === selectedCard : cardName === selectedStation?.name,
    );
  });
}

function selectStation(station, card) {
  selectedStation = station;
  settings.last_station = station.name;
  elements.currentStation.textContent = station.name;
  markSelectedStation(card);
  player.select(station);
  saveSettings();
}

function findLastStation() {
  return [...BUILT_IN_STATIONS, ...settings.stations]
    .find((station) => station.name === settings.last_station);
}

function updatePlayerState(event) {
  const stateCopy = {
    idle: "Выберите станцию из списка",
    empty: "Сначала выберите станцию",
    connecting: "Подключение к потоку…",
    buffering: "Буферизация…",
    playing: "Воспроизведение",
    paused: "Пауза",
    stopped: "Остановлено",
    reconnecting: `Нет соединения. Повтор через ${Math.ceil((event.delay || 0) / 1000)} сек.`,
  };

  elements.playerStatus.textContent = stateCopy[event.state] || "";
  elements.statusDot.dataset.state = event.state;

  if (event.state === "empty") {
    showToast("Выберите радиостанцию");
  }
}

function openStationDialog(index = -1) {
  elements.stationForm.reset();
  elements.stationFormError.textContent = "";
  elements.stationIndex.value = String(index);

  if (index >= 0) {
    const station = settings.stations[index];
    elements.stationDialogTitle.textContent = "Редактировать станцию";
    elements.stationName.value = station.name;
    elements.stationUrl.value = station.url;
  } else {
    elements.stationDialogTitle.textContent = "Новая станция";
  }

  elements.stationDialog.showModal();
  window.setTimeout(() => elements.stationName.focus(), 0);
}

function closeDialog(dialog) {
  if (dialog.open) {
    dialog.close();
  }
}

function handleStationSubmit(event) {
  event.preventDefault();

  const name = elements.stationName.value.trim();
  const url = elements.stationUrl.value.trim();
  const index = Number(elements.stationIndex.value);

  if (!name) {
    elements.stationFormError.textContent = "Введите название станции.";
    elements.stationName.focus();
    return;
  }
  if (!isValidStreamUrl(url)) {
    elements.stationFormError.textContent = "URL должен начинаться с http:// или https://.";
    elements.stationUrl.focus();
    return;
  }

  const station = { name, url };
  if (index >= 0) {
    const oldStation = settings.stations[index];
    settings.stations[index] = station;
    if (selectedStation === oldStation) {
      selectedStation = station;
      settings.last_station = station.name;
      elements.currentStation.textContent = station.name;
    }
    showToast("Станция обновлена");
  } else {
    settings.stations.push(station);
    showToast("Станция добавлена");
  }

  renderStations();
  saveSettings();
  closeDialog(elements.stationDialog);
}

function deleteStation(index) {
  const station = settings.stations[index];
  const confirmed = window.confirm(`Удалить станцию «${station.name}»?`);
  if (!confirmed) {
    return;
  }

  settings.stations.splice(index, 1);
  if (selectedStation === station) {
    player.stop();
    selectedStation = null;
    settings.last_station = "";
    elements.currentStation.textContent = "Станция не выбрана";
  }

  renderStations();
  saveSettings();
  showToast("Станция удалена");
}

async function saveSettings() {
  try {
    await window.radioAPI.saveSettings(settings);
  } catch {
    showToast("Не удалось сохранить настройки");
  }
}

function updateVolume() {
  const percentage = Number(elements.volume.value);
  settings.volume = percentage / 100;
  elements.volumeValue.textContent = `${percentage}%`;
  player.setVolume(settings.volume);

  window.clearTimeout(volumeSaveTimer);
  volumeSaveTimer = window.setTimeout(saveSettings, 250);
}

function showToast(message) {
  window.clearTimeout(toastTimer);
  elements.toast.textContent = message;
  elements.toast.classList.add("is-visible");
  toastTimer = window.setTimeout(() => elements.toast.classList.remove("is-visible"), 2400);
}

function visibleFocusableElements() {
  return [...document.querySelectorAll(
    "button:not([disabled]), input:not([type='hidden']):not([disabled])",
  )].filter((element) => element.offsetParent !== null);
}

function moveFocus(direction) {
  const focusable = visibleFocusableElements();
  if (!focusable.length) {
    return;
  }

  const current = document.activeElement;
  if (!focusable.includes(current)) {
    focusable[0].focus();
    return;
  }

  const currentRect = current.getBoundingClientRect();
  const originX = currentRect.left + currentRect.width / 2;
  const originY = currentRect.top + currentRect.height / 2;
  const candidates = focusable
    .filter((element) => element !== current)
    .map((element) => {
      const rect = element.getBoundingClientRect();
      const deltaX = rect.left + rect.width / 2 - originX;
      const deltaY = rect.top + rect.height / 2 - originY;
      const valid = {
        up: deltaY < -4,
        down: deltaY > 4,
        left: deltaX < -4,
        right: deltaX > 4,
      }[direction];
      const primary = direction === "up" || direction === "down" ? Math.abs(deltaY) : Math.abs(deltaX);
      const secondary = direction === "up" || direction === "down" ? Math.abs(deltaX) : Math.abs(deltaY);
      return { element, score: primary + secondary * 2, valid };
    })
    .filter((candidate) => candidate.valid)
    .sort((left, right) => left.score - right.score);

  candidates[0]?.element.focus();
}

function enableGamepadNavigation() {
  let previousButtons = [];

  function frame() {
    const gamepad = [...navigator.getGamepads()].find(Boolean);
    if (gamepad) {
      const pressed = gamepad.buttons.map((button) => button.pressed);
      const justPressed = (index) => pressed[index] && !previousButtons[index];

      if (justPressed(12)) moveFocus("up");
      if (justPressed(13)) moveFocus("down");
      if (justPressed(14)) moveFocus("left");
      if (justPressed(15)) moveFocus("right");
      if (justPressed(0) && document.activeElement instanceof HTMLElement) {
        document.activeElement.click();
      }
      if (justPressed(1)) {
        const dialog = document.querySelector("dialog[open]");
        if (dialog) closeDialog(dialog);
      }

      previousButtons = pressed;
    } else {
      previousButtons = [];
    }

    window.requestAnimationFrame(frame);
  }

  window.requestAnimationFrame(frame);
}

async function initialize() {
  const [savedSettings, appInfo] = await Promise.all([
    window.radioAPI.loadSettings(),
    window.radioAPI.getAppInfo(),
  ]);

  settings = savedSettings;
  elements.appVersion.textContent = `v${appInfo.version}`;
  elements.aboutVersion.textContent = appInfo.version;
  elements.volume.value = String(Math.round(settings.volume * 100));
  elements.volumeValue.textContent = `${elements.volume.value}%`;
  player.setVolume(settings.volume);

  selectedStation = findLastStation() || null;
  if (selectedStation) {
    elements.currentStation.textContent = selectedStation.name;
    elements.playerStatus.textContent = "Готово к воспроизведению";
  }

  renderStations();
  enableGamepadNavigation();
}

elements.addStation.addEventListener("click", () => openStationDialog());
elements.aboutButton.addEventListener("click", () => elements.aboutDialog.showModal());
elements.stationForm.addEventListener("submit", handleStationSubmit);
elements.playButton.addEventListener("click", () => player.play());
elements.pauseButton.addEventListener("click", () => player.pause());
elements.stopButton.addEventListener("click", () => player.stop());
elements.volume.addEventListener("input", updateVolume);
elements.websiteButton.addEventListener("click", () => window.radioAPI.openWebsite());

document.querySelectorAll("[data-close-dialog]").forEach((button) => {
  button.addEventListener("click", () => closeDialog(button.closest("dialog")));
});

document.addEventListener("keydown", (event) => {
  const editing = event.target instanceof HTMLInputElement;
  if (event.code === "Space" && !editing && !document.querySelector("dialog[open]")) {
    event.preventDefault();
    if (elements.audio.paused) {
      player.play();
    } else {
      player.pause();
    }
  }
});

initialize().catch(() => {
  elements.playerStatus.textContent = "Ошибка загрузки настроек";
  showToast("Не удалось запустить приложение");
});
