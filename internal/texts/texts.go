// Package texts collects all user-facing strings.
package texts

// General.
const (
	Start        = "Привет, %s!"
	Done         = "Готово!"
	AdminClaimed = "Привет, %s! Ты назначен админом бота."

	NoActiveGame     = "Нет активной игры."
	CannotLoadEngine = "Не могу загрузить движок."
	ChatNotFound     = "Чат не найден."

	EngineTimeout = "Движок не отвечает."
	EngineUnknown = "Не понимаю ответ движка."

	CancelDone    = "Команда /%s отменена."
	CancelNothing = "Нет текущих команд."

	HTMLFallback = "Не удалось разобрать разметку, отправляю как есть:"
)

// Permissions.
const (
	PermissionNeeded    = "Для использования команды /%s в этом чате нужно разрешение. /permission - сделать запрос."
	PermissionPending   = "Для использования команды /%s в этом чате нужно разрешение. Вы уже сделали запрос, ожидайте решения."
	PermissionForbidden = "Вам запрещено мной пользоваться!"

	PermissionStatusRequested = "Текущий статус - запрошено."
	PermissionStatusAllowed   = "Текущий статус - разрешено."
	PermissionStatusForbidden = "Текущий статус - запрещено."
	PermissionRequestSent     = "Запрос сделан, ожидайте решения."
	PermissionGranted         = "Разрешение получено \U0001f60d"
)

// Utility commands.
const (
	MorseAsk      = "Какую морзянку хочешь перевести?"
	MorseRequired = "Ты должен ввести морзянку!"

	NumbersAsk      = "Какие числа хочешь перевести?"
	NumbersRequired = "Ты должен ввести цифры!"

	LettersAsk      = "Какие буквы хочешь перевести?"
	LettersRequired = "Ты должен ввести буквы!"

	IntersectionAsk      = "Какие слова хочешь пересечь?"
	IntersectionRequired = "Ты должен ввести слова!"
	IntersectionTooFew   = "Ты должен ввести хотя бы 2 слова!"
	IntersectionEmpty    = "По введенным словам нет пересечений \U0001f614"

	AnagramAsk         = "Из каких буквы хочешь получить анаграмму?"
	AnagramRequired    = "Ты должен ввести буквы!"
	AnagramOnlyLetters = "Символы, кроме букв, запрещены!"
	AnagramNotFound    = "Из введенных буквы анаграмма не найдена \U0001f614"
	AnagramUnavailable = "Не могу получить анаграммы."

	MaskAsk         = "По какой маске хочешь получить слова?"
	MaskRequired    = "Ты должен ввести маску!"
	MaskOnlyLetters = "Символы, кроме букв, «-» и «_», запрещены!"
	MaskNotFound    = "По введенной маске слова не найдены \U0001f614"

	CoordinatesAsk      = "Кидай местоположение."
	CoordinatesRequired = "Ты должен скинуть местоположение!\n/cancel - отменить текущую команду."

	AvatarUsage = "Надо так - /avatar <цвета_фона> <цвет_текста> <[имена]>"
	AvatarAsk   = "Какие параметры? (<цвета_фона> <цвет_текста> <[имена]>)"

	TopAll = "Все топ!"
)

// Game commands.
const (
	GameOver           = "Игра окончена."
	GameCannotStart    = "Не могу получить сессию."
	QuestionNone       = "Задания нет."
	NotesNone          = "Примечаний нет."
	InfoNone           = "Информации нет."
	RatingNone         = "Рейтинга пока нет."
	RatingCleared      = "Рейтинг очищен."
	SubscribeOn        = "Подписка активирована."
	SubscribeOff       = "Подписка отменена."
	RestrictOn         = "Ввод кодов ограничен."
	RestrictOff        = "Ограничение на ввод кодов снято."
	BruteForceOn       = "Режим перебора активирован."
	BruteForceOff      = "Режим перебора отключен."
	CodeFormatsAsk     = "Какие форматы кода?"
	CodeFormatsInvalid = "Ты должен ввести форматы кода!"
	PinLevelAsk        = "Какой номер уровня?"
	PinLevelRequired   = "Ты должен ввести номер уровня!"
)

// Updater broadcasts.
const (
	LevelUp       = "АП!"
	HintFmt       = "Подсказка %d:\n\n%s"
	SpoilerSolved = "Спойлер %d - АП!"
)

// Admin commands.
const (
	AskChatID      = "Какой chat id?"
	AskChatIDWrite = "Какой chat_id?"
	WriteWhat      = "Что отправить?"

	ChatsChoose      = "Выберите чат"
	ChatsDeleted     = "Чат удален"
	ChatsClosed      = "Настройка чатов завершена"
	ChatsActionsFmt  = "%s | %d\n\nТип: %s\nСтатус: %s\n\nВыберите действие"
	ButtonClose      = "Закрыть"
	ButtonBack       = "Назад"
	ButtonDelete     = "Удалить"
	ButtonCancel     = "Отмена"
	ButtonReset      = "Сбросить"
	ButtonAdd        = "Добавить"
	ButtonAllow      = "✅ Разрешить"
	ButtonForbid     = "🚫 Запретить"
	SettingsTitle    = "Настройки:"
	ManagersTitle    = "Менеджеры:"
	ManagersCountFmt = "Менеджеры: %d"
	LeaveModeFmt     = "Режим выхода: %s"
	LeaveModeOnWord  = "включен"
	LeaveModeOffWord = "выключен"
	LeftChatFmt      = "Я покинул %s (%d)."

	PermissionRequestTitleFmt = "Запрос для @%s:\n"
	PermissionAllowedMark     = "Разрешено ✅"
	PermissionForbiddenMark   = "Запрещено 🚫"
	AlreadyProcessed          = "Уже обработано."
	NoAccess                  = "Нет доступа."

	PickUserAsk    = "Выберите пользователя:"
	ButtonPickUser = "Выбрать пользователя"

	TextRequired = "Ты должен ответить текстом!"
)

// Game config menu.
const (
	GameConfigEngineFmt      = "Движок: %s"
	GameConfigCityFmt        = "Город: %s"
	GameConfigLoginFmt       = "Логин: %s"
	GameConfigPasswordFmt    = "Пароль: %s"
	GameConfigPincodeFmt     = "Пинка: %s"
	GameConfigGameIDFmt      = "Номер игры: %s"
	GameConfigLeagueFmt      = "Лига: %s"
	GameConfigFormatsFmt     = "Форматы кода: %s"
	GameConfigSubscribersFmt = "Подписчики: %d"

	GameConfigEngineAsk   = "Движок:"
	GameConfigCityAsk     = "Город:"
	GameConfigLoginAsk    = "Логин:"
	GameConfigPasswordAsk = "Пароль:"
	GameConfigPincodeAsk  = "Пинка:"
	GameConfigGameIDAsk   = "Номер игры:"
	GameConfigLeagueAsk   = "Лига:"
	GameConfigFormatsAsk  = "Форматы кода:"

	SubscribersTitle        = "Подписчики:"
	GameSubscribersTitle    = "Подписчики текущей игры:"
	GameSubscribersCountFmt = "Подписчики игры: %d"
)

// Command descriptions for /help and the Telegram command menu.
const (
	DescNumbersToLetters = "цифры в буквы"
	DescLettersToNumbers = "буквы в цифры"
	DescIntersection     = "пересечение"
	DescMorse            = "морзе"
	DescAnagram          = "анаграмма"
	DescMask             = "слова по маске"
	DescLink             = "ссылка"
	DescQuestion         = "задание"
	DescNotes            = "примечания"
	DescRating           = "рейтинг"
	DescCancel           = "отменить текущую команду"
	DescHelp             = "помощь"
)
