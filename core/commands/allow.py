from .command import Command

from core.models import Chat


class AllowCommand(Command):
    name = 'allow'

    def init(self, arguments):
        if not self.is_admin():
            return

        self.set_last_command()
        self.reply_with_message('Какой chat id?')

    def handle(self, state):
        self.delete_last_command()

        chat_id = int(self.message.text)
        chat = Chat.get(chat_id)

        if chat is not None and chat.permission != 'allowed':
            chat.permission = 'allowed'
            chat.save()

            self.send_message(chat_id, 'Разрешение получено \ud83d\ude0d')
            self.reply_with_message('Готово!')
        else:
            self.reply_with_message('Чат не найден.')
