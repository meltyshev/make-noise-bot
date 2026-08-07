package game

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/meltyshev/make-noise-bot/internal/htmltext"
	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

const NameClassic = "DozorClassic"

var classicStatuses = map[int]string{
	1:  "Игра не началась.",
	2:  "Неверный PIN.",
	3:  "Авторизация пройдена успешно.",
	4:  "Не введен код.",
	5:  "Время на задание вышло. Решайте следующее задание.",
	7:  "⚠️ Вы уже вводили этот код.",
	8:  "✅ Код принят. Ищите следующий составной код.",
	9:  "✅ Код принят. Выполняйте следующее задание.",
	10: "Спасибо за игру. Игра закончена.",
	11: "❌ Код не принят.",
	12: "Вы вводили неверный код больше 4 раз. Прием данных от вас заблокирован на три минуты. Повторите попытку позже.",
	13: "Движок остановлен организатором.",
	14: "Игра вашей команды приостановлена.",
	15: "Вам не запланировано следующее задание. Организатор видит, что вы бездействуете и назначит вам уровень в ближайшее время. Периодически обновляйте страницу. Время за задержку будет вычтено из вашего результата. Если новое задание не будет выдаваться длительное время, свяжитесь с организатором.",
	16: "✅ Код принят.",
	17: "Время на отправку кода вышло.",
	21: "Акаунт заблокирован или не активирован. Выполните инструкции, высланные вам в письме-подтверждении.",
	22: "Авторизация пройдена успешно.",
	23: "Неверный пароль.",
	24: "Неизвестный пользователь.",
	25: "Ошибка авторизации.",
	26: "Вы уже взяли 2 перерыва. Больше вы не можете приостанавливать игру своей команды.",
	27: "Игра вашей команды приостановлена на 15 минут по решению штаба.",
	28: "Вы уже дважды отказывались от заданий. Больше вы не можете завершать задания досрочно.",
	29: "Вам уже выдано следующее задание.",
	30: "Вам не запланировано следующее задание. Чтобы отказаться от текущего, свяжитесь с организатором и попросите его назначить вам следующий уровень.",
	31: "❌ Код к сквозному бонусному заданию не принят.",
	32: "✅ Код к сквозному бонусному заданию принят.",
	33: "Ваше сообщение отправлено организатору.",
	34: "✅ Код к сквозному бонусному заданию принят.",
	35: "⚠️ Вы уже вводили этот код.",
	36: "✅ Код принят.",
	37: "✅ Код к сквозному бонусному заданию принят.",
	38: "❌ Код не принят. Вы превысили лимит попыток неправильного ввода кода. Предыдущее задание считается невзятым. Вам выдано следующее задание.",
	39: "Вы решили отказаться от выполнения задания. Вам выдано новое задание.",
	40: "Это ложный код. За его нахождение ваша команда получила штраф.",
	41: "Вы нашли все основные коды. Вы можете продолжать искать бонусные коды или перейти на следующий уровень.",
	42: "За досрочное использование подсказки вам начислен штраф.",
	43: "Вы ввели не обязательный основной код, задание уже считается выполненным. Ищите бонусные коды.",
	44: "В этом задании нельзя использовать универсальный код.",
	45: "Вы уже использовали универсальный код, повторное его использование невозможно.",
	46: "В бонусном сквозном задании нельзя использовать универсальный код.",
	47: "✅ Универсальный код принят.",
	48: "✅ Универсальный код принят. Выполняйте следующее задание.",
	49: "✅ Универсальный код принят.",
	50: "✅ Мастер-код принят.",
	51: "✅ Мастер-код принят. Выполняйте следующее задание.",
	52: "✅ Мастер-код принят.",
	53: "⚠️ Данный код уже был найден. Ищите другой код.",
	54: "⚠️ Код к сквозному заданию не принят, так как истекло время его выполнения.",
	55: "✅ Код к спойлеру принят.",
	56: "❌ Код к спойлеру не принят.",
	57: "У вас недостаточно прав для использования данной функции.",
	58: "Игра еще не началась.",
}

var classicAccepted = map[int]bool{8: true, 9: true, 16: true, 36: true, 55: true}

// fixResponseRe strips leading garbage before the first "{" or "[".
var fixResponseRe = regexp.MustCompile(`^[^\{\[]*`)

// obtainClassicSession is also used by the prequel engines. The API wants an
// empty basic auth header.
func obtainClassicSession(ctx context.Context, env *Env, city, login, password string) (string, error) {
	loginURL := fmt.Sprintf("%s/%s/API/login.php?%s", env.ClassicBaseURL, city, url.Values{
		"login":    {login},
		"password": {password},
	}.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, loginURL, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth("", "")
	req.Header.Set("User-Agent", userAgent)

	resp, err := env.HTTPFollow.Do(req)
	if err != nil {
		return "", err
	}
	raw, err := readBody(resp)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login: HTTP %d", resp.StatusCode)
	}

	var payload struct {
		Code      json.Number `json:"code"`
		UserToken string      `json:"userToken"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		env.debug("classic-login", raw)
		return "", fmt.Errorf("login: %w", err)
	}
	if code, ok := asInt(payload.Code); !ok || code != 2 {
		return "", fmt.Errorf("login: engine code %s", payload.Code)
	}
	return payload.UserToken, nil
}

func startClassic(ctx context.Context, cfg store.GameConfig, env *Env) (*store.Game, error) {
	session, err := obtainClassicSession(ctx, env, cfg.City, cfg.Login, cfg.Password)
	if err != nil {
		return nil, err
	}
	game := newGameFromConfig(cfg)
	game.Login = cfg.Login
	game.Password = cfg.Password
	game.Session = session
	return game, nil
}

type Classic struct {
	env      *Env
	city     string
	session  string
	login    string
	password string
	link     string
}

func newClassic(g store.Game, env *Env) *Classic {
	return &Classic{
		env:      env,
		city:     g.City,
		session:  g.Session,
		login:    g.Login,
		password: g.Password,
		link:     fmt.Sprintf("%s/%s/go/", env.ClassicBaseURL, g.City),
	}
}

func (c *Classic) Name() string { return NameClassic }
func (c *Classic) Link() string { return c.link }

func (c *Classic) relogin(ctx context.Context) bool {
	session, err := obtainClassicSession(ctx, c.env, c.city, c.login, c.password)
	if err != nil {
		return false
	}
	c.session = session
	c.env.sessionUpdated(session)
	return true
}

func (c *Classic) EnterCode(ctx context.Context, code string, pinnedLevel *int) EnterCodeResult {
	result, sessionSuspect := c.enterCodeOnce(ctx, code, pinnedLevel)
	if sessionSuspect && c.relogin(ctx) {
		result, _ = c.enterCodeOnce(ctx, code, pinnedLevel)
	}
	return result
}

func (c *Classic) enterCodeOnce(ctx context.Context, code string, pinnedLevel *int) (EnterCodeResult, bool) {
	form := url.Values{
		"action": {"entcod"},
		"cod":    {encodeCP1251(code)},
	}
	if pinnedLevel != nil {
		form.Set("level", fmt.Sprint(*pinnedLevel))
		form.Set("skvoz", "1")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.link, strings.NewReader(form.Encode()))
	if err != nil {
		return EnterCodeResult{Message: texts.EngineTimeout}, false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.setHeaders(req)

	resp, err := c.env.HTTP.Do(req)
	if err != nil {
		return EnterCodeResult{Message: texts.EngineTimeout}, false
	}
	raw, _ := readBody(resp)

	if resp.StatusCode == http.StatusFound {
		return resultFromStatus(classicStatuses, classicAccepted, resp.Header.Get("Location")), false
	}

	// A non-redirect answer usually means the session died; retry once
	// after re-login.
	c.env.debug("classic-entcod", raw)
	return EnterCodeResult{Message: texts.EngineUnknown}, true
}

func (c *Classic) setHeaders(req *http.Request) {
	req.Header.Set("Cookie", "dozorSiteSession="+c.session)
	req.Header.Set("Referer", c.link)
	req.Header.Set("User-Agent", userAgent)
}

func (c *Classic) Load(ctx context.Context) (Snapshot, error) {
	snap, err, sessionSuspect := c.loadOnce(ctx)
	if sessionSuspect && c.relogin(ctx) {
		snap, err, _ = c.loadOnce(ctx)
	}
	return snap, err
}

func (c *Classic) loadOnce(ctx context.Context) (Snapshot, error, bool) {
	loadURL := fmt.Sprintf("%s?%s", c.link, url.Values{
		"s":   {c.session},
		"api": {"true"},
	}.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, loadURL, nil)
	if err != nil {
		return nil, err, false
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.env.HTTPFollow.Do(req)
	if err != nil {
		return nil, err, false
	}
	raw, err := readBody(resp)
	if err != nil {
		return nil, err, false
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("classic load: HTTP %d", resp.StatusCode), false
	}

	content := decodeBody(resp, raw)
	// The engine emits invalid JSON: an unquoted null key and occasional
	// junk before the payload.
	content = strings.ReplaceAll(content, "null:{", `"null":{`)
	content = fixResponseRe.ReplaceAllString(content, "")

	decoder := json.NewDecoder(strings.NewReader(content))
	decoder.UseNumber()
	var data map[string]any
	if err := decoder.Decode(&data); err != nil {
		// Most likely an HTML login page: the session expired.
		c.env.debug("classic-load", raw)
		return nil, fmt.Errorf("classic load: %w", err), true
	}

	snap := &classicSnapshot{link: c.link}
	if level, ok := data["level"].(map[string]any); ok && truthy(data["level"]) {
		snap.level = level
		if _, ok := asInt(level["levelNumber"]); !ok {
			c.env.debug("classic-load", raw)
			return nil, errors.New("classic load: unreadable levelNumber"), false
		}
	}
	return snap, nil, false
}

type classicSnapshot struct {
	link  string
	level map[string]any // nil when there is no active level
}

func (s *classicSnapshot) LevelNumber() *int {
	if s.level == nil {
		return nil
	}
	n, _ := asInt(s.level["levelNumber"])
	return intPtr(n)
}

func (s *classicSnapshot) Progress() string {
	if s.level == nil {
		return ""
	}

	codesNeeded := asString(s.level["neededCodes"])
	if n, ok := asInt(s.level["neededCodes"]); ok && n == 0 {
		codesNeeded = asString(s.level["totalCodes"])
	}

	progress := []string{fmt.Sprintf("%s/%s", asString(s.level["codesFounded"]), codesNeeded)}

	if truthy(s.level["bonusCodesTotal"]) {
		progress = append(progress, fmt.Sprintf(
			"%s/%s",
			asString(s.level["bonusCodesFounded"]),
			asString(s.level["bonusCodesTotal"]),
		))
	}

	progress = append(progress, asString(s.level["timeOnLevel"]))
	return strings.Join(progress, " ")
}

func (s *classicSnapshot) TimeOnLevel() (int, bool) {
	if s.level == nil {
		return 0, false
	}
	return parseClock(asString(s.level["timeOnLevel"]))
}

// parseClock reads "H:MM:SS" and "MM:SS" into seconds.
func parseClock(value string) (int, bool) {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, false
	}

	total := 0
	for _, part := range parts {
		number, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || number < 0 {
			return 0, false
		}
		total = total*60 + number
	}
	return total, true
}

func (s *classicSnapshot) Question() string {
	if s.level == nil {
		return ""
	}
	return htmltext.Convert(asString(s.level["question"]), s.link)
}

func (s *classicSnapshot) Notes() string {
	if s.level == nil {
		return ""
	}
	return htmltext.Convert(asString(s.level["locationComment"]), s.link)
}

func (s *classicSnapshot) Sectors() []Sector {
	if s.level == nil {
		return nil
	}

	koline := asString(s.level["koline"])
	rows := strings.Split(koline, "<br>")
	if len(rows) < 2 {
		return nil
	}
	rows = rows[:len(rows)-1]

	mainCounter, bonusCounter := 1, 1
	var mainSectors, bonusSectors []Sector

	for _, row := range rows {
		idx := strings.LastIndex(row, ": ")
		if idx < 0 {
			continue
		}

		name := strings.TrimLeftFunc(row[:idx], unicode.IsSpace)
		name = strings.ReplaceAll(name, ":", ",")
		name = strings.ReplaceAll(name, "  ", " ")

		isMain := strings.HasSuffix(name, "основные коды")
		name = capitalize(name)

		var codes []SectorCode
		for _, rawCode := range strings.Split(row[idx+2:], ", ") {
			var number int
			if isMain {
				number = mainCounter
				mainCounter++
			} else {
				number = bonusCounter
				bonusCounter++
			}

			entered := strings.HasPrefix(rawCode, "<")
			hazard := rawCode
			if entered {
				if inner, ok := firstBetween(rawCode, ">", "<"); ok {
					hazard = inner
				}
			}

			codes = append(codes, SectorCode{
				Number:  number,
				Hazard:  classicHazard(hazard),
				Entered: entered,
			})
		}

		sector := Sector{Name: name, Codes: codes}
		if isMain {
			mainSectors = append(mainSectors, sector)
		} else {
			bonusSectors = append(bonusSectors, sector)
		}
	}

	return append(mainSectors, bonusSectors...)
}

func (s *classicSnapshot) Hint() (int, string) {
	if s.level == nil {
		return 0, ""
	}
	for i := 2; i >= 1; i-- {
		hint := s.level[fmt.Sprintf("hint%d", i)]
		if truthy(hint) {
			return i, htmltext.Convert(asString(hint), s.link)
		}
	}
	return 0, ""
}

func (s *classicSnapshot) Spoilers() []Spoiler {
	if s.level == nil {
		return nil
	}
	spoilers, ok := s.level["spoilers"].([]any)
	if !ok {
		return nil
	}

	var out []Spoiler
	for _, item := range spoilers {
		spoiler, ok := item.(map[string]any)
		if !ok {
			continue
		}
		number, ok := asInt(spoiler["spoilerNumber"])
		if !ok {
			continue
		}
		out = append(out, Spoiler{Number: number, Open: truthy(spoiler["spoilerSolved"])})
	}
	return out
}

func classicHazard(hazard string) string {
	hazard, _, _ = strings.Cut(hazard, ":")
	if hazard == "null" {
		return "N"
	}
	return hazard
}

func capitalize(s string) string {
	runes := []rune(strings.ToLower(s))
	if len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
	}
	return string(runes)
}
