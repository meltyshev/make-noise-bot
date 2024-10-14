class ChatPhoto:
    @staticmethod
    def optional(data):
        return ChatPhoto(data) if data is not None else None

    def __init__(self, data):
        self.small_file_id = data['small_file_id']
        self.big_file_id = data['big_file_id']
