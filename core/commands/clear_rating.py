from .command import Command

from core.models import Player


class ClearRatingCommand(Command):
    name = 'clearrating'

    def init(self, arguments):
        if not self.is_manager():
            return

        Player.clear()

        self.reply_with_message('Рейтинг очищен.')
