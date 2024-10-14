from .statement import Statement

from core.game_engines import get_game_engine
from core.models import Player


class EnterCodeStatement(Statement):
    @staticmethod
    def satisfies(message):
        return bool(message.text)

    def handle(self):
        if not self.is_allowed():
            return

        game_engine = get_game_engine()
        if game_engine is None or game_engine.is_restricted:
            return

        chat = self.get_chat()

        if chat.is_brute_force:
            code = self.message.text
        else:
            code = game_engine.prepare_code(self.message.text)
            if code is None:
                return

        # self.reply_with_message(code)

        enter_code_result = game_engine.enter_code(code)
        if enter_code_result.is_accepted:
            try:
                Player.increment(
                    self.message.from_user.id,
                    self.message.from_user.name
                )
            except:
                self.send_message_to_admin(traceback.format_exc())

        if chat.is_brute_force and not enter_code_result.is_accepted:
            return

        self.reply_with_message(
            enter_code_result.message,
            reply_to_message_id=self.message.message_id
        )
