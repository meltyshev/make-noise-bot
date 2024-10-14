from .command import Command


class TopCommand(Command):
    name = 'top'
    # description = 'кто топ?'

    def init(self, arguments):
        self.reply_with_message('Все топ!')
