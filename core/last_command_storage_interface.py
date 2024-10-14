from .redis import r
from .telegram import LastCommand, LastCommandStorageInterface as ILastCommandStorageInterface

COMMAND_KEY = 'chat:%s:user:%s:command'


class LastCommandStorageInterface(ILastCommandStorageInterface):
    def get(self, user_id, chat_id):
        data = r.pget(self._make_key(user_id, chat_id))
        if data is not None:
            return LastCommand.from_dict(data)

    def set(self, user_id, chat_id, last_command):
        r.pset(self._make_key(user_id, chat_id), last_command.to_dict(), 3600)

    def delete(self, user_id, chat_id):
        r.delete(self._make_key(user_id, chat_id))

    def _make_key(self, user_id, chat_id):
        return COMMAND_KEY % (user_id, chat_id)
