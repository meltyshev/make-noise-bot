package game

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/meltyshev/make-noise-bot/internal/htmltext"
	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

const NameLite = "DozorLite"

var liteStatuses = map[int]string{
	1:  "Игра не началась.",
	2:  "Неверный PIN.",
	3:  "Авторизация пройдена успешно.",
	4:  "Не введен код.",
	5:  "Время на отправку кода вышло. Для получения следующего задания обновите страницу.",
	6:  "Время на отправку кода вышло. Вы прошли все уровни.",
	7:  "⚠️ Код не принят. Вы уже ввели этот составной код. Ищите другой.",
	8:  "✅ Код принят. Ищите следующий составной код.",
	9:  "✅ Код принят. Выполняйте следующее задание.",
	10: "Игра закончена. Вы прошли все уровни.",
	11: "❌ Код не принят.",
	12: "Вы вводили неверный код больше 4 раз. Прием данных от Вас заблокирован на три минуты. Повторите попытку позже.",
	13: "Движок остановлен организатором.",
	14: "Игра вашей команды приостановлена.",
	15: "Вам не запланировано следующее задание.",
	16: "⚠️ Код не принят. Вы уже ввели этот код.",
	17: "✅ Код принят.",
	18: "Вы не можете взять подсказку.",
	19: "✅ Принят финишный код.",
	20: "⚠️ Код не принят. Время, отведенное на игру, вышло.",
	21: "Неверная авторизация.",
	22: "Заявка на игру не подана или не принята.",
	23: "⚠️ Вы пытаетесь ввести уже принятый бонусный код.",
	24: "✅ Принят бонусный код.",
	25: "✅ Принят бонусный код. Вам начислено дополнительное бонусное время на нахождение всех бонусных кодов.",
	26: "Это ложный код. За его нахождение ваша команда получила штраф.",
	27: "❌ Код не принят. Вы превысили лимит попыток неправильного ввода кода. Предыдущее задание считается невзятым. Вам выдано следующее задание.",
	28: "Вы превысили лимит попыток неправильного ввода кода. Предыдущее задание считается невзятым. Вам выдано следующее задание.",
	29: "Это ложный код. За его нахождение ваша команда получила штраф. Вы превысили лимит попыток неправильного ввода кода. Предыдущее задание считается невзятым. Вам выдано следующее задание.",
	30: "⚠️ Универсальный код не принят. Вы уже ввели этот универсальный код. Ищите другой.",
	31: "⚠️ Мастер-код не принят. Вы уже ввели этот мастер-код. Ищите другой.",
	32: "✅ Универсальный код принят. Ищите следующий составной код.",
	33: "✅ Мастер-код принят. Ищите следующий составной код.",
	34: "✅ Универсальный код принят. Выполняйте следующее задание.",
	35: "✅ Мастер-код принят. Выполняйте следующее задание.",
	36: "В этом задании нельзя использовать универсальный код.",
	37: "Вы нашли все основные коды. Вы можете продолжать искать бонусные коды или перейти на следующий уровень.",
	38: "Вы не можете отправить основной код после ввода финишного.",
	39: "⚠️ Код к сквозному заданию не принят, так как истекло время его выполнения.",
	40: "✅ Код принят.",
	41: "❌ Код к спойлеру не принят.",
	42: "Подсказка выдана.",
	43: "⚠️ Данный код уже был найден вашей или другой командой. Ищите другой код.",
}

var liteAccepted = map[int]bool{8: true, 9: true, 17: true, 24: true, 25: true, 40: true}

var (
	liteLevelNumberRe = regexp.MustCompile(`<!--levelNumberBegin-->(\d+)<!--levelNumberEnd-->`)
	liteQuestionRe    = regexp.MustCompile(`(?s)<!--levelTextBegin-->(.*?)<!--levelTextEnd-->`)
	liteSectorsRe     = regexp.MustCompile(`<strong>Коды сложности</strong>(.*?)</div>`)
	liteProgressRe    = regexp.MustCompile(`\(Всего - (\d+) ?(, для прохождения достаточно любых (\d+) ?)?, принято - (\d+)\)`)
	liteTimeRe        = regexp.MustCompile(`<!--timeOnLevelBegin (\d+) timeOnLevelEnd-->`)
	liteHintsRe       = regexp.MustCompile(`(?s)<!--LevelClue(\d)Text-->(.*?)<!--LevelClue\dTextEnd-->`)

	// Spoilers live between the level text and the code counts, either as an
	// open block or as a line offering the form to unlock them.
	liteSpoilerAreaRe   = regexp.MustCompile(`(?s)<!--levelTextEnd-->(.*?)(?:<!--bonusCodeCount|<!--mainCodeCount|<!--difficultyCods|<div class='dcodes'|<p>Введите код)`)
	liteSpoilerOpenRe   = regexp.MustCompile(`<div class=['"]?spoiler['"]?[^>]*>`)
	liteSpoilerTitleRe  = regexp.MustCompile(`(?is)<div class=['"]?title['"]?[^>]*>\s*Спойлер\s*(?:№\s*)?(\d*)[^<]*</div>`)
	liteSpoilerClosedRe = regexp.MustCompile(`(?i)спойлер\s*№\s*(\d+)`)
)

func startLite(cfg store.GameConfig) *store.Game {
	game := newGameFromConfig(cfg)
	game.Pincode = cfg.Pincode
	return game
}

type Lite struct {
	env  *Env
	link string
}

func newLite(g store.Game, env *Env) *Lite {
	return &Lite{
		env:  env,
		link: fmt.Sprintf("%s/%s/go/?pin=%s", env.LiteBaseURL, g.City, g.Pincode),
	}
}

func (l *Lite) Name() string { return NameLite }
func (l *Lite) Link() string { return l.link }

func (l *Lite) EnterCode(ctx context.Context, code string, _ *int) EnterCodeResult {
	return l.submit(ctx, "entcod", "cod", code)
}

// EnterSpoilerCode uses the page's own spoiler form, which differs from the
// level one only by the action and the field carrying the code.
func (l *Lite) EnterSpoilerCode(ctx context.Context, code string) EnterCodeResult {
	return l.submit(ctx, "spoilerCode", "spoilerCode", code)
}

func (l *Lite) submit(ctx context.Context, action, field, code string) EnterCodeResult {
	// The lite engine expects multipart here, not urlencoded.
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for name, value := range map[string]string{
		"action": action,
		field:    encodeCP1251(code),
	} {
		fw, err := writer.CreateFormField(name)
		if err != nil {
			return EnterCodeResult{Message: texts.EngineTimeout}
		}
		if _, err := fw.Write([]byte(value)); err != nil {
			return EnterCodeResult{Message: texts.EngineTimeout}
		}
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, l.link, &body)
	if err != nil {
		return EnterCodeResult{Message: texts.EngineTimeout}
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", userAgent)

	resp, err := l.env.HTTP.Do(req)
	if err != nil {
		return EnterCodeResult{Message: texts.EngineTimeout}
	}
	raw, _ := readBody(resp)

	if resp.StatusCode == http.StatusFound {
		return resultFromStatus(liteStatuses, liteAccepted, resp.Header.Get("Location"))
	}

	l.env.debug("lite-"+action, raw)
	return EnterCodeResult{Message: texts.EngineUnknown}
}

func (l *Lite) Load(ctx context.Context) (Snapshot, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, l.link, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := l.env.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	raw, err := readBody(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lite load: HTTP %d", resp.StatusCode)
	}

	data, _, _ := strings.Cut(decodeBody(resp, raw), "<!--BonusLevels-->")
	return &liteSnapshot{link: l.link, data: data}, nil
}

type liteSnapshot struct {
	link string
	data string
}

func (s *liteSnapshot) LevelNumber() *int {
	match := liteLevelNumberRe.FindStringSubmatch(s.data)
	if match == nil {
		return nil
	}
	n, err := strconv.Atoi(match[1])
	if err != nil {
		return nil
	}
	return intPtr(n)
}

func (s *liteSnapshot) Progress() string {
	var progress []string

	if match := liteProgressRe.FindStringSubmatch(s.data); match != nil {
		codesNeeded := match[1]
		if match[3] != "" {
			codesNeeded = match[3]
		}
		progress = append(progress, fmt.Sprintf("%s/%s", match[4], codesNeeded))
	}

	if match := liteTimeRe.FindStringSubmatch(s.data); match != nil {
		total, err := strconv.Atoi(match[1])
		if err == nil {
			progress = append(progress, fmt.Sprintf("%02d:%02d:%02d", total/3600, total/60%60, total%60))
		}
	}

	return strings.Join(progress, " ")
}

func (s *liteSnapshot) TimeOnLevel() (int, bool) {
	match := liteTimeRe.FindStringSubmatch(s.data)
	if match == nil {
		return 0, false
	}
	seconds, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, false
	}
	return seconds, true
}

func (s *liteSnapshot) Question() string {
	match := liteQuestionRe.FindStringSubmatch(s.data)
	if match == nil {
		return ""
	}
	return htmltext.Convert(match[1], s.link)
}

func (s *liteSnapshot) Notes() string { return "" }

func (s *liteSnapshot) Sectors() []Sector {
	blockMatch := liteSectorsRe.FindStringSubmatch(s.data)
	if blockMatch == nil {
		return nil
	}

	mainCounter, bonusCounter := 1, 1
	var mainSectors, bonusSectors []Sector

	for _, row := range betweenRows(blockMatch[1], "<br>") {
		idx := strings.LastIndex(row, ": ")
		if idx < 0 {
			continue
		}

		name := strings.TrimLeftFunc(row[:idx], unicode.IsSpace)
		name = strings.ReplaceAll(name, ":", ",")
		name = strings.ReplaceAll(name, "  ", " ")

		isMain := strings.HasSuffix(name, "основные коды")
		right := row[idx+2:]
		isNonstandard := strings.HasPrefix(right, "<br />")

		name = capitalize(name)

		var rawCodes []string
		if isNonstandard {
			rawCodes = strings.Split(right, "<br />")[1:]
		} else {
			rawCodes = strings.Split(right, ", ")
		}

		var codes []SectorCode
		for _, rawCode := range rawCodes {
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
			if isNonstandard {
				if inner, ok := firstBetween(rawCode, "(", ")"); ok {
					hazard = inner
				}
			} else if entered {
				if inner, ok := firstBetween(rawCode, ">", "<"); ok {
					hazard = inner
				}
			}

			codes = append(codes, SectorCode{
				Number:  number,
				Hazard:  liteHazard(hazard),
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

func (s *liteSnapshot) Hint() (int, string) {
	matches := liteHintsRe.FindAllStringSubmatch(s.data, -1)
	if len(matches) == 0 {
		return 0, ""
	}
	last := matches[len(matches)-1]
	number, err := strconv.Atoi(last[1])
	if err != nil {
		return 0, ""
	}
	return number, htmltext.Convert(last[2], s.link)
}

// Spoilers reads both forms and numbers them by their order on the page,
// since only the closed ones name their number.
func (s *liteSnapshot) Spoilers() []Spoiler {
	area := liteSpoilerAreaRe.FindStringSubmatch(s.data)
	if area == nil {
		return nil
	}
	region := area[1]

	type entry struct {
		at     int
		number int
		open   bool
		text   string
	}
	var (
		entries []entry
		blocks  [][2]int
	)

	for _, marker := range liteSpoilerOpenRe.FindAllStringIndex(region, -1) {
		if insideAny(blocks, marker[0]) {
			continue
		}
		body, after := divBody(region, marker[1])
		blocks = append(blocks, [2]int{marker[0], after})

		number := 0
		if title := liteSpoilerTitleRe.FindStringSubmatchIndex(body); title != nil {
			if digits := body[title[2]:title[3]]; digits != "" {
				number, _ = strconv.Atoi(digits)
			}
			body = body[title[1]:]
		}
		entries = append(entries, entry{
			at:     marker[0],
			number: number,
			open:   true,
			text:   htmltext.Convert(body, s.link),
		})
	}

	for _, closed := range liteSpoilerClosedRe.FindAllStringSubmatchIndex(region, -1) {
		if insideAny(blocks, closed[0]) {
			continue
		}
		number, err := strconv.Atoi(region[closed[2]:closed[3]])
		if err != nil {
			continue
		}
		entries = append(entries, entry{at: closed[0], number: number})
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].at < entries[j].at })

	spoilers := make([]Spoiler, 0, len(entries))
	for i, item := range entries {
		if item.number == 0 {
			item.number = i + 1
		}
		spoilers = append(spoilers, Spoiler{Number: item.number, Open: item.open, Text: item.text})
	}
	return spoilers
}

// insideAny reports whether a position falls into an open spoiler, where a
// mention of a spoiler number belongs to its text.
func insideAny(blocks [][2]int, at int) bool {
	for _, block := range blocks {
		if at >= block[0] && at < block[1] {
			return true
		}
	}
	return false
}

// divBody returns the content of a div whose opening tag ends at from, and
// the position right after its closing tag.
func divBody(region string, from int) (string, int) {
	depth := 1
	for at := from; at < len(region); {
		opening := strings.Index(region[at:], "<div")
		closing := strings.Index(region[at:], "</div>")
		if closing < 0 {
			break
		}

		if opening >= 0 && opening < closing {
			depth++
			at += opening + len("<div")
			continue
		}

		depth--
		if depth == 0 {
			return region[from : at+closing], at + closing + len("</div>")
		}
		at += closing + len("</div>")
	}
	return region[from:], len(region)
}

func liteHazard(hazard string) string {
	if hazard == "null" {
		return "N"
	}
	return hazard
}
