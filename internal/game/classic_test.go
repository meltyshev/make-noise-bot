package game

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/meltyshev/make-noise-bot/internal/store"
)

// classicPayload is shaped like the real classic API answer: junk before the
// JSON, an unquoted null key, numbers arriving as strings, and hints whose
// quotes are escaped twice.
const classicPayload = `jQuery163({"level": {` +
	`"levelNumber": "3",` +
	`"neededCodes": "0",` +
	`"totalCodes": "5",` +
	`"codesFounded": "2",` +
	`"bonusCodesTotal": 2,` +
	`"bonusCodesFounded": 1,` +
	`"timeOnLevel": "0:15:00",` +
	`"question": "<b>Вопрос</b><br>тут",` +
	`"locationComment": "<p>Комментарий</p>",` +
	`"koline": " основные коды: 1.2, <span>1.3:вз</span>, null<br> бонусные коды: 2.1<br>",` +
	`"hint1": "От каждой \\\"птицы\\\" нужен хвост",` +
	`"hint2": "",` +
	`"spoilers": [{"spoilerSolved": 1, "spoilerNumber": "2"}, {"spoilerSolved": 0, "spoilerNumber": "3"}]` +
	`}, null:{"x": 1}})`

func TestClassicLoadAndSnapshot(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/e-burg/go/" || r.URL.Query().Get("api") != "true" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("s") != "sess-1" {
			t.Errorf("missing session, got query %q", r.URL.RawQuery)
		}
		w.Write([]byte(classicPayload))
	}))
	defer srv.Close()

	engine := newClassic(classicGame(), testEnv(srv))
	snap, err := engine.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if level := snap.LevelNumber(); level == nil || *level != 3 {
		t.Errorf("LevelNumber = %v, want 3", level)
	}
	// neededCodes 0 means "all codes": totalCodes is shown instead.
	if got, want := snap.Progress(), "2/5 1/2 0:15:00"; got != want {
		t.Errorf("Progress = %q, want %q", got, want)
	}
	// The location comment rides along with the task, labelled.
	if got, want := snap.Question(), "<b>Вопрос</b>\nтут\n\n<b>Примечания к заданию:</b>\nКомментарий"; got != want {
		t.Errorf("Question = %q, want %q", got, want)
	}

	// The doubled backslashes are gone, so the quotes reach Telegram as
	// quotes rather than as \" .
	hintNumber, hintText := snap.Hint()
	if want := "От каждой &#34;птицы&#34; нужен хвост"; hintNumber != 1 || hintText != want {
		t.Errorf("Hint = (%d, %q), want (1, %q)", hintNumber, hintText, want)
	}

	spoilers := snap.Spoilers()
	if len(spoilers) != 2 {
		t.Fatalf("Spoilers = %+v, want 2", spoilers)
	}
	if spoilers[0].Number != 2 || !spoilers[0].Open {
		t.Errorf("first spoiler = %+v, want open number 2", spoilers[0])
	}
	if spoilers[1].Number != 3 || spoilers[1].Open {
		t.Errorf("second spoiler = %+v, want closed number 3", spoilers[1])
	}

	sectors := snap.Sectors()
	if len(sectors) != 2 {
		t.Fatalf("Sectors = %d, want 2", len(sectors))
	}
	main := sectors[0]
	if main.Name != "Основные коды" {
		t.Errorf("main sector name = %q", main.Name)
	}
	if len(main.Codes) != 3 {
		t.Fatalf("main codes = %d, want 3", len(main.Codes))
	}
	if main.Codes[0].Hazard != "1.2" || main.Codes[0].Entered {
		t.Errorf("code 1 = %+v", main.Codes[0])
	}
	if main.Codes[1].Hazard != "1.3" || !main.Codes[1].Entered {
		t.Errorf("code 2 = %+v", main.Codes[1])
	}
	if main.Codes[2].Hazard != "N" || main.Codes[2].Entered {
		t.Errorf("code 3 = %+v", main.Codes[2])
	}
	if bonus := sectors[1]; bonus.Name != "Бонусные коды" || len(bonus.Codes) != 1 || bonus.Codes[0].Number != 1 {
		t.Errorf("bonus sector = %+v", bonus)
	}
}

func TestClassicEnterCode(t *testing.T) {
	var gotBody url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if cookie := r.Header.Get("Cookie"); cookie != "dozorSiteSession=sess-1" {
			t.Errorf("cookie = %q", cookie)
		}
		r.ParseForm()
		gotBody = r.PostForm
		w.Header().Set("Location", "/e-burg/go/?err=8")
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	engine := newClassic(classicGame(), testEnv(srv))
	result := engine.EnterCode(t.Context(), "др12", nil)

	if !result.Accepted || result.StatusCode != 8 {
		t.Fatalf("result = %+v, want accepted status 8", result)
	}
	// "др12" must reach the engine as windows-1251 bytes.
	if got := gotBody.Get("cod"); got != "\xe4\xf012" {
		t.Errorf("cod = %q, want windows-1251 bytes", got)
	}
	if gotBody.Get("action") != "entcod" {
		t.Errorf("action = %q", gotBody.Get("action"))
	}
}

func TestClassicSpoilerCodeGoesThroughTheUsualForm(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if got := r.PostForm.Get("action"); got != "entcod" {
			t.Errorf("action = %q, want entcod", got)
		}
		if got := r.PostForm.Get("cod"); got != "\xea\xf0\xe8\xef\xf2\xe5\xea\xf1" {
			t.Errorf("cod = %q, want windows-1251 bytes", got)
		}
		w.Header().Set("Location", "?err=55")
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	engine := newClassic(classicGame(), testEnv(srv))
	result := engine.EnterSpoilerCode(t.Context(), "криптекс")
	if !result.Accepted || result.StatusCode != 55 {
		t.Fatalf("result = %+v, want accepted status 55", result)
	}
}

func TestClassicEnterCodePinnedLevel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.PostForm.Get("level") != "7" || r.PostForm.Get("skvoz") != "1" {
			t.Errorf("pinned fields missing: %v", r.PostForm)
		}
		w.Header().Set("Location", "?err=11")
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	engine := newClassic(classicGame(), testEnv(srv))
	pinned := 7
	result := engine.EnterCode(t.Context(), "dr1", &pinned)
	if result.Accepted || result.StatusCode != 11 {
		t.Fatalf("result = %+v, want rejected status 11", result)
	}
}

// TestClassicReloginOnDeadSession: the first load hits an HTML login page,
// the engine re-logs in, persists the new session and retries.
func TestClassicReloginOnDeadSession(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/e-burg/API/login.php":
			if r.URL.Query().Get("login") != "team" || r.URL.Query().Get("password") != "secret" {
				t.Errorf("login params = %q", r.URL.RawQuery)
			}
			w.Write([]byte(`{"code": "2", "userToken": "sess-2"}`))
		case "/e-burg/go/":
			if r.URL.Query().Get("s") == "sess-2" {
				w.Write([]byte(classicPayload))
			} else {
				w.Write([]byte("<html>login page</html>"))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	env := testEnv(srv)
	var updatedSession string
	env.OnSessionUpdate = func(session string) { updatedSession = session }

	engine := newClassic(classicGame(), env)
	snap, err := engine.Load(t.Context())
	if err != nil {
		t.Fatalf("Load after relogin: %v", err)
	}
	if level := snap.LevelNumber(); level == nil || *level != 3 {
		t.Errorf("LevelNumber = %v, want 3", level)
	}
	if updatedSession != "sess-2" {
		t.Errorf("OnSessionUpdate got %q, want sess-2", updatedSession)
	}
}

func TestClassicStart(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, pass, ok := r.BasicAuth(); !ok || user != "" || pass != "" {
			t.Errorf("expected empty basic auth, got %q %q", user, pass)
		}
		w.Write([]byte(`{"code": 2, "userToken": "fresh"}`))
	}))
	defer srv.Close()

	cfg := store.DefaultGameConfig()
	cfg.Login = "team"
	cfg.Password = "secret"

	g, err := Start(t.Context(), cfg, testEnv(srv))
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if g.Session != "fresh" || g.Engine != NameClassic || g.City != "e-burg" {
		t.Errorf("game = %+v", g)
	}
	if len(g.CodeFormats) != 1 || g.CodeFormats[0][0] != "dr" {
		t.Errorf("code formats not copied: %v", g.CodeFormats)
	}
}
