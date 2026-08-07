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
	if got, want := snap.Question(), "Вопрос <b>тут</b>"; got != want {
		t.Errorf("Question = %q, want %q", got, want)
	}
	// 3725 seconds = 01:02:05; 8 codes are enough out of 10.
	if got, want := snap.Progress(), "3/8 01:02:05"; got != want {
		t.Errorf("Progress = %q, want %q", got, want)
	}

	hintNumber, hintText := snap.Hint()
	if hintNumber != 2 || hintText != "Подсказка <b>два</b>" {
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

// The fixtures below are the spoiler markup of a real lite page.
const (
	liteOpenSpoiler = `<!--levelTextBegin--><p>Задание</p><div class=spoiler><strong>Примечания к заданию</strong>: ` +
		`<!--taskNotes--><p>ФО: слово</p><!--taskNotesEnd--></div><br/><!--levelTextEnd-->` +
		`<div style='font-size:120%;'><div class=spoiler><div class=title style='padding-left:0'>Спойлер</div>` +
		`<p>55.058638, 82.974920</p>` + "\n" + `<p><img src="http://classic.dzzzr.ru/uploaded/x.png" alt="" /></p></div></div>` +
		`<!--bonusCodeCount 1--><!--mainCodeCount 6-->`

	liteClosedSpoiler = `<!--levelTextEnd--><p>В этом задании есть спойлер № 1. Чтобы увидеть его введите специальный код.` +
		`<form  method=post data-ajax='false'><input type=hidden name=action value=spoilerCode>` +
		`<input type=text size=30 placeholder='код спойлера' id=spoilerCode name=spoilerCode>` +
		`<input id=spoilerCodeBtn type=submit value='показать спойлер'></form></p><!--bonusCodeCount 1-->`
)

func TestLiteSpoilerOpen(t *testing.T) {
	snap := &liteSnapshot{link: "https://lite.dzzzr.ru/e-burg/go/", data: liteOpenSpoiler}

	spoilers := snap.Spoilers()
	if len(spoilers) != 1 {
		t.Fatalf("spoilers = %d, want 1", len(spoilers))
	}
	if !spoilers[0].Open || spoilers[0].Number != 1 {
		t.Errorf("spoiler = %+v, want open number 1", spoilers[0])
	}
	if !strings.Contains(spoilers[0].Text, "55.058638, 82.974920") {
		t.Errorf("text lost the coordinates: %q", spoilers[0].Text)
	}
	if !strings.Contains(spoilers[0].Text, "x.png") {
		t.Errorf("text lost the image: %q", spoilers[0].Text)
	}
	// The notes carry their own spoiler class inside the level text.
	if strings.Contains(spoilers[0].Text, "Примечания") {
		t.Errorf("task notes leaked into the spoiler: %q", spoilers[0].Text)
	}
	if strings.Contains(spoilers[0].Text, "Спойлер") {
		t.Errorf("title leaked into the text: %q", spoilers[0].Text)
	}
}

func TestLiteSpoilerClosed(t *testing.T) {
	snap := &liteSnapshot{link: "https://lite.dzzzr.ru/e-burg/go/", data: liteClosedSpoiler}

	spoilers := snap.Spoilers()
	if len(spoilers) != 1 {
		t.Fatalf("spoilers = %d, want 1", len(spoilers))
	}
	if spoilers[0].Open || spoilers[0].Number != 1 || spoilers[0].Text != "" {
		t.Errorf("spoiler = %+v, want closed number 1 without text", spoilers[0])
	}
}

func TestLiteSpoilersMixed(t *testing.T) {
	data := `<!--levelTextEnd-->` +
		`<div style='font-size:120%;'><div class=spoiler><div class=title>Спойлер</div><p>первый</p></div></div>` +
		`<p>В этом задании есть спойлер № 2. Чтобы увидеть его введите специальный код.<form></form></p>` +
		`<!--bonusCodeCount 1-->`

	spoilers := (&liteSnapshot{link: "https://lite.dzzzr.ru/e-burg/go/", data: data}).Spoilers()
	if len(spoilers) != 2 {
		t.Fatalf("spoilers = %+v, want 2", spoilers)
	}
	if !spoilers[0].Open || spoilers[0].Number != 1 || spoilers[0].Text != "первый" {
		t.Errorf("first = %+v", spoilers[0])
	}
	if spoilers[1].Open || spoilers[1].Number != 2 {
		t.Errorf("second = %+v", spoilers[1])
	}
}

func TestLiteWithoutSpoilers(t *testing.T) {
	data := `<!--levelTextEnd--><!--bonusCodeCount 1--><!--mainCodeCount 6-->`
	if spoilers := (&liteSnapshot{data: data}).Spoilers(); len(spoilers) != 0 {
		t.Errorf("spoilers = %+v, want none", spoilers)
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
