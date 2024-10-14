from .command import Command


class HelpCommand(Command):
    name = 'help'
    description = 'помощь'

    def init(self, arguments):
        lines = []
        for name, description in self.help:
            lines.append('/%s - %s' % (name, description))

        self.reply_with_message('\n'.join(lines))
