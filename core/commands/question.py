from .command import Command

from core.game_engines import get_game_engine


class QuestionCommand(Command):
    name = 'question'
    description = 'задание'

    def init(self, arguments):
        if not self.is_allowed():
            return

        game_engine = get_game_engine()
        if game_engine is None:
            self.reply_with_message('Нет активной игры.')
            return

        if not game_engine.load_game_data():
            self.reply_with_message('Не могу загрузить движок.')
            return

        question = game_engine.get_question()
        if question:
            self.reply_with_message(question, parse_mode='HTML')
            '''import json
            self.send_message_to_admin(json.dumps([question]))'''
        else:
            self.reply_with_message('Задания нет.')
