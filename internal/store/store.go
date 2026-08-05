// Package store keeps all bot state in memory and mirrors every change to a
// JSON file with atomic writes.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type Store struct {
	mu   sync.Mutex
	path string
	data *Data
}

// Open loads the state file; a missing file starts with defaults.
func Open(path string) (*Store, error) {
	s := &Store{path: path, data: newData()}

	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}
	if err := json.Unmarshal(raw, s.data); err != nil {
		return nil, fmt.Errorf("parse state %s: %w", path, err)
	}

	// Guard against hand-edited files with nulled collections.
	if s.data.Chats == nil {
		s.data.Chats = map[int64]*Chat{}
	}
	if s.data.Players == nil {
		s.data.Players = map[int64]*Player{}
	}
	if s.data.Managers == nil {
		s.data.Managers = []int64{}
	}
	if s.data.UserNames == nil {
		s.data.UserNames = map[int64]string{}
	}
	if s.data.GameConfig.CodeFormats == nil {
		s.data.GameConfig.CodeFormats = [][]string{}
	}
	if s.data.GameConfig.Subscribers == nil {
		s.data.GameConfig.Subscribers = []int64{}
	}
	if s.data.Game != nil && s.data.Game.Subscribers == nil {
		s.data.Game.Subscribers = []int64{}
	}
	return s, nil
}

// View runs f under the lock; f must not retain references.
func (s *Store) View(f func(d *Data)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f(s.data)
}

func (s *Store) Update(f func(d *Data)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	f(s.data)
	return s.persist()
}

func (s *Store) persist() error {
	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	raw = append(raw, '\n')

	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".state-*.json")
	if err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return fmt.Errorf("write state: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	if err := os.Rename(tmp.Name(), s.path); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	return nil
}

// Accessors below return copies.

func (s *Store) Chat(id int64) (Chat, bool) {
	var (
		chat Chat
		ok   bool
	)
	s.View(func(d *Data) {
		if c, found := d.Chats[id]; found {
			chat, ok = *c, true
		}
	})
	return chat, ok
}

func (s *Store) Game() (Game, bool) {
	var (
		game Game
		ok   bool
	)
	s.View(func(d *Data) {
		if d.Game != nil {
			game, ok = copyGame(d.Game), true
		}
	})
	return game, ok
}

func (s *Store) GameConfig() GameConfig {
	var cfg GameConfig
	s.View(func(d *Data) {
		cfg = d.GameConfig
		cfg.CodeFormats = copyFormats(d.GameConfig.CodeFormats)
		cfg.Subscribers = append([]int64{}, d.GameConfig.Subscribers...)
	})
	return cfg
}

func (s *Store) IsManager(userID int64) bool {
	var ok bool
	s.View(func(d *Data) { ok = d.IsManager(userID) })
	return ok
}

func (s *Store) LeaveMode() bool {
	var on bool
	s.View(func(d *Data) { on = d.LeaveMode })
	return on
}

func (s *Store) UpdateGame(f func(g *Game)) error {
	return s.Update(func(d *Data) {
		if d.Game != nil {
			f(d.Game)
		}
	})
}

func (s *Store) IncrementPlayer(userID int64, name string) error {
	return s.Update(func(d *Data) {
		p, ok := d.Players[userID]
		if !ok {
			p = &Player{}
			d.Players[userID] = p
		}
		p.Name = name
		p.Total++
	})
}

// Rating is sorted best first.
func (s *Store) Rating() []Player {
	var players []Player
	s.View(func(d *Data) {
		for _, p := range d.Players {
			players = append(players, *p)
		}
	})
	sort.SliceStable(players, func(i, j int) bool {
		return players[i].Total > players[j].Total
	})
	return players
}

// List fields come out non-nil: a nil slice marshals as JSON null.
func copyGame(g *Game) Game {
	out := *g
	out.CodeFormats = copyFormats(g.CodeFormats)
	out.Subscribers = append([]int64{}, g.Subscribers...)
	out.SolvedSpoilers = append([]int{}, g.SolvedSpoilers...)
	out.LevelNumber = copyInt(g.LevelNumber)
	out.HintNumber = copyInt(g.HintNumber)
	out.PinnedLevel = copyInt(g.PinnedLevel)
	return out
}

func copyFormats(formats [][]string) [][]string {
	out := make([][]string, len(formats))
	for i, f := range formats {
		out[i] = append([]string{}, f...)
	}
	return out
}

func copyInt(v *int) *int {
	if v == nil {
		return nil
	}
	n := *v
	return &n
}
