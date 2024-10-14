import json
import re
import requests

from .command import Command

from alphabet_detector import AlphabetDetector

WORDS_REGEX = re.compile(r'(?:<br>)([a-zA-Z]+)(?=<br>)')

alphabet_detector = AlphabetDetector()


class AnagramCommand(Command):
    name = 'anagram'
    description = 'анаграмма'

    def init(self, arguments):
        if not self.is_allowed():
            return

        if arguments is not None:
            self.delete_last_command()
            self._reply_with_result(arguments)
        else:
            self.set_last_command()
            self.reply_with_message(
                'Из каких буквы хочешь получить анаграмму?'
            )

    def handle(self, state):
        self.delete_last_command()
        if self.message.text is not None:
            self._reply_with_result(self.message.text)
        else:
            self.reply_with_message('Ты должен ввести буквы!')

    def _reply_with_result(self, letters):
        if not letters.isalpha():
            self.reply_with_message('Символы, кроме букв, запрещены!')
            return

        words = self._get_anagram_words(letters)
        if words is not None:
            if words:
                self.reply_with_message(', '.join(words))
            else:
                self.reply_with_message(
                    'Из введенных буквы анаграмма не найдена \ud83d\ude14'
                )
        else:
            self.reply_with_message('Не могу получить анаграммы.')

    def _get_anagram_words(self, letters):
        letters = letters.lower()

        if alphabet_detector.is_latin(letters):
            response = requests.get('http://wordsmith.org/anagram/anagram.cgi', params={
                'anagram': letters,
                'd': 1,
                'k': 0
            })
            if response.status_code != 200:
                return

            return WORDS_REGEX.findall(response.text.replace('\n', ''))

        if alphabet_detector.is_cyrillic(letters):
            letters = letters.replace('ё', 'е')

            response = requests.get('http://anagram.poncy.ru/anagram-decoding.cgi', params={
                'name': 'anagram_main',
                'inword': letters,
                'answer_type': 1
            })
            if response.status_code != 200:
                return

            return [word.lower() for word in response.json()['result']]
