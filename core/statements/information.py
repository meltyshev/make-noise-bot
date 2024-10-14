from .statement import Statement

from core.game_engines import get_game_engine


class InformationStatement(Statement):
    @staticmethod
    def satisfies(message):
        return message.text == '?'

    def handle(self):
        if not self.is_allowed():
            return

        game_engine = get_game_engine()
        if game_engine is None:
            return

        if not game_engine.load_game_data():
            self.reply_with_message('Не могу загрузить движок.')
            return

        messages = []

        sectors = game_engine.get_sectors()
        if sectors is not None:
            sector_lines = []
            for sector in sectors:
                format_1_column = '{:>3} {:<2} {:2}'
                format_2_column = '%s   %s' % (
                    format_1_column,
                    format_1_column
                )

                codes = sector['codes']
                codes_amount = len(sector['codes'])

                codes_balance = codes_amount % 2
                codes_per_column = codes_amount // 2

                codes_1_column = codes[:codes_per_column]
                codes_2_column = codes[codes_per_column + codes_balance:]

                code_lines = [sector['name'] + ':']
                for code_1_column, code_2_column in zip(codes_1_column, codes_2_column):
                    code_lines.append(format_2_column.format(*(
                        '%s)' % code_1_column['number'],
                        code_1_column['hazard'],
                        'ok' if code_1_column['is_entered'] else '',
                        '%s)' % code_2_column['number'],
                        code_2_column['hazard'],
                        'ok' if code_2_column['is_entered'] else ''
                    )))

                if codes_balance == 1:
                    code = codes[codes_per_column]
                    code_lines.append(format_1_column.format(*(
                        '%s)' % code['number'],
                        code['hazard'],
                        'ok' if code['is_entered'] else ''
                    )))

                sector_lines.append('\n'.join(code_lines))

            messages.append('\n\n'.join(sector_lines))

        progress = game_engine.get_progress()
        if progress is not None:
            messages.append(progress)

        if messages:
            self.reply_with_message(
                '<pre>%s</pre>' % '\n\n'.join(messages),
                parse_mode='HTML'
            )
        else:
            self.reply_with_message('Информации нет.')
