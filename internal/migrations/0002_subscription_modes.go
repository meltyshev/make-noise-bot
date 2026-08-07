package migrations

// Subscriptions used to carry a flag per update kind. They become a single
// mode: a chat that wanted the level texts keeps everything, the rest keep
// the short notices.
func init() {
	register(2, "subscription modes", func(state map[string]any) error {
		for _, key := range []string{"game_config", "game"} {
			owner, ok := object(state, key)
			if !ok {
				continue
			}

			subscriptions, ok := owner["subscriptions"].([]any)
			if !ok {
				continue
			}

			for _, item := range subscriptions {
				sub, ok := item.(map[string]any)
				if !ok {
					continue
				}

				wantedTexts := sub["question"] == true || sub["notes"] == true
				for _, kind := range []string{"level_up", "hints", "spoilers", "question", "notes"} {
					delete(sub, kind)
				}
				if !wantedTexts {
					sub["events_only"] = true
				}
			}
		}
		return nil
	})
}
