package game

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/meltyshev/make-noise-bot/internal/store"
)

const prequelPayload = `<html><strong id=orang>Код сложности:<br>Игра: 1.1, <span>2</span>, null<br>Бонус: 3.3<br></strong></html>`

func prequelGame(name string) store.Game {
	return store.Game{
		Engine:  name,
		City:    "e-burg",
		Login:   "team",
		Session: "sess-1",
		GameID:  "42",
		League:  "1",
	}
}

func TestPrequelSectors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(prequelPayload))
	}))
	defer srv.Close()

	engine := newPrequel(prequelGame(NameClassicPrequel), testEnv(srv))
	snap, err := engine.Load(context.Background())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Prequels report a constant level 0.
	if level := snap.LevelNumber(); level == nil || *level != 0 {
		t.Errorf("LevelNumber = %v, want 0", level)
	}

	sectors := snap.Sectors()
	if len(sectors) != 2 {
		t.Fatalf("Sectors = %d, want 2", len(sectors))
	}
	// Prequel sector names stay raw: no trimming, no capitalization.
	if sectors[0].Name != "Игра" || sectors[1].Name != "Бонус" {
		t.Errorf("sector names = %q, %q", sectors[0].Name, sectors[1].Name)
	}
	// One shared counter across all prequel sectors.
	if sectors[0].Codes[2].Number != 3 || sectors[1].Codes[0].Number != 4 {
		t.Errorf("numbering = %+v %+v", sectors[0].Codes, sectors[1].Codes)
	}
	if sectors[0].Codes[1].Hazard != "2" || !sectors[0].Codes[1].Entered {
		t.Errorf("entered code = %+v", sectors[0].Codes[1])
	}
	if sectors[0].Codes[2].Hazard != "N" {
		t.Errorf("null hazard = %+v", sectors[0].Codes[2])
	}
}

func TestPrequelEnterCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.PostForm.Get("action") != "prequel_code_new" || r.PostForm.Get("league") != "1" || r.PostForm.Get("game") != "42" {
			t.Errorf("form = %v", r.PostForm)
		}
		w.Header().Set("Location", "?err=54")
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	engine := newPrequel(prequelGame(NameLitePrequel), testEnv(srv))
	result := engine.EnterCode(context.Background(), "code", nil)
	if !result.Accepted || result.StatusCode != 54 {
		t.Fatalf("result = %+v, want accepted status 54", result)
	}
}

func TestPrequelLinks(t *testing.T) {
	env := DefaultEnv()

	classic := newPrequel(prequelGame(NameClassicPrequel), env)
	if classic.Link() != "https://classic.dzzzr.ru/e-burg/?section=anons&league=1" {
		t.Errorf("classic prequel link = %q", classic.Link())
	}

	lite := newPrequel(prequelGame(NameLitePrequel), env)
	if lite.Link() != "https://lite.dzzzr.ru/e-burg/?league=1" {
		t.Errorf("lite prequel link = %q", lite.Link())
	}
}

func TestStartUnknownEngine(t *testing.T) {
	cfg := store.DefaultGameConfig()
	cfg.Engine = "Nonsense"
	if _, err := Start(context.Background(), cfg, DefaultEnv()); err == nil {
		t.Fatal("Start should reject unknown engines")
	}
	if engine := New(store.Game{Engine: "Nonsense"}, DefaultEnv()); engine != nil {
		t.Fatal("New should return nil for unknown engines")
	}
}
