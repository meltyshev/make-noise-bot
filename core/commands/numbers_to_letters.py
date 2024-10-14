from .command import Command

from core.utils.string_ru import lowercase as lowercase_ru
from string import ascii_lowercase as lowercase_en


# TODO: format
class NumbersToLettersCommand(Command):
    name = 'numberstoletters'
    description = 'цифры в буквы'

    def init(self, arguments):
        if not self.is_allowed():
            return

        if arguments is not None:
            self.delete_last_command()
            self._reply_with_result(arguments)
        else:
            self.set_last_command()
            self.reply_with_message('Какие числа хочешь перевести?')

    def handle(self, state):
        self.delete_last_command()
        if self.message.text is not None:
            self._reply_with_result(self.message.text)
        else:
            self.reply_with_message('Ты должен ввести цифры!')

    def _reply_with_result(self, numbers):
        word_en = ''
        word_ru = ''

        for number in numbers.split():
            if number.isdigit() and len(number) <= 2:
                number = int(number)

                word_en += lowercase_en[number - 1] if number <= 26 else '(%s)' % number
                word_ru += lowercase_ru[number - 1] if number <= 33 else '(%s)' % number
            else:
                word_en += '(%s)' % number
                word_ru += '(%s)' % number

        self.reply_with_message(word_en)
        self.reply_with_message(word_ru)
