from .interaction import Interaction


class Statement(Interaction):
    def __init__(self, bot, message):
        Interaction.__init__(self, bot, message)

    @staticmethod
    def satisfies(message):
        raise NotImplementedError

    def handle(self):
        raise NotImplementedError
