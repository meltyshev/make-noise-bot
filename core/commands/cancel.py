from .command import Command


class CancelCommand(Command):
    name = 'cancel'
    description = 'отменить текущую команду'

    def init(self, arguments):
        last_command = self.get_last_command()
        if last_command is not None:
            self.delete_last_command()
            self.reply_with_message(
                'Команда /%s отменена.' % last_command.name
            )
        else:
            self.reply_with_message('Нет текущих команд.')
