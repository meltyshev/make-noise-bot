from .command import Command

# FIXME: replace photo ids
HINTS = [{
    'name': 'Таблица Менделеева',
    'keys': ['таблица менделеева', 'менделеев', 'химия'],
    'replies': [{'photo': 'AgADAgADQagxG8d7yUhw1r185xd49VIzSw0ABK7xC-yulBm-mUYSAAEC'}]
}, {
    'name': 'Шрифт Брайля',
    'keys': ['шрифт брайля', 'брайль'],
    'replies': [
        {'photo': 'AgADAgADUqgxG2nnuUirgMejXAzZzOEUSw0ABKoj9UJ3S8V79DwSAAEC'},
        {'message': 'http://braille.su/decoder.php'}
    ]
}, {
    'name': 'Шифр пляшущие человечки',
    'keys': ['шифр пляшущие человечки', 'пляшущие человечки'],
    'replies': [{'photo': 'AgADAgADU6gxG2nnuUhEZgABsWC8hE54BzMOAATcgrTGedX4ztnaAAIC'}]
}, {
    'name': 'Цифры Майя',
    'keys': ['цифры майя'],
    'replies': [
        {'photo': 'AgADAgADVKgxG2nnuUhP1YWbDkwAAWINFEsNAASqSj-sfi9kgFI6EgABAg'},
        {'photo': 'AgADAgADVagxG2nnuUhINq5sGUiIthsGMw4ABImy7-2NeiELAtkAAgI'}
    ]
}, {
    'name': 'Шифр Цезаря',
    'keys': ['шифр цезаря', 'цезарь'],
    'replies': [{'message': 'https://planetcalc.ru/1434/'}]
}, {
    'name': 'Шифр ДНК',
    'keys': ['шифр днк', 'днк'],
    'replies': [{'photo': 'AgADAgADT6gxG8d7wUh1Q1jnN1tokuQFMw4ABE2hYQXtVb5axNgAAgI'}]
}, {
    'name': 'Шифр Бэкона',
    'keys': ['шифр бэкона', 'бэкон'],
    'replies': [{'photo': 'AgADAgADtKgxG8bdwEh-Y92dY5AOMp80Sw0ABKsoXYMooSf-akwSAAEC'}]
}, {
    'name': 'Шифр Виженера',
    'keys': ['шифр виженера', 'виженер'],
    'replies': [{'message': 'https://planetcalc.ru/2463/'}]
}, {
    'name': 'Азбука жестов военных',
    'keys': ['азбука жестов военных', 'жесты военных', '[военные]', '[военный]'],
    'replies': [
        {'photo': 'AgADAgADSagxG8d7wUgIBUZ6OmXnBQ4FMw4ABC_ZMhcjGALp-98AAgI'},
        {'photo': 'AgADAgADSqgxG8d7wUhRxy8tjp9tUSXpAw4ABGMC52bAdPHtcxYAAgI'},
        {'photo': 'AgADAgADS6gxG8d7wUgIWMEbTktpE5LWDw4ABN2PPmq-zS0RZj8DAAEC'},
        {'photo': 'AgADAgADTKgxG8d7wUg-yn8mgSQes_4AATMOAARpU3_f028RTXLcAAIC'}
    ]
}, {
    'name': 'Декодер',
    'keys': ['декодер', 'кодировка'],
    'replies': [{'message': 'https://www.artlebedev.ru/decoder/'}]
}, {
    'name': 'Врачебный алфавит',
    'keys': ['врачебный алфавит', 'врач'],
    'replies': [{'photo': 'AgADAgADQqgxG8d7yUiZVMQMiM1jPJLpAw4ABLNnfHmBYdQ0ghYAAgI'}]
}, {
    'name': 'Раскладка клавиатуры',
    'keys': ['раскладка клавиатуры', 'клавиатура', 'клава'],
    'replies': [{'photo': 'AgADAgADRKgxG8d7yUhPafgExIEzBA72Aw4ABE6QWfhehgABQ_gWAAIC'}]
}]


class HintCommand(Command):
    name = 'hint'
    description = 'поиск шифра'

    def init(self, arguments):
        if not self.is_allowed():
            return

        if arguments is not None:
            self.delete_last_command()
            self._reply_with_result(arguments)
        else:
            self.set_last_command()
            self.reply_with_message('По какой строке хочешь найти шифр?')

    def handle(self, state):
        self.delete_last_command()
        if self.message.text is not None:
            self._reply_with_result(self.message.text)
        else:
            self.reply_with_message('Ты должен ввести строку!')

    def _reply_with_result(self, search):
        hints = []

        search = search.lower()
        for hint in HINTS:
            if search in hint['keys']:
                hints.append(hint)

        if not hints:
            seach_parts = search.split()
            for hint in HINTS:
                for seach_part in seach_parts:
                    if self._is_hint_part(hint, seach_part):
                        hints.append(hint)
                        break

        if hints:
            if len(hints) > 5:
                self.reply_with_message('Возможные варианты, уточните:\n%s' % '\n'.join(
                    [hint['name'] for hint in hints]
                ))
            else:
                for hint in hints:
                    for reply in hint['replies']:
                        if 'message' in reply:
                            self.reply_with_message('%s\n%s' % (
                                hint['name'],
                                reply['message']
                            ))
                        elif 'photo' in reply:
                            self.reply_with_photo(reply['photo'], hint['name'])
        else:
            self.reply_with_message(
                'По введенной строке шифры не найдены \ud83d\ude14'
            )

    def _is_hint_part(self, hint, seach_part):
        for key in hint['keys']:
            if seach_part in key:
                return True
        return False
