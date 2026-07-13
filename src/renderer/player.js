window.SdOnRadio = window.SdOnRadio || {};

window.SdOnRadio.RadioPlayer = class RadioPlayer {
  constructor(audio, onStateChange) {
    this.audio = audio;
    this.onStateChange = onStateChange;
    this.station = null;
    this.reconnectAttempt = 0;
    this.reconnectTimer = null;
    this.manuallyPaused = false;
    this.stopped = true;

    this.audio.addEventListener("playing", () => {
      this.reconnectAttempt = 0;
      this.emit("playing");
    });
    this.audio.addEventListener("waiting", () => this.emit("buffering"));
    this.audio.addEventListener("stalled", () => this.scheduleReconnect());
    this.audio.addEventListener("error", () => this.scheduleReconnect());
    this.audio.addEventListener("ended", () => this.scheduleReconnect());
    this.audio.addEventListener("pause", () => {
      if (this.manuallyPaused) {
        this.emit("paused");
      }
    });
  }

  setVolume(value) {
    this.audio.volume = Math.min(1, Math.max(0, value));
  }

  select(station) {
    this.clearReconnect();
    this.station = station;
    this.audio.src = station.url;
    this.manuallyPaused = false;
    this.stopped = false;
    this.reconnectAttempt = 0;
    this.emit("connecting");
    this.tryPlay();
  }

  play() {
    if (!this.station) {
      this.emit("empty");
      return;
    }

    this.clearReconnect();
    this.manuallyPaused = false;
    this.stopped = false;

    if (!this.audio.src) {
      this.audio.src = this.station.url;
    }

    this.emit("connecting");
    this.tryPlay();
  }

  pause() {
    if (!this.station) {
      this.emit("empty");
      return;
    }

    this.clearReconnect();
    this.manuallyPaused = true;
    this.audio.pause();
    this.emit("paused");
  }

  stop() {
    this.clearReconnect();
    this.manuallyPaused = false;
    this.stopped = true;
    this.audio.pause();
    this.audio.removeAttribute("src");
    this.audio.load();
    this.emit("stopped");
  }

  tryPlay() {
    this.audio.play().catch((error) => {
      if (error.name !== "AbortError") {
        this.scheduleReconnect();
      }
    });
  }

  scheduleReconnect() {
    if (!this.station || this.manuallyPaused || this.stopped || this.reconnectTimer) {
      return;
    }

    const delay = Math.min(30000, 2000 * (2 ** this.reconnectAttempt));
    this.reconnectAttempt += 1;
    this.emit("reconnecting", { delay, attempt: this.reconnectAttempt });
    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectTimer = null;
      this.audio.src = this.station.url;
      this.tryPlay();
    }, delay);
  }

  clearReconnect() {
    if (this.reconnectTimer) {
      window.clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  emit(state, detail = {}) {
    this.onStateChange({
      state,
      station: this.station,
      ...detail,
    });
  }
};
