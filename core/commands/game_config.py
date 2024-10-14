import json

from .command import Command

from core.models import GameConfig
from core.telegram import KeyboardButton, ReplyKeyboardMarkup, ReplyKeyboardRemove


# TODO: refactoring
class GameConfigCommand(Command):
    name = 'gameconfig'

    def init(self, arguments):
        if not self.is_manager():
            return

        self.set_last_command()
        self._reply_with_menu(GameConfig.get_or_create())

    def handle(self, state):
        text = self.message.text
        if text is None:
            self.reply_with_message('Ты должен ответить текстом!')
            return

        if state is not None:
            config = GameConfig.get_or_create()
            if text != 'Отмена':
                if state == 'code_formats':  # TODO: validation
                    try:
                        config.code_formats = json.loads(text)
                    except ValueError:
                        pass
                    else:
                        config.save()
                elif state == 'subscribers':  # TODO: validation
                    try:
                        config.subscribers = json.loads(text)
                    except ValueError:
                        pass
                    else:
                        config.save()
                # TODO
                elif state != 'engine' or text in ('DozorClassic', 'DozorLite', 'DozorClassicPrequel', 'DozorLitePrequel'):
                    setattr(config, state, text)
                    config.save()

            self.set_last_command()
            self._reply_with_menu(config)
            return

        if text.startswith('Движок'):
            self.set_last_command('engine')
            self.reply_with_message(
                'Движок:',
                reply_markup=ReplyKeyboardMarkup([
                    [KeyboardButton('DozorClassic')],
                    [KeyboardButton('DozorLite')],
                    [KeyboardButton('DozorClassicPrequel')],
                    [KeyboardButton('DozorLitePrequel')],
                    [KeyboardButton('Отмена')]
                ])
            )
        elif text.startswith('Город'):
            self.set_last_command('city')
            self._reply_with_cancel('Город:')
        elif text.startswith('Логин'):
            self.set_last_command('login')
            self._reply_with_cancel('Логин:')
        elif text.startswith('Пароль'):
            self.set_last_command('password')
            self._reply_with_cancel('Пароль:')
        elif text.startswith('Пинка'):
            self.set_last_command('pincode')
            self._reply_with_cancel('Пинка:')
        elif text.startswith('Номер игры'):
            self.set_last_command('game_id')
            self._reply_with_cancel('Номер игры:')
        elif text.startswith('Лига'):
            self.set_last_command('league')
            self._reply_with_cancel('Лига:')
        elif text.startswith('Форматы кода'):
            self.set_last_command('code_formats')
            self._reply_with_cancel('Форматы кода:')
        elif text.startswith('Подписчики'):
            self.set_last_command('subscribers')
            self._reply_with_cancel('Подписчики:')
        elif text == 'Сбросить':
            config = GameConfig.reset()
            self._reply_with_menu(config)
        elif text == 'Завершить':
            self.delete_last_command()
            self.reply_with_message('Готово!', reply_markup=ReplyKeyboardRemove())

    def _reply_with_menu(self, config):
        self.reply_with_message(
            'Настройки:',
            reply_markup=ReplyKeyboardMarkup([
                [KeyboardButton('Движок: %s' % config.engine)],
                [KeyboardButton('Город: %s' % config.city)],
                [KeyboardButton('Логин: %s' % config.login)],
                [KeyboardButton('Пароль: %s' % config.password)],
                [KeyboardButton('Пинка: %s' % config.pincode)],
                [KeyboardButton('Номер игры: %s' % config.game_id)],
                [KeyboardButton('Лига: %s' % config.league)],
                [KeyboardButton('Форматы кода: %s' % json.dumps(config.code_formats, ensure_ascii=False))],
                [KeyboardButton('Подписчики: %s' % json.dumps(config.subscribers))],
                [KeyboardButton('Сбросить'), KeyboardButton('Завершить')]
            ])
        )

    def _reply_with_cancel(self, message):
        self.reply_with_message(
            message,
            reply_markup=ReplyKeyboardMarkup([
                [KeyboardButton('Отмена')]
            ])
        )
