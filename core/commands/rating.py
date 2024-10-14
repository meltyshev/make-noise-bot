from .command import Command

from core.models import Player


class RatingCommand(Command):
    name = 'rating'
    description = 'рейтинг'

    def init(self, arguments):
        players = Player.rating()

        if not players:
            return self.reply_with_message('Рейтинга пока нет.')

        lines = []
        line_format = '{:>3} {:<2} {}'

        for i, player in enumerate(players, 1):
            lines.append(line_format.format(*(
                '%s)' % i,
                player['total'],
                player['name']
            )))

        self.reply_with_message(
            '<pre>%s</pre>' % '\n'.join(lines),
            parse_mode='HTML'
        )
