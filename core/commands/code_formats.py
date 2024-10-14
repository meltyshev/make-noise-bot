import json

from .command import Command

from core.game_engines import get_game_engine


class CodeFormatsCommand(Command):
    name = 'codeformats'

    def init(self, arguments):
        if not self.is_manager():
            return

        game_engine = get_game_engine()
        if game_engine is not None:
            self.set_last_command()
            self.reply_with_message('Какие форматы кода?')
        else:
            self.reply_with_message('Нет активной игры.')

    def handle(self, state):
        self.delete_last_command()

        text = self.message.text
        if text is not None:
            game_engine = get_game_engine()
            if game_engine is not None:
                try:
                    code_formats = json.loads(text)
                except ValueError:
                    # TODO
                    self.reply_with_message('Ты должен ввести форматы кода!')
                else:
                    # TODO: validation
                    game_engine.set_code_formats(code_formats)
                    self.reply_with_message('Готово!')
            else:
                self.reply_with_message('Нет активной игры.')
        else:
            self.reply_with_message('Ты должен ввести форматы кода!')
