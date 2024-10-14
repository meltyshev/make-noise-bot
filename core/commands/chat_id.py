from .command import Command


class ChatIdCommand(Command):
    name = 'chatid'

    def init(self, arguments):
        self.reply_with_message(self.message.chat.id)
