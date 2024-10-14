from .command import Command

from core.game_engines import DozorClassic


class UnpinLevelCommand(Command):
    name = 'unpinlevel'

    def init(self, arguments):
        if not self.is_manager():
            return

        game_engine = DozorClassic.get()
        if game_engine is not None:
            game_engine.pin_level(None)
            self.reply_with_message('Готово!')
        else:
            self.reply_with_message('Нет активной игры.')
