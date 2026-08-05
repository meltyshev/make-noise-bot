package bot

// Registration order defines the /help listing.
func (a *App) registerCommands() {
	for _, cmd := range []*Command{
		cmdNumbersToLetters(),
		cmdLettersToNumbers(),
		cmdIntersection(),
		cmdMorse(),
		cmdAnagram(),
		cmdMask(),
		cmdLink(),
		cmdQuestion(),
		cmdNotes(),
		cmdRating(),
		cmdClearRating(),
		cmdTop(),
		cmdMaxwell(),
		cmdRomka(),
		cmdCancel(),
		cmdHelp(),
		cmdStart(),
		cmdPermission(),
		cmdCoordinates(),
		cmdChatID(),
		cmdUserID(),
		cmdAvatar(),
		cmdGame(),
		cmdCodeFormats(),
		cmdRestrict(),
		cmdBruteForce(),
		cmdSubscribe(),
		cmdPinLevel(),
		cmdUnpinLevel(),
		cmdGameConfig(),
		cmdChats(),
		cmdAllow(),
		cmdForbid(),
		cmdDrop(),
		cmdWrite(),
		cmdConfig(),
	} {
		a.register(cmd)
	}
}

// askFlow: with arguments the command answers right away, otherwise it asks
// and waits for the next message.
func askFlow(name, description, ask, required string, run func(c *Ctx, input string)) *Command {
	return &Command{
		Name:        name,
		Description: description,
		Init: func(c *Ctx, args string) {
			if !c.EnsureAllowed(name) {
				return
			}
			if args != "" {
				c.DelConv()
				run(c, args)
			} else {
				c.SetConv(name)
				c.Reply(ask)
			}
		},
		Handle: func(c *Ctx, _ any) {
			c.DelConv()
			if c.Text() != "" {
				run(c, c.Text())
			} else {
				c.Reply(required)
			}
		},
	}
}
