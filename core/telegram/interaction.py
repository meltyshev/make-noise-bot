class Interaction:
    def __init__(self, bot, message):
        self.message = message
        self._bot = bot

    @property
    def me(self):
        return self._bot.me

    def send_message(self, *args, **kwargs):
        return self._bot.send_message(*args, **kwargs)

    def send_photo(self, *args, **kwargs):
        return self._bot.send_photo(*args, **kwargs)

    def send_audio(self, *args, **kwargs):
        return self._bot.send_audio(*args, **kwargs)

    def send_document(self, *args, **kwargs):
        return self._bot.send_document(*args, **kwargs)

    def send_video(self, *args, **kwargs):
        return self._bot.send_video(*args, **kwargs)

    def send_voice(self, *args, **kwargs):
        return self._bot.send_voice(*args, **kwargs)

    def send_video_note(self, *args, **kwargs):
        return self._bot.send_video_note(*args, **kwargs)

    def send_location(self, *args, **kwargs):
        return self._bot.send_location(*args, **kwargs)

    def send_sticker(self, *args, **kwargs):
        return self._bot.send_sticker(*args, **kwargs)

    def reply_with_message(self, *args, **kwargs):
        if kwargs.get('reply_to_message_id') is None and self.message.chat.type in ('group', 'supergroup'):
            kwargs['reply_to_message_id'] = self.message.message_id

        return self.send_message(self.message.chat.id, *args, **kwargs)

    def reply_with_photo(self, *args, **kwargs):
        if kwargs.get('reply_to_message_id') is None and self.message.chat.type in ('group', 'supergroup'):
            kwargs['reply_to_message_id'] = self.message.message_id

        return self.send_photo(self.message.chat.id, *args, **kwargs)

    def reply_with_audio(self, *args, **kwargs):
        if kwargs.get('reply_to_message_id') is None and self.message.chat.type in ('group', 'supergroup'):
            kwargs['reply_to_message_id'] = self.message.message_id

        return self.send_audio(self.message.chat.id, *args, **kwargs)

    def reply_with_document(self, *args, **kwargs):
        if kwargs.get('reply_to_message_id') is None and self.message.chat.type in ('group', 'supergroup'):
            kwargs['reply_to_message_id'] = self.message.message_id

        return self.send_document(self.message.chat.id, *args, **kwargs)

    def reply_with_video(self, *args, **kwargs):
        if kwargs.get('reply_to_message_id') is None and self.message.chat.type in ('group', 'supergroup'):
            kwargs['reply_to_message_id'] = self.message.message_id

        return self.send_video(self.message.chat.id, *args, **kwargs)

    def reply_with_voice(self, *args, **kwargs):
        if kwargs.get('reply_to_message_id') is None and self.message.chat.type in ('group', 'supergroup'):
            kwargs['reply_to_message_id'] = self.message.message_id

        return self.send_voice(self.message.chat.id, *args, **kwargs)

    def reply_with_video_note(self, *args, **kwargs):
        if kwargs.get('reply_to_message_id') is None and self.message.chat.type in ('group', 'supergroup'):
            kwargs['reply_to_message_id'] = self.message.message_id

        return self.send_video_note(self.message.chat.id, *args, **kwargs)

    def reply_with_location(self, *args, **kwargs):
        if kwargs.get('reply_to_message_id') is None and self.message.chat.type in ('group', 'supergroup'):
            kwargs['reply_to_message_id'] = self.message.message_id

        return self.send_location(self.message.chat.id, *args, **kwargs)

    def reply_with_sticker(self, *args, **kwargs):
        if kwargs.get('reply_to_message_id') is None and self.message.chat.type in ('group', 'supergroup'):
            kwargs['reply_to_message_id'] = self.message.message_id

        return self.send_sticker(self.message.chat.id, *args, **kwargs)

    def send_message_to_admin(self, *args, **kwargs):
        return self._bot.send_message_to_admin(*args, **kwargs)

    def get_file(self, *args, **kwargs):
        return self._bot.get_file(*args, **kwargs)

    def leave_chat(self, *args, **kwargs):
        return self._bot.leave_chat(*args, **kwargs)
