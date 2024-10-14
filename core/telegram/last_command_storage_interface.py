class LastCommandStorageInterface:
    def __init__(self):
        pass

    def get(self, user_id, chat_id):
        raise NotImplementedError

    def set(self, user_id, chat_id, last_command):
        raise NotImplementedError

    def delete(self, user_id, chat_id):
        raise NotImplementedError
