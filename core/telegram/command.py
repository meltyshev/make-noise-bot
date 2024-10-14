from .interaction import Interaction
from .last_command import LastCommand


class Command(Interaction):
    name = None
    description = None

    def __init__(self, bot, message):
        Interaction.__init__(self, bot, message)

    @property
    def help(self):
        return self._bot.help

    def init(self, arguments):
        raise NotImplementedError

    def handle(self, state):
        raise NotImplementedError

    def get_last_command(self):
        return self._bot.get_last_command(self.message.from_user.id, self.message.chat.id)

    def set_last_command(self, state=None):
        self._bot.set_last_command(
            self.message.from_user.id,
            self.message.chat.id,
            LastCommand(self.name, state)
        )

    def delete_last_command(self):
        self._bot.delete_last_command(
            self.message.from_user.id,
            self.message.chat.id
        )

    def is_admin(self):
        return self.message.from_user.id == self._bot.admin_id
