from .command import Command

from core.game_engines import get_game_engine


class RestrictCommand(Command):
    name = 'restrict'

    def init(self, arguments):
        if not self.is_manager():
            return

        game_engine = get_game_engine()
        if game_engine is not None:
            if game_engine.restrict():
                self.reply_with_message('Ввод кодов ограничен.')
            else:
                self.reply_with_message('Ограничение на ввод кодов снято.')
        else:
            self.reply_with_message('Нет активной игры.')
