from .command import Command

from core.game_engines import get_game_engine


class LinkCommand(Command):
    name = 'link'
    description = 'ссылка'

    def init(self, arguments):
        if not self.is_allowed():
            return

        game_engine = get_game_engine()
        if game_engine is not None:
            self.reply_with_message(game_engine.link)
        else:
            self.reply_with_message('Нет активной игры.')
