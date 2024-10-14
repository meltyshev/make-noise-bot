from .command import Command


class BruteForceCommand(Command):
    name = 'bruteforce'

    def init(self, arguments):
        if not self.is_allowed() or (self.message.chat.type != 'private' and not self.is_manager()):
            return

        chat = self.get_chat()
        chat.is_brute_force = not chat.is_brute_force
        chat.save()

        if chat.is_brute_force:
            self.reply_with_message('Режим перебора активирован.')
        else:
            self.reply_with_message('Режим перебора отключен.')
