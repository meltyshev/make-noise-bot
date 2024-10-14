from .command import Command

from core.telegram import KeyboardButton, ReplyKeyboardMarkup, ReplyKeyboardRemove

VALUES = [
    ['Вариант 1', 'Вариант 2'],
    ['Вариант 3', 'Вариант 4'],
    ['Вариант 5', 'Вариант 6']
]


class TestCommand(Command):
    name = 'test'

    def init(self, arguments):
        self._reset()

    def handle(self, state):
        if self.message.text == 'Сбросить':
            return self._reset()

        if self.message.text == 'Закрыть':
            return self._close()

        value = self.message.text
        if value.endswith(' \u2705'):
            value = value[:-2]
            is_checked = True
        else:
            is_checked = False

        if value in [cell for row in VALUES for cell in row]:
            if is_checked:
                del state[value]
            else:
                state[value] = ''

            self.set_last_command(state)
        else:
            self.reply_with_message('Такого варианта нет :(')

        return self._reply_with_menu(state)

    def _reset(self):
        state = {}

        self.set_last_command(state)
        self._reply_with_menu(state)

    def _close(self):
        self.delete_last_command()
        self.reply_with_message('Готово!', reply_markup=ReplyKeyboardRemove())

    def _reply_with_menu(self, state):
        keyboard = []
        for row in VALUES:
            keyboard.append(
                [KeyboardButton('%s%s' % (cell, ' \u2705' if cell in state else '')) for cell in row]
            )

        keyboard.append([KeyboardButton('Сбросить'), KeyboardButton('Закрыть')])

        self.reply_with_message('Выберите варианты:', reply_markup=ReplyKeyboardMarkup(keyboard))
