package game

import (
	"net/http/httptest"

	"github.com/meltyshev/make-noise-bot/internal/store"
)

func testEnv(srv *httptest.Server) *Env {
	env := DefaultEnv()
	env.ClassicBaseURL = srv.URL
	env.LiteBaseURL = srv.URL
	return env
}

func classicGame() store.Game {
	return store.Game{
		Engine:   NameClassic,
		City:     "e-burg",
		Login:    "team",
		Password: "secret",
		Session:  "sess-1",
	}
}
