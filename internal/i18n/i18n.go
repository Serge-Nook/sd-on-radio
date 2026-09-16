// Package i18n provides the interface translations.
package i18n

// Language is a supported interface language code.
type Language string

const (
	Russian Language = "ru"
	English Language = "en"
	Chinese Language = "zh"
)

// DefaultLanguage is used when no valid language is configured.
const DefaultLanguage = Russian

// Languages lists the supported languages in display order.
func Languages() []Language {
	return []Language{Russian, English, Chinese}
}

// IsValid reports whether code is a supported language.
func IsValid(code Language) bool {
	for _, lang := range Languages() {
		if lang == code {
			return true
		}
	}
	return false
}

// NativeName returns the language name written in that language.
func NativeName(code Language) string {
	switch code {
	case English:
		return "English"
	case Chinese:
		return "中文"
	default:
		return "Русский"
	}
}

// FromNativeName resolves a native name back to its language code.
func FromNativeName(name string) (Language, bool) {
	for _, lang := range Languages() {
		if NativeName(lang) == name {
			return lang, true
		}
	}
	return "", false
}

// T returns the translation of key for lang, falling back to Russian and then
// to the key itself.
func T(lang Language, key string) string {
	if entry, ok := messages[key]; ok {
		if text, ok := entry[lang]; ok && text != "" {
			return text
		}
		if text, ok := entry[DefaultLanguage]; ok {
			return text
		}
	}
	return key
}

var messages = map[string]map[Language]string{
	"version": {
		Russian: "Версия %s",
		English: "Version %s",
		Chinese: "版本 %s",
	},
	"add_station": {
		Russian: "Добавить станцию",
		English: "Add station",
		Chinese: "添加电台",
	},
	"edit_station": {
		Russian: "Редактировать станцию",
		English: "Edit station",
		Chinese: "编辑电台",
	},
	"delete_station": {
		Russian: "Удалить станцию",
		English: "Delete station",
		Chinese: "删除电台",
	},
	"delete_confirm": {
		Russian: "Удалить «%s»?",
		English: "Delete “%s”?",
		Chinese: "删除“%s”？",
	},
	"settings": {
		Russian: "Настройки",
		English: "Settings",
		Chinese: "设置",
	},
	"about": {
		Russian: "О программе",
		English: "About",
		Chinese: "关于",
	},
	"station": {
		Russian: "Станция",
		English: "Station",
		Chinese: "电台",
	},
	"custom_station": {
		Russian: "Пользовательская станция",
		English: "Custom station",
		Chinese: "自定义电台",
	},
	"builtin_station": {
		Russian: "Встроенная станция",
		English: "Built-in station",
		Chinese: "内置电台",
	},
	"no_station": {
		Russian: "Станция не выбрана",
		English: "No station selected",
		Chinese: "未选择电台",
	},
	"now_playing": {
		Russian: "Сейчас играет: %s",
		English: "Now playing: %s",
		Chinese: "正在播放：%s",
	},
	"stopped": {
		Russian: "Остановлено",
		English: "Stopped",
		Chinese: "已停止",
	},
	"connecting": {
		Russian: "Подключение…",
		English: "Connecting…",
		Chinese: "正在连接…",
	},
	"playing": {
		Russian: "Воспроизведение",
		English: "Playing",
		Chinese: "播放中",
	},
	"paused": {
		Russian: "Пауза",
		English: "Paused",
		Chinese: "已暂停",
	},
	"stream_error_retry": {
		Russian: "Ошибка потока, повторное подключение…",
		English: "Stream error, reconnecting…",
		Chinese: "流媒体错误，正在重新连接…",
	},
	"error": {
		Russian: "Ошибка: %s",
		English: "Error: %s",
		Chinese: "错误：%s",
	},
	"play": {
		Russian: "Воспроизвести",
		English: "Play",
		Chinese: "播放",
	},
	"pause": {
		Russian: "Пауза",
		English: "Pause",
		Chinese: "暂停",
	},
	"stop": {
		Russian: "Стоп",
		English: "Stop",
		Chinese: "停止",
	},
	"volume": {
		Russian: "Громкость: %.0f%%",
		English: "Volume: %.0f%%",
		Chinese: "音量：%.0f%%",
	},
	"save_error": {
		Russian: "не удалось сохранить настройки: %w",
		English: "failed to save settings: %w",
		Chinese: "无法保存设置：%w",
	},
	"station_name": {
		Russian: "Название станции",
		English: "Station name",
		Chinese: "电台名称",
	},
	"name_required": {
		Russian: "название обязательно, не более 80 символов",
		English: "name is required, at most 80 characters",
		Chinese: "名称为必填项，不超过 80 个字符",
	},
	"url_invalid": {
		Russian: "нужен корректный URL http:// или https://",
		English: "a valid http:// or https:// URL is required",
		Chinese: "需要有效的 http:// 或 https:// 网址",
	},
	"name": {
		Russian: "Название",
		English: "Name",
		Chinese: "名称",
	},
	"stream_url": {
		Russian: "URL потока",
		English: "Stream URL",
		Chinese: "流媒体网址",
	},
	"save": {
		Russian: "Сохранить",
		English: "Save",
		Chinese: "保存",
	},
	"cancel": {
		Russian: "Отмена",
		English: "Cancel",
		Chinese: "取消",
	},
	"close": {
		Russian: "Закрыть",
		English: "Close",
		Chinese: "关闭",
	},
	"station_invalid": {
		Russian: "укажите название (до 80 символов) и корректный URL http:// или https://",
		English: "enter a name (up to 80 characters) and a valid http:// or https:// URL",
		Chinese: "请输入名称（最多 80 个字符）和有效的 http:// 或 https:// 网址",
	},
	"about_tagline": {
		Russian: "Автономное интернет-радио для Steam Deck",
		English: "Standalone internet radio for Steam Deck",
		Chinese: "Steam Deck 独立网络电台",
	},
	"about_author": {
		Russian: "Разработчик: Serge Nook",
		English: "Developer: Serge Nook",
		Chinese: "开发者：Serge Nook",
	},
	"about_safe": {
		Russian: "Не изменяет Steam и не требует sudo.",
		English: "Does not modify Steam and does not require sudo.",
		Chinese: "不修改 Steam，也不需要 sudo。",
	},
	"open_website": {
		Russian: "Открыть SD-ON.RU",
		English: "Open SD-ON.RU",
		Chinese: "打开 SD-ON.RU",
	},
	"language": {
		Russian: "Язык",
		English: "Language",
		Chinese: "语言",
	},
	"gamepad_hotkeys": {
		Russian: "Горячие клавиши геймпада",
		English: "Gamepad hotkeys",
		Chinese: "手柄快捷键",
	},
	"gamepad_hotkeys_hint": {
		Russian: "Управление станциями, воспроизведением и громкостью с геймпада",
		English: "Control stations, playback and volume from the gamepad",
		Chinese: "通过手柄控制电台、播放和音量",
	},
	"hide_builtin": {
		Russian: "Скрыть станции по умолчанию",
		English: "Hide default stations",
		Chinese: "隐藏默认电台",
	},
	"hide_builtin_hint": {
		Russian: "В списке останутся только пользовательские станции",
		English: "Only custom stations remain in the list",
		Chinese: "列表中仅保留自定义电台",
	},
	"enabled": {
		Russian: "Включены",
		English: "Enabled",
		Chinese: "已启用",
	},
	"hidden": {
		Russian: "Скрыты",
		English: "Hidden",
		Chinese: "已隐藏",
	},
}
