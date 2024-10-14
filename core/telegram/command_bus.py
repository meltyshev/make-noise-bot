class CommandBus:
    def __init__(self, bot, command_classes):
        self._bot = bot
        self._commands = {}
        self._help = []

        for command_class in command_classes:
            self._commands[command_class.name] = command_class

            if command_class.description is not None:
                self._help.append(
                    (command_class.name, command_class.description)
                )

    @property
    def help(self):
        return self._help

    def get_command(self, name, message):
        if name in self._commands:
            return self._commands[name](self._bot, message)
