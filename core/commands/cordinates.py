from .command import Command

from math import floor


class CoordinatesCommand(Command):
    name = 'coordinates'
    # description = 'координаты'

    def init(self, arguments):
        if not self.is_allowed():
            return

        self.set_last_command()
        self.reply_with_message('Кидай местоположение.')

    def handle(self, state):
        if self.message.location is not None:
            self.delete_last_command()

            latitude = '%s°%02d\'%05.2f"N' % self._convert(
                self.message.location.latitude
            )
            longitude = '%s°%02d\'%05.2f"E' % self._convert(
                self.message.location.longitude
            )

            self.reply_with_message('%s %s' % (latitude, longitude))
        else:
            self.reply_with_message(
                'Ты должен скинуть местоположение!\n/cancel - отменить текущую команду.'
            )

    def _convert(self, decimal_degrees):
        degrees = floor(decimal_degrees)
        minutes = floor(60 * (decimal_degrees - degrees))
        seconds = round(3600 * ((decimal_degrees - degrees) - minutes / 60), 2)

        return int(degrees), int(minutes), seconds
