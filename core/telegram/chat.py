from .chat_photo import ChatPhoto
# from .message import Message


class Chat:
    @staticmethod
    def optional(data):
        return Chat(data) if data is not None else None

    def __init__(self, data):
        self.id = data['id']
        self.type = data['type']
        self.title = data.get('title')
        self.username = data.get('username')
        self.first_name = data.get('first_name')
        self.last_name = data.get('last_name')
        self.all_members_are_administrators = data.get(
            'all_members_are_administrators'
        )
        self.photo = ChatPhoto.optional(data.get('photo'))
        self.description = data.get('description')
        self.invite_link = data.get('invite_link')
        # self.pinned_message = Message.optional(data.get('pinned_message'))
        self.sticker_set_name = data.get('sticker_set_name')
        self.can_set_sticker_set = data.get('can_set_sticker_set')
