// Package ui implements the SD-ON RADIO graphical interface.
package ui

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Serge-Nook/sd-on-radio/assets"
	"github.com/Serge-Nook/sd-on-radio/internal/gamepad"
	"github.com/Serge-Nook/sd-on-radio/internal/i18n"
	"github.com/Serge-Nook/sd-on-radio/internal/player"
	"github.com/Serge-Nook/sd-on-radio/internal/settings"
	"github.com/Serge-Nook/sd-on-radio/internal/stations"
)

const (
	appID      = "ru.sd-on.radio"
	appName    = "SD-ON RADIO"
	appVersion = "2.1.0"
	websiteURL = "https://sd-on.ru/"
)

type stationItem struct {
	settings.Station
	Custom      bool
	CustomIndex int
}

type keyboardShortcut struct {
	name string
	key  fyne.KeyName
	mod  fyne.KeyModifier
}

func (s keyboardShortcut) ShortcutName() string { return s.name }
func (s keyboardShortcut) Key() fyne.KeyName    { return s.key }
func (s keyboardShortcut) Mod() fyne.KeyModifier {
	return s.mod
}

// Application owns the UI, playback engine, and persisted state.
type Application struct {
	fyneApp fyne.App
	window  fyne.Window

	settingsPath string
	settings     settings.Settings
	player       *player.Player
	controller   *gamepad.Reader

	stationList    *widget.List
	allStations    []stationItem
	selected       int
	suppressSelect bool

	stationLabel *widget.Label
	statusLabel  *widget.Label
	volumeLabel  *widget.Label
	volume       *widget.Slider
	playButton   *widget.Button
	pauseButton  *widget.Button
	stopButton   *widget.Button

	actionCh chan gamepad.Action
	closeMu  sync.Once
}

// New creates a ready-to-run application.
func New(settingsPath string) (*Application, error) {
	prefs, err := settings.Read(settingsPath)
	if err != nil {
		return nil, err
	}
	ffmpegPath, err := player.FindFFmpeg()
	if err != nil {
		return nil, err
	}

	fyneApp := app.NewWithID(appID)
	fyneApp.Settings().SetTheme(newDeckTheme())
	window := fyneApp.NewWindow(fmt.Sprintf("%s %s", appName, appVersion))
	window.SetIcon(fyne.NewStaticResource("icon.svg", assets.IconSVG))
	window.Resize(fyne.NewSize(1280, 800))
	window.SetPadded(false)
	window.CenterOnScreen()

	a := &Application{
		fyneApp:      fyneApp,
		window:       window,
		settingsPath: settingsPath,
		settings:     prefs,
		selected:     -1,
		actionCh:     make(chan gamepad.Action, 32),
	}

	a.player, err = player.New(ffmpegPath, prefs.Volume, a.onPlayerStatus)
	if err != nil {
		return nil, err
	}

	a.build()
	a.window.Canvas().AddShortcut(
		keyboardShortcut{name: "Quit SD-ON RADIO", key: fyne.KeyQ, mod: fyne.KeyModifierControl},
		func(fyne.Shortcut) {
			a.Quit()
		},
	)
	a.controller = gamepad.Start(func(action gamepad.Action) {
		select {
		case a.actionCh <- action:
		default:
		}
		fyne.Do(a.drainGamepadActions)
	})
	return a, nil
}

// Run displays the window and blocks until it is closed.
func (a *Application) Run() {
	a.window.ShowAndRun()
	a.close()
}

// Quit closes playback and exits the application event loop.
func (a *Application) Quit() {
	fyne.Do(func() {
		a.close()
		a.fyneApp.Quit()
	})
}

func (a *Application) t(key string) string {
	return i18n.T(a.settings.Language, key)
}

func (a *Application) tf(key string, args ...any) string {
	return fmt.Sprintf(a.t(key), args...)
}

func (a *Application) close() {
	a.closeMu.Do(func() {
		if a.controller != nil {
			a.controller.Stop()
		}
		if a.player != nil {
			a.player.Close()
		}
	})
}

func (a *Application) build() {
	a.rebuildStations()

	title := canvas.NewText(appName, theme.ForegroundColor())
	title.TextSize = 28
	title.TextStyle = fyne.TextStyle{Bold: true}
	version := widget.NewLabel(a.tf("version", appVersion))

	addButton := widget.NewButtonWithIcon(a.t("add_station"), theme.ContentAddIcon(), a.showAddStation)
	addButton.Importance = widget.HighImportance
	settingsButton := widget.NewButtonWithIcon(a.t("settings"), theme.SettingsIcon(), a.showSettings)
	aboutButton := widget.NewButtonWithIcon(a.t("about"), theme.InfoIcon(), a.showAbout)

	headerLeft := container.NewVBox(title, version)
	header := container.NewBorder(nil, nil, headerLeft, container.NewHBox(addButton, settingsButton, aboutButton))

	a.stationList = widget.NewList(
		func() int { return len(a.allStations) },
		func() fyne.CanvasObject {
			name := widget.NewLabel(a.t("station"))
			name.TextStyle = fyne.TextStyle{Bold: true}
			badge := widget.NewLabel("")
			edit := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
			remove := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			actions := container.NewHBox(edit, remove)
			return container.NewBorder(nil, nil, nil, actions, container.NewVBox(name, badge))
		},
		func(id widget.ListItemID, object fyne.CanvasObject) {
			item := a.allStations[id]
			border := object.(*fyne.Container)
			labels := border.Objects[0].(*fyne.Container)
			name := labels.Objects[0].(*widget.Label)
			badge := labels.Objects[1].(*widget.Label)
			actions := border.Objects[1].(*fyne.Container)
			edit := actions.Objects[0].(*widget.Button)
			remove := actions.Objects[1].(*widget.Button)

			name.SetText(item.Name)
			if item.Custom {
				badge.SetText(a.t("custom_station"))
				edit.Show()
				remove.Show()
				customIndex := item.CustomIndex
				edit.OnTapped = func() { a.showEditStation(customIndex) }
				remove.OnTapped = func() { a.confirmDeleteStation(customIndex) }
			} else {
				badge.SetText(a.t("builtin_station"))
				edit.Hide()
				remove.Hide()
				edit.OnTapped = nil
				remove.OnTapped = nil
			}
			edit.Refresh()
			remove.Refresh()
		},
	)
	a.stationList.OnSelected = func(id widget.ListItemID) {
		a.selected = id
		if a.suppressSelect {
			return
		}
		a.playSelected()
	}
	if a.selected >= 0 && a.selected < len(a.allStations) {
		a.suppressSelect = true
		a.stationList.Select(a.selected)
		a.suppressSelect = false
	}

	a.stationLabel = widget.NewLabel(a.t("no_station"))
	a.stationLabel.TextStyle = fyne.TextStyle{Bold: true}
	a.statusLabel = widget.NewLabel(a.t("stopped"))

	a.playButton = widget.NewButtonWithIcon(a.t("play"), theme.MediaPlayIcon(), a.playSelected)
	a.playButton.Importance = widget.HighImportance
	a.pauseButton = widget.NewButtonWithIcon(a.t("pause"), theme.MediaPauseIcon(), a.player.Pause)
	a.stopButton = widget.NewButtonWithIcon(a.t("stop"), theme.MediaStopIcon(), a.player.Stop)

	a.volumeLabel = widget.NewLabel("")
	a.volume = widget.NewSlider(0, 1)
	a.volume.Step = 0.01
	a.volume.Value = a.settings.Volume
	a.volume.OnChanged = func(value float64) {
		a.player.SetVolume(value)
		a.settings.Volume = value
		a.volumeLabel.SetText(a.tf("volume", value*100))
	}
	a.volume.OnChangeEnded = func(value float64) { a.persist() }
	a.volume.OnChanged(a.settings.Volume)

	controls := container.NewHBox(a.playButton, a.pauseButton, a.stopButton)
	volumeBox := container.NewBorder(nil, nil, a.volumeLabel, nil, a.volume)
	playerPanel := container.NewVBox(
		container.NewHBox(a.stationLabel, layout.NewSpacer(), a.statusLabel),
		controls,
		volumeBox,
	)

	content := container.NewBorder(
		container.NewPadded(header),
		container.NewPadded(playerPanel),
		nil,
		nil,
		container.NewPadded(a.stationList),
	)
	a.window.SetContent(content)
	if a.player != nil {
		a.applyPlayerStatus(a.player.Status())
	}
}

func (a *Application) rebuildStations() {
	var builtIns []settings.Station
	if !a.settings.HideBuiltIn {
		builtIns = stations.BuiltIn()
	}
	a.allStations = make([]stationItem, 0, len(builtIns)+len(a.settings.Stations))
	for _, station := range builtIns {
		a.allStations = append(a.allStations, stationItem{Station: station})
	}
	for index, station := range a.settings.Stations {
		a.allStations = append(a.allStations, stationItem{
			Station:     station,
			Custom:      true,
			CustomIndex: index,
		})
	}

	a.selected = -1
	for index, station := range a.allStations {
		if station.Name == a.settings.LastStation {
			a.selected = index
			break
		}
	}
}

func (a *Application) refreshStationList() {
	currentName := ""
	if a.selected >= 0 && a.selected < len(a.allStations) {
		currentName = a.allStations[a.selected].Name
	}
	a.rebuildStations()
	if currentName != "" {
		for index, station := range a.allStations {
			if station.Name == currentName {
				a.selected = index
				break
			}
		}
	}
	a.stationList.UnselectAll()
	a.stationList.Refresh()
	if a.selected >= 0 && a.selected < len(a.allStations) {
		a.suppressSelect = true
		a.stationList.Select(a.selected)
		a.suppressSelect = false
	}
}

func (a *Application) playSelected() {
	if a.selected < 0 || a.selected >= len(a.allStations) {
		if len(a.allStations) == 0 {
			return
		}
		a.selected = 0
		a.suppressSelect = true
		a.stationList.Select(0)
		a.suppressSelect = false
	}
	station := a.allStations[a.selected]
	a.settings.LastStation = station.Name
	a.stationLabel.SetText(a.tf("now_playing", station.Name))
	a.persist()
	a.player.Play(station.Name, station.URL)
}

func (a *Application) persist() {
	stored, err := settings.Write(a.settingsPath, a.settings)
	if err != nil {
		dialog.ShowError(fmt.Errorf(a.t("save_error"), err), a.window)
		return
	}
	a.settings = stored
}

func (a *Application) onPlayerStatus(status player.Status) {
	fyne.Do(func() { a.applyPlayerStatus(status) })
}

func (a *Application) applyPlayerStatus(status player.Status) {
	switch status.State {
	case player.StateConnecting:
		a.statusLabel.SetText(a.t("connecting"))
		a.statusLabel.Importance = widget.MediumImportance
		a.stationLabel.SetText(a.tf("now_playing", status.Station))
	case player.StatePlaying:
		a.statusLabel.SetText(a.t("playing"))
		a.statusLabel.Importance = widget.SuccessImportance
		a.stationLabel.SetText(a.tf("now_playing", status.Station))
	case player.StatePaused:
		a.statusLabel.SetText(a.t("paused"))
		a.statusLabel.Importance = widget.WarningImportance
		a.stationLabel.SetText(a.tf("now_playing", status.Station))
	case player.StateStopped:
		a.statusLabel.SetText(a.t("stopped"))
		a.statusLabel.Importance = widget.MediumImportance
		a.stationLabel.SetText(a.t("no_station"))
	case player.StateError:
		message := a.t("stream_error_retry")
		if status.Err != nil {
			message = a.tf("error", status.Err.Error())
		}
		a.statusLabel.SetText(message)
		a.statusLabel.Importance = widget.DangerImportance
	}
	a.statusLabel.Refresh()
}

func (a *Application) showAddStation() {
	a.showStationForm(a.t("add_station"), settings.Station{}, func(station settings.Station) {
		a.settings.Stations = append(a.settings.Stations, station)
		a.persist()
		a.refreshStationList()
	})
}

func (a *Application) showEditStation(index int) {
	if index < 0 || index >= len(a.settings.Stations) {
		return
	}
	a.showStationForm(a.t("edit_station"), a.settings.Stations[index], func(station settings.Station) {
		oldName := a.settings.Stations[index].Name
		a.settings.Stations[index] = station
		if a.settings.LastStation == oldName {
			a.settings.LastStation = station.Name
		}
		a.persist()
		a.refreshStationList()
	})
}

func (a *Application) showStationForm(title string, initial settings.Station, save func(settings.Station)) {
	name := widget.NewEntry()
	name.SetPlaceHolder(a.t("station_name"))
	name.SetText(initial.Name)
	name.Validator = func(value string) error {
		if _, ok := settings.NormalizeStation(settings.Station{
			Name: value,
			URL:  "https://example.com/stream",
		}); !ok {
			return errors.New(a.t("name_required"))
		}
		return nil
	}
	streamURL := widget.NewEntry()
	streamURL.SetPlaceHolder("https://example.com/stream.mp3")
	streamURL.SetText(initial.URL)
	streamURL.Validator = func(value string) error {
		if !settings.IsValidHTTPURL(strings.TrimSpace(value)) || len(strings.TrimSpace(value)) > 2048 {
			return errors.New(a.t("url_invalid"))
		}
		return nil
	}

	items := []*widget.FormItem{
		widget.NewFormItem(a.t("name"), name),
		widget.NewFormItem(a.t("stream_url"), streamURL),
	}
	form := dialog.NewForm(title, a.t("save"), a.t("cancel"), items, func(ok bool) {
		if !ok {
			return
		}
		station, valid := settings.NormalizeStation(settings.Station{Name: name.Text, URL: streamURL.Text})
		if !valid {
			dialog.ShowError(errors.New(a.t("station_invalid")), a.window)
			return
		}
		save(station)
	}, a.window)
	form.Resize(fyne.NewSize(720, 280))
	form.Show()
}

func (a *Application) confirmDeleteStation(index int) {
	if index < 0 || index >= len(a.settings.Stations) {
		return
	}
	station := a.settings.Stations[index]
	dialog.ShowConfirm(
		a.t("delete_station"),
		a.tf("delete_confirm", station.Name),
		func(ok bool) {
			if !ok {
				return
			}
			if a.settings.LastStation == station.Name {
				a.settings.LastStation = ""
				a.player.Stop()
			}
			a.settings.Stations = append(a.settings.Stations[:index], a.settings.Stations[index+1:]...)
			a.persist()
			a.refreshStationList()
		},
		a.window,
	)
}

func (a *Application) showAbout() {
	name := canvas.NewText(appName, theme.ForegroundColor())
	name.TextSize = 26
	name.TextStyle = fyne.TextStyle{Bold: true}
	info := widget.NewLabel(strings.Join([]string{
		a.tf("version", appVersion),
		a.t("about_tagline"),
		a.t("about_author"),
		"",
		a.t("about_safe"),
	}, "\n"))
	info.Alignment = fyne.TextAlignCenter
	openWebsite := widget.NewButton(a.t("open_website"), func() {
		parsed, _ := url.Parse(websiteURL)
		_ = a.fyneApp.OpenURL(parsed)
	})
	content := container.NewVBox(
		container.NewCenter(name),
		info,
		container.NewCenter(openWebsite),
	)
	dlg := dialog.NewCustom(a.t("about"), a.t("close"), content, a.window)
	dlg.Resize(fyne.NewSize(560, 360))
	dlg.Show()
}

func (a *Application) showSettings() {
	names := make([]string, 0, len(i18n.Languages()))
	for _, lang := range i18n.Languages() {
		names = append(names, i18n.NativeName(lang))
	}
	language := widget.NewSelect(names, nil)
	language.SetSelected(i18n.NativeName(a.settings.Language))

	gamepadCheck := widget.NewCheck(a.t("enabled"), nil)
	gamepadCheck.SetChecked(a.settings.GamepadEnabled)
	gamepadHint := widget.NewLabel(a.t("gamepad_hotkeys_hint"))
	gamepadHint.Wrapping = fyne.TextWrapWord

	hideCheck := widget.NewCheck(a.t("hidden"), nil)
	hideCheck.SetChecked(a.settings.HideBuiltIn)
	hideHint := widget.NewLabel(a.t("hide_builtin_hint"))
	hideHint.Wrapping = fyne.TextWrapWord

	items := []*widget.FormItem{
		widget.NewFormItem(a.t("language"), language),
		widget.NewFormItem(a.t("gamepad_hotkeys"), container.NewVBox(gamepadCheck, gamepadHint)),
		widget.NewFormItem(a.t("hide_builtin"), container.NewVBox(hideCheck, hideHint)),
	}
	form := dialog.NewForm(a.t("settings"), a.t("save"), a.t("cancel"), items, func(ok bool) {
		if !ok {
			return
		}
		if lang, valid := i18n.FromNativeName(language.Selected); valid {
			a.settings.Language = lang
		}
		a.settings.GamepadEnabled = gamepadCheck.Checked
		a.settings.HideBuiltIn = hideCheck.Checked
		a.persist()
		a.build()
	}, a.window)
	form.Resize(fyne.NewSize(720, 400))
	form.Show()
}

func (a *Application) drainGamepadActions() {
	for {
		select {
		case action := <-a.actionCh:
			a.handleGamepadAction(action)
		default:
			return
		}
	}
}

func (a *Application) handleGamepadAction(action gamepad.Action) {
	if !a.settings.GamepadEnabled || len(a.allStations) == 0 {
		return
	}
	switch action {
	case gamepad.ActionUp:
		if a.selected <= 0 {
			a.selected = len(a.allStations) - 1
		} else {
			a.selected--
		}
		a.suppressSelect = true
		a.stationList.Select(a.selected)
		a.suppressSelect = false
		a.stationList.ScrollTo(a.selected)
	case gamepad.ActionDown:
		if a.selected < 0 || a.selected >= len(a.allStations)-1 {
			a.selected = 0
		} else {
			a.selected++
		}
		a.suppressSelect = true
		a.stationList.Select(a.selected)
		a.suppressSelect = false
		a.stationList.ScrollTo(a.selected)
	case gamepad.ActionSelect:
		a.playSelected()
	case gamepad.ActionBack:
		a.player.Stop()
	case gamepad.ActionPlayPause:
		if a.player.State() == player.StatePlaying {
			a.player.Pause()
		} else {
			a.playSelected()
		}
	case gamepad.ActionVolumeUp:
		a.volume.SetValue(min(1, a.volume.Value+0.05))
		a.persist()
	case gamepad.ActionVolumeDown:
		a.volume.SetValue(max(0, a.volume.Value-0.05))
		a.persist()
	}
}
