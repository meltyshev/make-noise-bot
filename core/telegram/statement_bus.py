class StatementBus:
    def __init__(self, bot, statement_classes):
        self._bot = bot
        self._statement_classes = statement_classes

    def handle_statement(self, message):
        for statement_class in self._statement_classes:
            if statement_class.satisfies(message):
                statement_class(self._bot, message).handle()
                return True

        return False
