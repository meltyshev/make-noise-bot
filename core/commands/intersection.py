from .command import Command


class IntersectionCommand(Command):
    name = 'intersection'
    description = 'пересечение'

    def init(self, arguments):
        if not self.is_allowed():
            return

        if arguments is not None:
            self.delete_last_command()
            self._reply_with_result(arguments)
        else:
            self.set_last_command()
            self.reply_with_message('Какие слова хочешь пересечь?')

    def handle(self, state):
        self.delete_last_command()
        if self.message.text is not None:
            self._reply_with_result(self.message.text)
        else:
            self.reply_with_message('Ты должен ввести слова!')

    def _reply_with_result(self, words):
        words = [set(word.lower()) for word in words.split()]
        if len(words) == 1:
            self.reply_with_message('Ты должен ввести хотя бы 2 слова!')
            return

        intersection = set.intersection(*words)
        if intersection:
            self.reply_with_message(''.join(intersection))
        else:
            self.reply_with_message(
                'По введенным словам нет пересечений \ud83d\ude14'
            )
