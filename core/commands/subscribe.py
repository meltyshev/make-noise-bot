from .command import Command

from core.game_engines import get_game_engine


class SubscribeCommand(Command):
    name = 'subscribe'

    def init(self, arguments):
        if not self.is_manager():
            return

        game_engine = get_game_engine()
        if game_engine is None:
            self.reply_with_message('Нет активной игры.')
            return

        chat_id = self.message.chat.id
        if chat_id not in game_engine.subscribers:
            game_engine.subscribe(chat_id)
            self.reply_with_message('Подписка активирована.')
        else:
            game_engine.unsubscribe(chat_id)
            self.reply_with_message('Подписка отменена.')
