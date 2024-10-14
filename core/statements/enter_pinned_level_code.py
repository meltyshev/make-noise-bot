from .statement import Statement

from core.game_engines import DozorClassic


class EnterPinnedLevelCodeStatement(Statement):
    @staticmethod
    def satisfies(message):
        return message.text is not None and message.text.startswith('$')

    def handle(self):
        if not self.is_allowed():
            return

        game_engine = DozorClassic.get()
        if game_engine is None:
            return

        pinned_level_number = game_engine.pinned_level_number
        if pinned_level_number is None:
            return

        code = self.message.text[1:].lower()
        enter_code_result = game_engine.enter_code(code, pinned_level_number)

        self.reply_with_message(
            enter_code_result.message,
            reply_to_message_id=self.message.message_id
        )
