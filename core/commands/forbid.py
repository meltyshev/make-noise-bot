from .command import Command

from core.models import Chat


class ForbidCommand(Command):
    name = 'forbid'

    def init(self, arguments):
        if not self.is_admin():
            return

        self.set_last_command()
        self.reply_with_message('Какой chat id?')

    def handle(self, state):
        self.delete_last_command()

        chat_id = int(self.message.text)
        chat = Chat.get(chat_id)

        if chat is not None and chat.permission != 'forbidden':
            chat.permission = 'forbidden'
            chat.save()

            self.send_message(chat_id, 'Вам запрещено мной пользоваться!')
            self.reply_with_message('Готово!')
        else:
            self.reply_with_message('Чат не найден.')
