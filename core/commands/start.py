from .command import Command


class StartCommand(Command):
    name = 'start'

    def init(self, arguments):
        self.reply_with_message(
            'Привет, %s!' % self.message.from_user.first_name
        )
