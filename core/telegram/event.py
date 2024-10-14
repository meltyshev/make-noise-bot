from .interaction import Interaction


class Event(Interaction):
    def __init__(self, bot, message):
        Interaction.__init__(self, bot, message)

    @staticmethod
    def satisfies(message):
        raise NotImplementedError

    def handle(self):
        raise NotImplementedError

    def reply_with_message(self, *args, **kwargs):
        return self.send_message(self.message.chat.id, *args, **kwargs)
