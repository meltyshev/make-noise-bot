package updater

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	tgbot "github.com/go-telegram/bot"

	"github.com/meltyshev/make-noise-bot/internal/game"
	"github.com/meltyshev/make-noise-bot/internal/store"
)

// Lite uses class=spoiler for task notes and for both closed and open
// spoilers. Only opening a spoiler should trigger its announcement.
const (
	liteSpoilerTask = `<!--errorText--><p><strong>Код к спойлеру принят</strong></p><!--errorTextEnd-->` +
		`<!--levelNumberBegin-->3<!--levelNumberEnd-->` +
		`<!--levelTextBegin--><p>Задание</p><p><img src="../../uploaded/task.png" alt="" /></p>` +
		`<div class=spoiler><strong>Примечания к заданию</strong>: ` +
		`<!--taskNotes--><p>ФО: слово</p><!--taskNotesEnd--></div><br/><!--levelTextEnd-->`
	liteSpoilerCodes = `<!--bonusCodeCount 0--><!--mainCodeCount 2-->` +
		`<div class='dcodes'><strong>Коды сложности</strong><br>Сектор 1: основные коды: 1+, 1+<br>` +
		` (Всего - 2 , для прохождения достаточно любых 2 , принято - 0) <br /></p></div>`
	liteClosedSpoilerPage = liteSpoilerTask + `<div class="spoiler"><div class="title">Спойлер №1</div>` +
		`<p>Чтобы открыть спойлер, введите специальный код.</p><form method="post" data-ajax="false">` +
		`<input type="hidden" name="action" value="spoilerCode"><input type="hidden" name="pin" value="pin-1">` +
		`<input type="hidden" name="nomessage" value=""><input type="text" name="spoilerCode">` +
		`<input type="submit" value="Открыть спойлер"></form></div>` +
		liteSpoilerCodes + `<!--timeOnLevelBegin 1 timeOnLevelEnd-->`
	liteOpenSpoilerPage = liteSpoilerTask + `<div class="spoiler"><div class="title">Спойлер №1</div>` +
		`<p>55.058638, 82.974920</p>` + "\n" +
		`<p><img src="../../uploaded/x.png" alt="" /></p>` + "\n" +
		`<p><img src="../../uploaded/y.jpg" alt="" /></p></div>` +
		liteSpoilerCodes + `<!--timeOnLevelBegin 111 timeOnLevelEnd-->`
)

func TestLiteSpoilerBroadcast(t *testing.T) {
	type message struct{ chatID, text string }
	var (
		mu   sync.Mutex
		page = liteClosedSpoilerPage
		sent []message
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch r.URL.Path {
		case "/e-burg/go/":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			io.WriteString(w, page)
		case "/bottest/sendMessage":
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				t.Errorf("ParseMultipartForm: %v", err)
				http.Error(w, "bad form", http.StatusBadRequest)
				return
			}
			defer r.MultipartForm.RemoveAll()
			sent = append(sent, message{chatID: r.FormValue("chat_id"), text: r.FormValue("text")})
			io.WriteString(w, `{"ok":true,"result":{"message_id":1}}`)
		default:
			t.Errorf("unexpected request: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	statePath := filepath.Join(t.TempDir(), "state.json")
	st, err := store.Open(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Update(func(d *store.Data) {
		d.Game = &store.Game{
			Engine:  game.NameLite,
			City:    "e-burg",
			Pincode: "pin-1",
			Subscriptions: store.Subscriptions{
				store.AllUpdates(-100), store.OnlyEvents(-200),
			},
		}
	}); err != nil {
		t.Fatal(err)
	}
	tg, err := tgbot.New("test", tgbot.WithSkipGetMe(), tgbot.WithServerURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	env := &game.Env{LiteBaseURL: srv.URL, HTTP: srv.Client()}
	u := New(st, env, tg, slog.New(slog.DiscardHandler), time.Second, func(err error) {
		t.Errorf("updater: %v", err)
	})

	// Each poll continues from the state saved by the previous one.
	tests := []struct {
		name       string
		page       string
		want       []string
		wantSolved []int
		restart    bool
	}{
		{
			name: "closed spoiler and task notes are not announced",
			page: liteClosedSpoilerPage,
			want: []string{"АП!"},
		},
		{
			name: "still closed",
			page: liteClosedSpoilerPage,
		},
		{
			name:       "opened spoiler",
			page:       liteOpenSpoilerPage,
			want:       []string{"Спойлер 1 - АП!"},
			wantSolved: []int{1},
		},
		{
			name:       "same page again",
			page:       liteOpenSpoilerPage,
			wantSolved: []int{1},
		},
		{
			name:       "after restart",
			page:       liteOpenSpoilerPage,
			wantSolved: []int{1},
			restart:    true,
		},
		{
			name: "same spoiler number on next level",
			page: strings.ReplaceAll(strings.ReplaceAll(liteOpenSpoilerPage,
				"<!--levelNumberBegin-->3", "<!--levelNumberBegin-->4"),
				"<!--timeOnLevelBegin 111", "<!--timeOnLevelBegin 1"),
			want:       []string{"АП!", "Спойлер 1 - АП!"},
			wantSolved: []int{1},
		},
	}
	for _, tt := range tests {
		if !t.Run(tt.name, func(t *testing.T) {
			if tt.restart {
				st, err := store.Open(statePath)
				if err != nil {
					t.Fatal(err)
				}
				u.store = st
			}
			mu.Lock()
			page = tt.page
			mu.Unlock()
			u.tick(t.Context())
			mu.Lock()
			messages := slices.Clone(sent)
			sent = nil
			mu.Unlock()

			g, ok := u.store.Game()
			if !ok {
				t.Fatal("Game() = none, want an active game")
			}
			if !slices.Equal(g.SolvedSpoilers, tt.wantSolved) {
				t.Errorf("SolvedSpoilers = %v, want %v", g.SolvedSpoilers, tt.wantSolved)
			}
			if got := len(messages); got != 2*len(tt.want) {
				t.Fatalf("messages = %d, want %d", got, 2*len(tt.want))
			}
			for i, notice := range tt.want {
				full, event := messages[2*i], messages[2*i+1]
				if full.chatID != "-100" || !strings.HasPrefix(full.text, notice+"\n\n") {
					t.Errorf("full message = %+v, want task or spoiler in chat -100", full)
				}
				if event != (message{chatID: "-200", text: notice}) {
					t.Errorf("event message = %+v, want %q in chat -200", event, notice)
				}
				if notice == "Спойлер 1 - АП!" {
					for _, part := range []string{
						"55.058638, 82.974920",
						srv.URL + "/uploaded/x.png",
						srv.URL + "/uploaded/y.jpg",
					} {
						if !strings.Contains(full.text, part) {
							t.Errorf("spoiler text = %q, want %q", full.text, part)
						}
					}
					if strings.Contains(full.text, "Примечания") || strings.Contains(full.text, "Спойлер №") {
						t.Errorf("spoiler text = %q, want no task notes or repeated title", full.text)
					}
				}
			}
		}) {
			return
		}
	}
}
