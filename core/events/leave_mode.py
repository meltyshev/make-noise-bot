from core.models import Config
from core.telegram import Event


class LeaveModeEvent(Event):
    @staticmethod
    def satisfies(message):
        return message.chat.type in ('group', 'supergroup') and message.text and Config.get_or_create().is_leave_mode

    def handle(self):
        self.leave_chat(self.message.chat.id)

        self.send_message_to_admin(
            'Я покинул %s (%s).' % (self.message.chat.title, self.message.chat.id)
        )
