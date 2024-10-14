import re

from core.models import Game

REQUEST_TIMEOUT = 22
USER_AGENT = 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36'

REQUEST_TIMEOUT_MESSAGE = 'Движок не отвечает.'
# INVALID_PINCODE_MESSAGE = 'Неверная пинка.'
UNKNOWN_ERROR_MESSAGE = 'Не понимаю ответ движка.'

BETWEEN_TAGS_REGEX = re.compile(r'(?<=>).*?(?=<)', re.DOTALL)


class EnterCodeResult:
    def __init__(self, message, status_code=None, is_accepted=False):
        self.message = message
        self.status_code = status_code
        self.is_accepted = is_accepted


class GameEngine():
    @staticmethod
    def start(game_engine_class, config, **kwargs):
        return game_engine_class(Game.create(
            engine=game_engine_class.name,
            city=config.city,
            code_formats=config.code_formats,
            subscribers=config.subscribers,
            **kwargs
        ))

    @staticmethod
    def get(game_engine_class):
        game = Game.get()
        if game is None:
            return

        if game.engine == game_engine_class.name:
            return game_engine_class(game)

    def __init__(self, game):
        self.game = game

    @property
    def subscribers(self):
        return self.game.subscribers

    @property
    def is_restricted(self):
        return self.game.is_restricted

    @property
    def level_number(self):
        return self.game.level_number

    @property
    def hint_number(self):
        return self.game.hint_number

    @property
    def solved_spoilers(self):
        return self.game.solved_spoilers

    def set_code_formats(self, code_formats):
        self.game.code_formats = code_formats
        self.game.save()

    def subscribe(self, subscriber):
        self.game.modify(add_to_set__subscribers=[subscriber])

    def unsubscribe(self, subscriber):
        self.game.modify(pull__subscribers=subscriber)

    def restrict(self):
        self.game.is_restricted = not self.game.is_restricted
        self.game.save()

        return self.game.is_restricted

    def set_level_number(self, level_number, hint_number=None, solved_spoilers=[]):
        self.game.level_number = level_number
        self.game.hint_number = hint_number
        self.game.solved_spoilers = solved_spoilers
        self.game.save()

    def set_hint_number(self, hint_number):
        self.game.hint_number = hint_number
        self.game.save()

    def set_solved_spoilers(self, solved_spoilers):
        self.game.solved_spoilers = solved_spoilers
        self.game.save()

    def stop(self):
        self.game.delete()

    def prepare_code(self, code):
        code = code.lower()

        has_mark = code[0] in ('!', '.')
        if has_mark:
            return code[1:]

        non_digits = ''.join(char for char in code if not char.isdigit())
        for code_format in self.game.code_formats:
            first_code_format = code_format[0]
            if first_code_format == non_digits:
                return code

            for code_format_replacement in code_format[1:]:
                if code_format_replacement == non_digits:
                    for search_char, replace_char in zip(code_format_replacement, first_code_format):
                        code = code.replace(search_char, replace_char, 1)
                    return code

    def enter_code(self, code, **kwargs):
        raise NotImplementedError

    def load_game_data(self):
        raise NotImplementedError

    def get_level_number(self):
        raise NotImplementedError

    def get_progress(self):
        raise NotImplementedError

    def get_question(self):
        raise NotImplementedError

    def get_notes(self):
        raise NotImplementedError

    def get_sectors(self):
        raise NotImplementedError

    def get_hint(self):
        raise NotImplementedError

    def get_solved_spoilers(self):
        raise NotImplementedError
