package migrations

// Subscribers used to be a plain list of chat ids where the first one also
// received the level text and notes. They become per-chat subscriptions with
// the same effective behaviour: the first chat keeps everything, the rest
// keep notifications only.
func init() {
	register(1, "subscription kinds", func(state map[string]any) error {
		for _, key := range []string{"game_config", "game"} {
			owner, ok := object(state, key)
			if !ok {
				continue
			}

			ids, ok := owner["subscribers"].([]any)
			delete(owner, "subscribers")
			if !ok {
				continue
			}

			subscriptions := []any{}
			for i, rawID := range ids {
				id, ok := rawID.(float64)
				if !ok {
					continue
				}
				subscriptions = append(subscriptions, map[string]any{
					"chat_id":  id,
					"level_up": true,
					"hints":    true,
					"spoilers": true,
					"question": i == 0,
					"notes":    i == 0,
				})
			}
			owner["subscriptions"] = subscriptions
		}
		return nil
	})
}
