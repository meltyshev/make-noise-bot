import random

from .command import Command

PHRASES = (
    'Укропчика не желаете?',
    'Развооорооот..',
    'А метку «DR не светить» тоже снимать?',
    'Привет, хе-хе:)',
    'Здорова, ёптыть!',
    'ПЕРВЫЙ КОД ЗА ИГРУ!',
    'Причина остановки? Не выходи из машины!',
    'Через изоплит быстрее! Я по Яндексу пробил..',
    'По-брааатски, включи бутырку!'
)


class RomkaCommand(Command):
    name = 'romka'
    # description = 'вызвать Ромку'

    def init(self, arguments):
        self.reply_with_message(random.choice(PHRASES))
