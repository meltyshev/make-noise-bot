from .command import Command

from core.models import Chat


class PermissionCommand(Command):
    name = 'permission'

    def init(self, arguments):
        message_chat = self.message.chat

        chat = Chat.get(message_chat.id)
        if chat is not None:
            if chat.permission == 'requested':
                self.reply_with_message('Текущий статус - запрошено.')
            elif chat.permission == 'allowed':
                self.reply_with_message('Текущий статус - разрешено.')
            elif chat.permission == 'forbidden':
                self.reply_with_message('Текущий статус - запрещено.')
            return

        Chat.create(
            key=message_chat.id,
            type=message_chat.type,
            title=message_chat.title,
            username=message_chat.username,
            first_name=message_chat.first_name,
            last_name=message_chat.last_name
        )

        text = 'Запрос для @%s:\n' % self.me.username
        text += 'type: %s' % message_chat.type
        if message_chat.title is not None:
            text += '\ntitle: %s' % message_chat.title
        if message_chat.username is not None:
            text += '\nusername: %s' % message_chat.username
        if message_chat.first_name is not None:
            text += '\nfirst_name: %s' % message_chat.first_name
        if message_chat.last_name is not None:
            text += '\nlast_name: %s' % message_chat.last_name

        message = self.send_message_to_admin(text)
        self.send_message_to_admin(
            message_chat.id,
            reply_to_message_id=message.message_id
        )

        self.reply_with_message('Запрос сделан, ожидайте решения.')
