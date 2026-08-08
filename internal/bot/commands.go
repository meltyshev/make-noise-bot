package bot

// registerCommands registers every command in the order /help lists them.
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

// askFlow answers right away when the command has arguments, and otherwise
// asks and waits for the next message.
func askFlow(name, description, ask, required string, run func(c *Ctx, input string)) *Command {
	return &Command{
		Name:        name,
		Description: description,
		Init: func(c *Ctx, args string) {
			if !c.EnsureAllowed() {
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
