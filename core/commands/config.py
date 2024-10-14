import json

from .command import Command

from core.models import Config
from core.telegram import KeyboardButton, ReplyKeyboardMarkup, ReplyKeyboardRemove


class ConfigCommand(Command):
    name = 'config'

    def init(self, arguments):
        if not self.is_admin():
            return

        self._reply_with_menu(Config.get_or_create())

    def handle(self, state):
        config = Config.get_or_create()

        if state is None:
            return self._handle_menu_choise(config)

        if self.message.text != 'Отмена':
            if state == 'managers':
                try:
                    config.managers = json.loads(self.message.text)
                except ValueError:
                    pass
                else:
                    config.save()

        self._reply_with_menu(config)

    def _handle_menu_choise(self, config):
        if self.message.text.startswith('Менеджеры'):
            self.set_last_command('managers')
            return self._reply_with_cancel('Менеджеры:')

        if self.message.text.startswith('Режим выхода'):
            config.is_leave_mode = not config.is_leave_mode
            config.save()

            return self._reply_with_menu(config)

        if self.message.text == 'Завершить':
            return self._close()

        if self.message.text == 'Сбросить':
            config = Config.reset()

        self._reply_with_menu(config=config)

    def _reply_with_menu(self, config):
        self.set_last_command()
        self.reply_with_message(
            'Настройки:',
            reply_markup=ReplyKeyboardMarkup([
                [KeyboardButton('Менеджеры: %s' % json.dumps(config.managers))],
                [KeyboardButton('Режим выхода: %s' % ('включен' if config.is_leave_mode else 'выключен'))],
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

    def _close(self):
        self.delete_last_command()
        self.reply_with_message('Готово!', reply_markup=ReplyKeyboardRemove())
