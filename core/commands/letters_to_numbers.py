from .command import Command

from core.utils.string_ru import lowercase as lowercase_ru
from string import ascii_lowercase as lowercase_en


class LettersToNumbersCommand(Command):
    name = 'letterstonumbers'
    description = 'буквы в цифры'

    def init(self, arguments):
        if not self.is_allowed():
            return

        if arguments is not None:
            self.delete_last_command()
            self._reply_with_result(arguments)
        else:
            self.set_last_command()
            self.reply_with_message('Какие буквы хочешь перевести?')

    def handle(self, state):
        self.delete_last_command()
        if self.message.text is not None:
            self._reply_with_result(self.message.text)
        else:
            self.reply_with_message('Ты должен ввести буквы!')

    def _reply_with_result(self, letters):
        for word in letters.split():
            numbers = []
            for letter in word:
                letter = letter.lower()
                if letter in lowercase_en:
                    numbers.append(str(lowercase_en.index(letter) + 1))
                elif letter in lowercase_ru:
                    numbers.append(str(lowercase_ru.index(letter) + 1))
                else:
                    numbers.append('(%s)' % letter)

            self.reply_with_message(' '.join(numbers))
