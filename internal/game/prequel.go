package game

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/meltyshev/make-noise-bot/internal/store"
	"github.com/meltyshev/make-noise-bot/internal/texts"
)

const (
	NameClassicPrequel = "DozorClassicPrequel"
	NameLitePrequel    = "DozorLitePrequel"
)

var prequelSectorsRe = regexp.MustCompile(`<strong id=orang>Код сложности:(.*?)</strong>`)

// prequelStatuses is shared by both prequel engines, which differ only in
// their page URL.
var prequelStatuses = map[int]string{
	52: "⚠️ Код к приквелу не принят. Ваша команда уже отправила этот код к приквелу.",
	53: "❌ Код к приквелу не принят. Вы ввели неверный код.",
	54: "✅ Код к приквелу принят.",
	55: "⚠️ Код к приквелу не принят. Закончилось отведенное время на прием кода.",
	56: "⚠️ Код к приквелу не принят. Вы исчерпали попытки для ввода кода приквела.",
}

var prequelAccepted = map[int]bool{54: true}

func startPrequel(ctx context.Context, name string, cfg store.GameConfig, env *Env) (*store.Game, error) {
	session, err := obtainClassicSession(ctx, env, cfg.City, cfg.Login, cfg.Password)
	if err != nil {
		return nil, err
	}
	game := newGameFromConfig(cfg)
	game.Engine = name
	game.Login = cfg.Login
	game.Password = cfg.Password
	game.GameID = cfg.GameID
	game.League = cfg.League
	game.Session = session
	return game, nil
}

type Prequel struct {
	env      *Env
	name     string
	city     string
	session  string
	login    string
	password string
	gameID   string
	league   string
	link     string
}

func newPrequel(g store.Game, env *Env) *Prequel {
	p := &Prequel{
		env:      env,
		name:     g.Engine,
		city:     g.City,
		session:  g.Session,
		login:    g.Login,
		password: g.Password,
		gameID:   g.GameID,
		league:   g.League,
	}
	if g.Engine == NameLitePrequel {
		p.link = fmt.Sprintf("%s/%s/?league=%s", env.LiteBaseURL, g.City, g.League)
	} else {
		p.link = fmt.Sprintf("%s/%s/?section=anons&league=%s", env.ClassicBaseURL, g.City, g.League)
	}
	return p
}

func (p *Prequel) Name() string { return p.name }
func (p *Prequel) Link() string { return p.link }

func (p *Prequel) relogin(ctx context.Context) bool {
	session, err := obtainClassicSession(ctx, p.env, p.city, p.login, p.password)
	if err != nil {
		return false
	}
	p.session = session
	p.env.sessionUpdated(session)
	return true
}

func (p *Prequel) setHeaders(req *http.Request) {
	req.Header.Set("Cookie", "dozorSiteSession="+p.session)
	req.Header.Set("Referer", p.link)
	req.Header.Set("User-Agent", userAgent)
}

func (p *Prequel) EnterCode(ctx context.Context, code string, _ *int) EnterCodeResult {
	result, sessionSuspect := p.enterCodeOnce(ctx, code)
	if sessionSuspect && p.relogin(ctx) {
		result, _ = p.enterCodeOnce(ctx, code)
	}
	return result
}

func (p *Prequel) EnterSpoilerCode(ctx context.Context, code string) EnterCodeResult {
	return p.EnterCode(ctx, code, nil)
}

func (p *Prequel) enterCodeOnce(ctx context.Context, code string) (EnterCodeResult, bool) {
	form := url.Values{
		"action": {"prequel_code_new"},
		"league": {p.league},
		"game":   {p.gameID},
		"cod":    {encodeCP1251(code)},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.link, strings.NewReader(form.Encode()))
	if err != nil {
		return EnterCodeResult{Message: texts.EngineTimeout}, false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	p.setHeaders(req)

	resp, err := p.env.HTTP.Do(req)
	if err != nil {
		return EnterCodeResult{Message: texts.EngineTimeout}, false
	}
	raw, _ := readBody(resp)

	if resp.StatusCode == http.StatusFound {
		return resultFromStatus(prequelStatuses, prequelAccepted, resp.Header.Get("Location")), false
	}

	p.env.debug("prequel-entcod", raw)
	return EnterCodeResult{Message: texts.EngineUnknown}, true
}

func (p *Prequel) Load(ctx context.Context) (Snapshot, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.link, nil)
	if err != nil {
		return nil, err
	}
	p.setHeaders(req)

	resp, err := p.env.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	raw, err := readBody(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prequel load: HTTP %d", resp.StatusCode)
	}

	return &prequelSnapshot{data: decodeBody(resp, raw)}, nil
}

type prequelSnapshot struct {
	data string
}

// LevelNumber is always 0: a prequel has one code board and no levels.
func (s *prequelSnapshot) LevelNumber() *int         { return new(0) }
func (s *prequelSnapshot) TimeOnLevel() (int, bool)  { return 0, false }
func (s *prequelSnapshot) Progress() string          { return "" }
func (s *prequelSnapshot) Question() string          { return "" }
func (s *prequelSnapshot) Hint() (int, string, bool) { return 0, "", false }
func (s *prequelSnapshot) Spoilers() []Spoiler       { return nil }

func (s *prequelSnapshot) Sectors() []Sector {
	blockMatch := prequelSectorsRe.FindStringSubmatch(s.data)
	if blockMatch == nil {
		return nil
	}

	counter := 1
	var sectors []Sector

	for _, row := range betweenRows(blockMatch[1], "<br>") {
		idx := strings.LastIndex(row, ": ")
		if idx < 0 {
			continue
		}

		var codes []SectorCode
		for rawCode := range strings.SplitSeq(row[idx+2:], ", ") {
			entered := strings.HasPrefix(rawCode, "<")
			hazard := rawCode
			if entered {
				if inner, ok := firstBetween(rawCode, ">", "<"); ok {
					hazard = inner
				}
			}

			codes = append(codes, SectorCode{
				Number:  counter,
				Hazard:  liteHazard(hazard),
				Entered: entered,
			})
			counter++
		}

		sectors = append(sectors, Sector{Name: row[:idx], Codes: codes})
	}

	return sectors
}
