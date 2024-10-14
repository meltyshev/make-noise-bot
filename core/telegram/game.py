from .animation import Animation
from .message_entity import MessageEntity
from .photo_size import PhotoSize


class Game:
    @staticmethod
    def optional(data):
        return Game(data) if data is not None else None

    def __init__(self, data):
        self.title = data['title']
        self.description = data['description']

        self.photo = []
        for photo in data['photo']:
            self.photo.append(PhotoSize(photo))

        self.text = data.get('text')

        self.text_entities = []
        for text_entity in data.get('text_entities', []):
            self.text_entities.append(MessageEntity(text_entity))

        self.animation = Animation.optional(data.get('animation'))
