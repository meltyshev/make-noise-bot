from .command import Command


class UserIdCommand(Command):
    name = 'userid'

    def init(self, arguments):
        self.reply_with_message(self.message.from_user.id)
