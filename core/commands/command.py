from core.models import Chat, Config
from core.telegram import Command as BaseCommand


class Command(BaseCommand):
    def __init__(self, bot, message):
        BaseCommand.__init__(self, bot, message)
        self._chat = None

    def get_chat(self):
        if self._chat and self._chat.key == self.message.chat.id:
            return self._chat

        self._chat = Chat.get(self.message.chat.id)
        return self._chat

    def is_manager(self):
        return self.is_admin() or self.message.from_user.id in Config.get_or_create().managers

    def is_allowed(self):
        chat = self.get_chat()

        if chat is None:
            self.reply_with_message(
                'Для использования команды /%s в этом чате нужно разрешение. /permission - сделать запрос.' % self.name
            )
            return False

        if chat.permission == 'requested':
            self.reply_with_message(
                'Для использования команды /%s в этом чате нужно разрешение. Вы уже сделали запрос, ожидайте решения.' % self.name
            )
        elif chat.permission == 'forbidden':
            self.reply_with_message('Вам запрещено мной пользоваться!')

        return chat.permission == 'allowed'
