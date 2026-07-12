window.SdOnRadio = window.SdOnRadio || {};

window.SdOnRadio.BUILT_IN_STATIONS = Object.freeze([
  { name: "Hard Rock Heaven", url: "http://hydra.cdnstream.com:80/1521_128" },
  { name: "НАШЕ Радио", url: "https://nashe1.hostingradio.ru/nashe-128.mp3" },
  { name: "Европа Плюс", url: "http://ep256.hostingradio.ru:8052/europaplus256.mp3" },
  { name: "Relax FM", url: "http://23.105.238.4/gpm-relaxfm495.aacp" },
  { name: "Love Radio", url: "https://stream2.n340.com/12_love_64_reg_44?type=aac" },
  { name: "ENERGY FM", url: "http://23.105.238.4/gpm-energyfm495.aacp" },
  { name: "Радио Maximum", url: "http://23.105.238.4/maximum96.aacp" },
  { name: "ULTRA", url: "https://nashe1.hostingradio.ru/ultra-128.mp3" },
]);

window.SdOnRadio.isValidStreamUrl = (value) => {
  try {
    const url = new URL(value);
    return (url.protocol === "http:" || url.protocol === "https:") && Boolean(url.hostname);
  } catch {
    return false;
  }
};
