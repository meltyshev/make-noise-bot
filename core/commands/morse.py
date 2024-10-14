from .command import Command

MORSE_EN = {
    '.-': 'a',
    '-...': 'b',
    '-.-.': 'c',
    '-..': 'd',
    '.': 'e',
    '..-.': 'f',
    '--.': 'g',
    '....': 'h',
    '..': 'i',
    '.---': 'j',
    '-.-': 'k',
    '.-..': 'l',
    '--': 'm',
    '-.': 'n',
    '---': 'o',
    '.--.': 'p',
    '--.-': 'q',
    '.-.': 'r',
    '...': 's',
    '-': 't',
    '..-': 'u',
    '...-': 'v',
    '.--': 'w',
    '-..-': 'x',
    '-.--': 'y',
    '--..': 'z'
}

MORSE_RU = {
    '.-': 'а',
    '-...': 'б',
    '.--': 'в',
    '--.': 'г',
    '-..': 'д',
    '.': 'е',
    '...-': 'ж',
    '--..': 'з',
    '..': 'и',
    '.---': 'й',
    '-.-': 'к',
    '.-..': 'л',
    '--': 'м',
    '-.': 'н',
    '---': 'о',
    '.--.': 'п',
    '.-.': 'р',
    '...': 'с',
    '-': 'т',
    '..-': 'у',
    '..-.': 'ф',
    '....': 'х',
    '-.-.': 'ц',
    '---.': 'ч',
    '----': 'ш',
    '--.-': 'щ',
    '--.--': 'ъ',
    '-.--': 'ы',
    '-..-': 'ь',
    '..-..': 'э',
    '..--': 'ю',
    '.-.-': 'я'
}

MORSE_COMMON = {
    '-----': '0',
    '.----': '1',
    '..---': '2',
    '...--': '3',
    '....-': '4',
    '.....': '5',
    '-....': '6',
    '--...': '7',
    '---..': '8',
    '----.': '9',
    '......': '.',
    '.-.-.-': ',',
    '---...': ':',
    '-.-.-': ';',
    '-.--.-': '|',
    '.----.': '\'',
    '.-..-.': '"',
    '-....-': '-',
    '-..-.': '/',
    '..--..': '?',
    '--..--': '!',
    '.--.-.': '@'
}


class MorseCommand(Command):
    name = 'morse'
    description = 'морзе'

    def init(self, arguments):
        if not self.is_allowed():
            return

        if arguments is not None:
            self.delete_last_command()
            self._reply_with_result(arguments)
        else:
            self.set_last_command()
            self.reply_with_message('Какую морзянку хочешь перевести?')

    def handle(self, state):
        self.delete_last_command()
        if self.message.text is not None:
            self._reply_with_result(self.message.text)
        else:
            self.reply_with_message('Ты должен ввести морзянку!')

    def _reply_with_result(self, letters):
        letters = letters.replace('_', '-')

        word_en = ''
        word_ru = ''

        for letter in letters.split():
            if letter in MORSE_COMMON:
                word_en += MORSE_COMMON[letter]
                word_ru += MORSE_COMMON[letter]
            else:
                word_en += MORSE_EN[letter] if letter in MORSE_EN else '(%s)' % letter
                word_ru += MORSE_RU[letter] if letter in MORSE_RU else '(%s)' % letter

        self.reply_with_message(word_en)
        self.reply_with_message(word_ru)
