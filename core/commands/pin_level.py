from .command import Command

from core.game_engines import DozorClassic


class PinLevelCommand(Command):
    name = 'pinlevel'

    def init(self, arguments):
        if not self.is_manager():
            return

        game_engine = DozorClassic.get()
        if game_engine is not None:
            self.set_last_command()
            self.reply_with_message('Какой номер уровня?')
        else:
            self.reply_with_message('Нет активной игры.')

    def handle(self, state):
        self.delete_last_command()

        level = self.message.text
        if level is not None and level.isdigit():
            level = int(level)

            game_engine = DozorClassic.get()
            if game_engine is not None:
                game_engine.pin_level(level)
                self.reply_with_message('Готово!')
            else:
                self.reply_with_message('Нет активной игры.')
        else:
            self.reply_with_message('Ты должен ввести номер уровня!')
