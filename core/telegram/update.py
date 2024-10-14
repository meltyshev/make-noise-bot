from .message import Message


class Update:
    def __init__(self, data):
        self.update_id = data['update_id']
        self.message = Message.optional(data.get('message'))
