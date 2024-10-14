from .command import Command

from core import heroku
from core.game_engines import get_game_engine


class BroadcastCommand(Command):
    name = 'broadcast'

    def init(self, arguments):
        if not self.is_manager():
            return

        clock = heroku.get_clock()
        if clock.quantity == 1:
            clock.scale(0)

            self.reply_with_message('Обновления остановлены.')
            return

        game_engine = get_game_engine()
        if game_engine is None:
            self.reply_with_message('Нет активной игры.')
            return

        if not game_engine.load_game_data():
            self.reply_with_message('Не могу загрузить движок.')
            return

        level_number = game_engine.get_level_number()
        hint_number = game_engine.get_hint()[0]
        solved_spoilers = game_engine.get_solved_spoilers() or []

        game_engine.set_level_number(level_number, hint_number, solved_spoilers)
        clock.scale(1)

        self.reply_with_message('Обновления запущены.')
