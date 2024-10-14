from .user import User


class MessageEntity:
    def __init__(self, data):
        self.type = data['type']
        self.offset = data['offset']
        self.length = data['length']
        self.url = data.get('url')
        self.user = User.optional(data.get('user'))
