import requests

from .command import Command


class MaskCommand(Command):
    name = 'mask'
    description = 'слова по маске'

    def init(self, arguments):
        if not self.is_allowed():
            return

        if arguments is not None:
            self.delete_last_command()
            self._reply_with_result(arguments)
        else:
            self.set_last_command()
            self.reply_with_message('По какой маске хочешь получить слова?')

    def handle(self, state):
        self.delete_last_command()
        if self.message.text is not None:
            self._reply_with_result(self.message.text)
        else:
            self.reply_with_message('Ты должен ввести маску!')

    def _reply_with_result(self, letters):
        for letter in letters:
            if not letter.isalpha() and letter not in ('-', '_'):
                self.reply_with_message(
                    'Символы, кроме букв, «-» и «_», запрещены!'
                )
                return

        words = self._get_mask_words(letters)
        if words is not None:
            if words:
                self.reply_with_message(', '.join(words))
            else:
                self.reply_with_message(
                    'По введенной маске слова не найдены \ud83d\ude14'
                )

    def _get_mask_words(self, letters):
        letters = letters.replace('-', '*')
        letters = letters.replace('_', '*')
        letters = letters.lower()

        response = requests.get('http://anagram.poncy.ru/anagram-decoding.cgi', params={
            'name': 'anagram_main',
            'inword': letters,
            'answer_type': 4
        })
        if response.status_code != 200:
            return

        return [word.lower() for word in response.json()['result']]
