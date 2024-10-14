from core.redis import r
from mongoengine import BooleanField, Document, IntField, StringField

DOCUMENT_KEY = 'chat:%s'


class Chat(Document):
    key = IntField(required=True, unique=True)

    type = StringField(required=True, choices=(
        'private',
        'group',
        'supergroup',
        'channel'
    ))
    permission = StringField(required=True, choices=(
        'requested',
        'allowed',
        'forbidden'
    ), default='requested')
    is_brute_force = BooleanField(required=True, default=False)

    title = StringField()
    username = StringField()
    first_name = StringField()
    last_name = StringField()

    @staticmethod
    def create(**kwargs):
        chat = Chat(**kwargs)
        chat.save()
        return chat

    @staticmethod
    def get(id):
        chat = r.pget(DOCUMENT_KEY % id)
        if chat is not None:
            return chat

        chat = Chat.objects(key=id).first()
        if chat is None:
            return

        chat._cache()
        return chat

    @staticmethod
    def all():
        return Chat.objects

    def __str__(self):
        return ' '.join(filter(None, (self.title, self.first_name, self.last_name)))

    def save(self):
        super().save()
        self._cache()

    def delete(self):
        super().delete()
        r.delete(DOCUMENT_KEY % self.key)

    def _cache(self):
        r.pset(DOCUMENT_KEY % self.key, self)
