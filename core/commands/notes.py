from .command import Command

from core.game_engines import DozorClassic


class NotesCommand(Command):
    name = 'notes'
    description = 'примечания'

    def init(self, arguments):
        if not self.is_allowed():
            return

        game_engine = DozorClassic.get()
        if game_engine is None:
            self.reply_with_message('Нет активной игры.')
            return

        if not game_engine.load_game_data():
            self.reply_with_message('Не могу загрузить движок.')
            return

        notes = game_engine.get_notes()
        if notes:
            self.reply_with_message(notes, parse_mode='HTML')
        else:
            self.reply_with_message('Примечаний нет.')
