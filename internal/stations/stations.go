// Package stations provides the built-in radio station list.
package stations

import "github.com/Serge-Nook/sd-on-radio/internal/settings"

// BuiltIn returns the hard-coded stations shipped with the application.
func BuiltIn() []settings.Station {
	return []settings.Station{
		{Name: "Hard Rock Heaven", URL: "http://hydra.cdnstream.com:80/1521_128"},
		{Name: "НАШЕ Радио", URL: "https://nashe1.hostingradio.ru/nashe-128.mp3"},
		{Name: "Европа Плюс", URL: "http://ep256.hostingradio.ru:8052/europaplus256.mp3"},
		{Name: "Relax FM", URL: "http://23.105.238.4/gpm-relaxfm495.aacp"},
		{Name: "ENERGY FM", URL: "http://23.105.238.4/gpm-energyfm495.aacp"},
		{Name: "Радио Maximum", URL: "http://23.105.238.4/maximum96.aacp"},
		{Name: "ULTRA", URL: "https://nashe1.hostingradio.ru/ultra-128.mp3"},
	}
}
