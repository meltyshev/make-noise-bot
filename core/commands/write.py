from .command import Command


class WriteCommand(Command):
    name = 'write'

    def init(self, arguments):
        if not self.is_admin():
            return

        self.set_last_command()
        self.reply_with_message('Какой chat_id?')

    def handle(self, state):
        if state is None:
            self.set_last_command({'chat_id': int(self.message.text)})
            self.reply_with_message('Что отправить?')
            return

        self.delete_last_command()

        self.send_message(state['chat_id'], self.message.text)
        self.reply_with_message('Готово!')
