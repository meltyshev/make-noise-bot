from .command import Command

from core.models import Spot


class DeleteSpotCommand(Command):
    name = 'deletespot'

    def init(self, arguments):
        if not self.is_manager():
            return

        self.set_last_command()
        self.reply_with_message('Название точки?')

    def handle(self, state):
        self.delete_last_command()

        spot = Spot.get(self.message.text)
        if spot is not None:
            spot.delete()
            self.reply_with_message('Готово!')
        else:
            self.reply_with_message('Точка не найдена.')
