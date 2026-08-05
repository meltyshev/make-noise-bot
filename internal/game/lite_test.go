package game

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/meltyshev/make-noise-bot/internal/store"
)

// litePayload mimics the lite.dzzzr.ru page: marker comments, a progress
// line, a single-line sector block with a "nonstandard" sector, hints, and
// bonus levels that must be cut off.
const litePayload = `<html>
<!--levelNumberBegin-->7<!--levelNumberEnd-->
<!--levelTextBegin-->Вопрос <b>тут</b><!--levelTextEnd-->
(Всего - 10, для прохождения достаточно любых 8, принято - 3)
<!--timeOnLevelBegin 3725 timeOnLevelEnd-->
<strong>Коды сложности</strong><br> основные коды: 1, <font>2</font>, null<br>бонусные коды: <br />(1.1) первый<br />(2) второй<br></div>
<!--LevelClue1Text-->Подсказка один<!--LevelClue1TextEnd-->
<!--LevelClue2Text-->Подсказка <b>два</b><!--LevelClue2TextEnd-->
<!--BonusLevels-->
<!--levelNumberBegin-->999<!--levelNumberEnd-->
</html>`

func liteGame() store.Game {
	return store.Game{Engine: NameLite, City: "e-burg", Pincode: "pin-1"}
}

func TestLiteLoadAndSnapshot(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/e-burg/go/" || r.URL.Query().Get("pin") != "pin-1" {
			t.Errorf("unexpected request: %s", r.URL)
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(litePayload))
	}))
	defer srv.Close()

	engine := newLite(liteGame(), testEnv(srv))
	snap, err := engine.Load(context.Background())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// The second levelNumber sits below the BonusLevels cut and must be
	// invisible.
	if level := snap.LevelNumber(); level == nil || *level != 7 {
		t.Errorf("LevelNumber = %v, want 7", level)
	}
	// Every text node is trimmed, so the space before inline tags is lost.
	if got, want := snap.Question(), "Вопрос<b>тут</b>"; got != want {
		t.Errorf("Question = %q, want %q", got, want)
	}
	// 3725 seconds = 01:02:05; 8 codes are enough out of 10.
	if got, want := snap.Progress(), "3/8 01:02:05"; got != want {
		t.Errorf("Progress = %q, want %q", got, want)
	}

	hintNumber, hintText := snap.Hint()
	if hintNumber != 2 || hintText != "Подсказка<b>два</b>" {
		t.Errorf("Hint = (%d, %q)", hintNumber, hintText)
	}

	sectors := snap.Sectors()
	if len(sectors) != 2 {
		t.Fatalf("Sectors = %d, want 2", len(sectors))
	}

	main := sectors[0]
	if main.Name != "Основные коды" || len(main.Codes) != 3 {
		t.Fatalf("main sector = %+v", main)
	}
	// Lite hazards keep the colon part; only null becomes N.
	if main.Codes[0].Hazard != "1" || main.Codes[1].Hazard != "2" || !main.Codes[1].Entered || main.Codes[2].Hazard != "N" {
		t.Errorf("main codes = %+v", main.Codes)
	}

	bonus := sectors[1]
	if bonus.Name != "Бонусные коды" || len(bonus.Codes) != 2 {
		t.Fatalf("bonus sector = %+v", bonus)
	}
	// The nonstandard sector lists hazards in parentheses.
	if bonus.Codes[0].Hazard != "1.1" || bonus.Codes[1].Hazard != "2" {
		t.Errorf("bonus codes = %+v", bonus.Codes)
	}
	if bonus.Codes[0].Number != 1 || bonus.Codes[1].Number != 2 {
		t.Errorf("bonus numbering = %+v", bonus.Codes)
	}
}

func TestLiteEnterCodeMultipart(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "multipart/form-data") {
			t.Errorf("content type = %q, want multipart", ct)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		if got := r.FormValue("cod"); got != "\xe4\xf01" {
			t.Errorf("cod = %q, want windows-1251 bytes", got)
		}
		w.Header().Set("Location", "?err=17")
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	engine := newLite(liteGame(), testEnv(srv))
	result := engine.EnterCode(context.Background(), "др1", nil)
	if !result.Accepted || result.StatusCode != 17 {
		t.Fatalf("result = %+v, want accepted status 17", result)
	}
}

func TestLiteLoadDecodesWindows1251(t *testing.T) {
	// The real site serves windows-1251; the body must be transcoded.
	raw := []byte("<!--levelTextBegin-->\xcf\xf0\xe8\xe2\xe5\xf2<!--levelTextEnd-->")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=windows-1251")
		w.Write(raw)
	}))
	defer srv.Close()

	engine := newLite(liteGame(), testEnv(srv))
	snap, err := engine.Load(context.Background())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := snap.Question(); got != "Привет" {
		t.Errorf("Question = %q, want Привет", got)
	}
}

func TestLiteLoadRejectsRedirect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/somewhere")
		w.WriteHeader(http.StatusFound)
		io.WriteString(w, "redirecting")
	}))
	defer srv.Close()

	engine := newLite(liteGame(), testEnv(srv))
	if _, err := engine.Load(context.Background()); err == nil {
		t.Fatal("Load should fail on a redirect answer")
	}
}
