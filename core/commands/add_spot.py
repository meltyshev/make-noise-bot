from .command import Command

from core.models import Spot


# TODO: check spot uniquness
class AddSpotCommand(Command):
    name = 'addspot'

    def init(self, arguments):
        if not self.is_manager():
            return

        self.set_last_command()
        self.reply_with_message('Название точки?')

    def handle(self, state):
        if state is None:
            self.set_last_command({'name': self.message.text})
            self.reply_with_message('Адрес/координаты точки?')
            return

        self.delete_last_command()
        Spot.create(key=state['name'], location=self.message.text)

        self.reply_with_message('Готово!')
