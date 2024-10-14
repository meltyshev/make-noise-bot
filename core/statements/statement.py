from core.models import Chat
from core.telegram import Statement as BaseStatement


class Statement(BaseStatement):
    def __init__(self, bot, message):
        BaseStatement.__init__(self, bot, message)
        self._chat = None

    def get_chat(self):
        if self._chat and self._chat.key == self.message.chat.id:
            return self._chat

        self._chat = Chat.get(self.message.chat.id)
        return self._chat

    def is_allowed(self):
        chat = self.get_chat()
        if chat is None:
            return False

        return chat.permission == 'allowed'
