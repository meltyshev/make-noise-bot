from .audio import Audio
from .chat import Chat
from .contact import Contact
from .document import Document
from .game import Game
from .invoice import Invoice
from .location import Location
from .message_entity import MessageEntity
from .photo_size import PhotoSize
from .sticker import Sticker
from .successful_payment import SuccessfulPayment
from .user import User
from .venue import Venue
from .video import Video
from .video_note import VideoNote
from .voice import Voice


class Message:
    @staticmethod
    def optional(data):
        return Message(data) if data is not None else None

    def __init__(self, data):
        self.message_id = data['message_id']
        self.from_user = User.optional(data.get('from'))
        self.date = data['date']
        self.chat = Chat(data['chat'])
        self.forward_from = User.optional(data.get('forward_from'))
        self.forward_from_chat = Chat.optional(data.get('forward_from_chat'))
        self.forward_from_message_id = data.get('forward_from_message_id')
        self.forward_signature = data.get('forward_signature')
        self.forward_date = data.get('forward_date')
        self.reply_to_message = Message.optional(data.get('reply_to_message'))
        self.edit_date = data.get('edit_date')
        self.media_group_id = data.get('media_group_id')
        self.author_signature = data.get('author_signature')
        self.text = data.get('text')

        self.entities = []
        for entity in data.get('entities', []):
            self.entities.append(MessageEntity(entity))

        self.caption_entities = []
        for caption_entity in data.get('caption_entities', []):
            self.caption_entities.append(MessageEntity(caption_entity))

        self.audio = Audio.optional(data.get('audio'))
        self.document = Document.optional(data.get('document'))
        self.game = Game.optional(data.get('game'))

        self.photo = []
        for photo in data.get('photo', []):
            self.photo.append(PhotoSize(photo))

        self.sticker = Sticker.optional(data.get('sticker'))
        self.video = Video.optional(data.get('video'))
        self.voice = Voice.optional(data.get('voice'))
        self.video_note = VideoNote.optional(data.get('video_note'))
        self.caption = data.get('caption')
        self.contact = Contact.optional(data.get('contact'))
        self.location = Location.optional(data.get('location'))
        self.venue = Venue.optional(data.get('venue'))

        self.new_chat_members = []
        for new_chat_member in data.get('new_chat_members', []):
            self.new_chat_members.append(User(new_chat_member))

        self.left_chat_member = User.optional(data.get('left_chat_member'))
        self.new_chat_title = data.get('new_chat_title')

        self.new_chat_photo = []
        for new_chat_photo in data.get('new_chat_photo', []):
            self.new_chat_photo.append(PhotoSize(new_chat_photo))

        self.delete_chat_photo = data.get('delete_chat_photo')
        self.group_chat_created = data.get('group_chat_created')
        self.supergroup_chat_created = data.get('supergroup_chat_created')
        self.channel_chat_created = data.get('channel_chat_created')
        self.migrate_to_chat_id = data.get('migrate_to_chat_id')
        self.migrate_from_chat_id = data.get('migrate_from_chat_id')
        self.pinned_message = Message.optional(data.get('pinned_message'))
        self.invoice = Invoice.optional(data.get('invoice'))
        self.successful_payment = SuccessfulPayment.optional(
            data.get('successful_payment')
        )
