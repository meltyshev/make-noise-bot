import re

from .client import Client
from .command_bus import CommandBus
from .event_bus import EventBus
from .statement_bus import StatementBus
from .update import Update


class Bot:
    def __init__(self, api_token, admin_id, request_interface, last_command_storage_interface, command_classes=(), event_classes=(), statement_classes=()):
        self.api_token = api_token
        self.admin_id = admin_id

        self._client = Client(api_token, request_interface)
        self.me = self._client.get_me()

        self._command_bus = CommandBus(self, command_classes)
        self._event_bus = EventBus(self, event_classes)
        self._statement_bus = StatementBus(self, statement_classes)
        self._last_command_storage_interface = last_command_storage_interface

    @property
    def help(self):
        return self._command_bus.help

    def handle(self, data):
        update = Update(data)
        if update.message is None:
            return

        message = update.message

        # Commands
        if message.text and message.text[0] == '/' and len(message.text) > 1:
            arguments = message.text[1:].split(maxsplit=1)

            command_name = arguments[0]
            if command_name.endswith('@%s' % self.me.username):
                command_name = command_name[:-(len(self.me.username) + 1)]

            command = self._command_bus.get_command(command_name, message)
            if command is not None:
                command.init(arguments[1] if len(arguments) == 2 else None)

            return

        # Last commands
        last_command = self.get_last_command(
            message.from_user.id,
            message.chat.id
        )
        if last_command is not None:
            command = self._command_bus.get_command(
                last_command.name,
                message
            )
            if command is not None:
                command.handle(last_command.state)
                return

        # Events
        if self._event_bus.handle_event(message):
            return

        # Statements
        self._statement_bus.handle_statement(message)

    def set_webhook(self, *args, **kwargs):
        return self._client.set_webhook(*args, **kwargs)

    def send_message(self, *args, **kwargs):
        return self._client.send_message(*args, **kwargs)

    def send_photo(self, *args, **kwargs):
        return self._client.send_photo(*args, **kwargs)

    def send_audio(self, *args, **kwargs):
        return self._client.send_audio(*args, **kwargs)

    def send_document(self, *args, **kwargs):
        return self._client.send_document(*args, **kwargs)

    def send_video(self, *args, **kwargs):
        return self._client.send_video(*args, **kwargs)

    def send_voice(self, *args, **kwargs):
        return self._client.send_voice(*args, **kwargs)

    def send_video_note(self, *args, **kwargs):
        return self._client.send_video_note(*args, **kwargs)

    def send_location(self, *args, **kwargs):
        return self._client.send_location(*args, **kwargs)

    def send_sticker(self, *args, **kwargs):
        return self._client.send_sticker(*args, **kwargs)

    def send_message_to_admin(self, *args, **kwargs):
        return self.send_message(self.admin_id, *args, **kwargs)

    def get_file(self, *args, **kwargs):
        return self._client.get_file(*args, **kwargs)

    def leave_chat(self, *args, **kwargs):
        return self._client.leave_chat(*args, **kwargs)

    def get_last_command(self, *args, **kwargs):
        return self._last_command_storage_interface.get(*args, **kwargs)

    def set_last_command(self, *args, **kwargs):
        self._last_command_storage_interface.set(*args, **kwargs)

    def delete_last_command(self, *args, **kwargs):
        self._last_command_storage_interface.delete(*args, **kwargs)
