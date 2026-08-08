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
	PermissionGranted         = "Разрешение получено 😍"
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
	IntersectionEmpty    = "По введенным словам нет пересечений 😔"

	AnagramAsk         = "Из каких буквы хочешь получить анаграмму?"
	AnagramRequired    = "Ты должен ввести буквы!"
	AnagramOnlyLetters = "Символы, кроме букв, запрещены!"
	AnagramNotFound    = "Из введенных буквы анаграмма не найдена 😔"
	AnagramUnavailable = "Не могу получить анаграммы."

	MaskAsk         = "По какой маске хочешь получить слова?"
	MaskRequired    = "Ты должен ввести маску!"
	MaskOnlyLetters = "Символы, кроме букв, «-» и «_», запрещены!"
	MaskNotFound    = "По введенной маске слова не найдены 😔"

	CoordinatesAsk      = "Кидай местоположение."
	CoordinatesRequired = "Ты должен скинуть местоположение!\n/cancel - отменить текущую команду."

	AvatarUsage = "Надо так - /avatar <цвета_фона> <цвет_текста> <[имена]>"
	AvatarAsk   = "Какие параметры? (<цвета_фона> <цвет_текста> <[имена]>)"

	TopAll = "Все топ!"
)

// Game commands.
const (
	GameOver        = "Игра окончена."
	GameCannotStart = "Не могу получить сессию."
	QuestionNone    = "Задания нет."
	NotesHeader     = "Примечания к заданию:"
	InfoNone        = "Информации нет."
	// SectorCodeEntered marks a code the team already took on the board.
	SectorCodeEntered = "ok"
	RatingNone        = "Рейтинга пока нет."
	RatingCleared     = "Рейтинг очищен."
	RestrictOn        = "Ввод кодов ограничен."
	RestrictOff       = "Ограничение на ввод кодов снято."
	BruteForceOn      = "Режим перебора активирован."
	BruteForceOff     = "Режим перебора отключен."
	PinLevelAsk       = "Какой номер уровня?"
	PinLevelRequired  = "Ты должен ввести номер уровня!"
)

// Updater broadcasts.
const (
	LevelUp          = "АП!"
	LevelGone        = "Уровня больше нет."
	ButtonAllowCodes = "Разрешить ввод кодов"
	ButtonStopGame   = "Завершить игру"
	// The callback data is not shown to anyone; it lives here so the
	// updater and the bot agree on it.
	CallbackAllowCodes = "res:off"
	CallbackStopGame   = "gm:stop"
	HintFmt            = "Подсказка %d:\n\n%s"
	HintNotice         = "Подсказка %d."
	SpoilerSolved      = "Спойлер %d - АП!"
)

// Admin commands.
const (
	AskChatID = "Какой chat id?"
	WriteWhat = "Что отправить?"

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
	MapServiceFmt    = "Карты для координат: %s"
	MapServiceTitle  = "Карты для координат:"
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

	PrivateOnly = "Эта команда доступна только в личке."

	TextRequired = "Ты должен ответить текстом!"

	// ErrorPrefix marks the error DM the maintainer gets from reportError.
	ErrorPrefix = "⚠️ "
	// ActiveMark prefixes the menu option that is currently in effect.
	ActiveMark = "✓ "

	// The permission request DM lists the raw chat fields, so the admin can
	// tell two similarly named chats apart.
	PermissionRequestTypeFmt  = "type: %s"
	PermissionFieldTitle      = "title"
	PermissionFieldUsername   = "username"
	PermissionFieldFirstName  = "first_name"
	PermissionFieldLastName   = "last_name"
	PermissionRequestFieldFmt = "\n%s: %s"
	PermissionRequestIDFmt    = "\nid: %d"
	UnknownNameFmt            = "ID %d"
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

	SubscribersTitle        = "Подписчики:"
	GameSubscribersTitle    = "Подписчики текущей игры:"
	GameSubscribersCountFmt = "Подписчики игры: %d"

	SubscriptionTitleFmt = "%s\n\nЧто получает этот чат:"
	ButtonAllUpdates     = "Всё"
	ButtonEventsOnly     = "Только события"
	ButtonUnsubscribe    = "Отписать"

	SummaryAll    = "всё"
	SummaryEvents = "события"

	// Summaries ride in the label next to mark(), the way subscriber modes do,
	// because a chat's permission has three states and a checkmark has two.
	SummaryRequested = "запрошено"
	SummaryForbidden = "запрещено"

	FormatsTitle     = "Форматы кода:"
	ButtonManual     = "Ввести вручную"
	FormatsManualAsk = "Введи форматы: группы через запятую, синонимы через =, первый вариант уходит в движок. Например: dr=др=--, rd=рд"
	FormatsInvalid   = "Не понял. Пример: dr=др=--, rd=рд"
	PresetDigitsOnly = "Только цифры"
	PresetDR         = "DR (dr, др, --)"
	PresetMoscow     = "Москва (dr, др, rd, рд, d, д, r, р)"
)

// Easter eggs: one phrase is picked at random per /maxwell and /romka.
var (
	MaxwellPhrases = []string{
		"Внутри тебя не будет пустоты, если ты шаурма",
		"Не имей 100 рублей, а имей 100 рецептов шаурмы",
		"Не откладывай на завтра ту шаурму, которую можно съесть сегодня",
		"Сколько волка шаурмой не корми, он все равно в лес смотрит",
		"А Васька слушает да ест (шаурму)",
		"В гостях хорошо, а в шаурмечной лучше",
		"Век живи, век шаурму ешь",
		"Ты на кого тут шаурму крошишь?",
		"Шаурмей, шаурмей - кто успел, тот и съел",
		"Шаурма всему голова",
		"Шаурма человека кормит, а дозор портит",
		"Шаурма человеку друг",
		"Шаурме - время, дозору - час",
		"Язык до шаурмячной доведет",
		"Голод не шаурма, вообще только шаурма - шаурма",
		"На чужую шаурму рот не разевай",
		"Готовь сани летом, а ингредиенты для шаурмы - постоянно",
		"Съел свою шаурму - помоги соседу",
		"Глаза боятся, а руки готовят шаурму",
		"Любишь жрать, люби и шаурму готовить",
		"Под лежачую шаурму соус не течет",
		"2 шаурмы - пара",
		"Шаурма не волк - в лес не убежит",
		"Глаза боятся, а руки шаурму делают",
		"После поедания шаурмы грязными кулаками не машут",
		"В чужую шаурмячную со своей шаурмой не ходят",
		"Семь раз съешь - один закажи",
	}

	RomkaPhrases = []string{
		"Укропчика не желаете?",
		"Развооорооот..",
		"А метку «DR не светить» тоже снимать?",
		"Привет, хе-хе:)",
		"Здорова, ёптыть!",
		"ПЕРВЫЙ КОД ЗА ИГРУ!",
		"Причина остановки? Не выходи из машины!",
		"Через изоплит быстрее! Я по Яндексу пробил..",
		"По-брааатски, включи бутырку!",
	}
)

// First-run wizard, printed to the console rather than to Telegram.
const (
	WizardNoConfig    = "Файл конфигурации не найден - настроим бота."
	WizardCreateBot   = "Создайте бота у @BotFather в Telegram и получите токен."
	WizardAskToken    = "Вставьте токен бота: "
	WizardBadToken    = "Токен не подошел (%v), попробуйте еще раз.\n"
	WizardNoToken     = "не введен токен"
	WizardNoGoodToken = "не введен рабочий токен"
	WizardSavedFmt    = "Готово, это @%s. Конфигурация сохранена в %s.\n"
	WizardClaimAdmin  = "Теперь отправьте боту /start - первый написавший станет админом.\n\n"
	WizardUseToken    = "либо запустите с флагом --token"
)

// Map services, shown in the /config picker.
const (
	MapYandex = "Яндекс.Карты"
	MapGoogle = "Google Maps"
	MapTwoGIS = "2ГИС"
	MapOSM    = "OpenStreetMap"
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
	DescRating           = "рейтинг"
	DescCancel           = "отменить текущую команду"
	DescHelp             = "помощь"
)
