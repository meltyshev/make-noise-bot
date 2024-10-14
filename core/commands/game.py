from .command import Command

from core.game_engines import get_game_engine, start_game_engine


class GameCommand(Command):
    name = 'game'

    def init(self, arguments):
        if not self.is_manager():
            return

        game_engine = get_game_engine()
        if game_engine is None:
            game_engine = start_game_engine()
            if game_engine is not None:
                self.reply_with_message(game_engine.link)
            else:
                self.reply_with_message('Не могу получить сессию.')
        else:
            game_engine.stop()
            self.reply_with_message('Игра окончена.')
