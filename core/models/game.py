from core.redis import r
from mongoengine import BooleanField, Document, IntField, ListField, StringField

DOCUMENT_KEY = 'game'


class Game(Document):
    key = StringField(required=True, unique=True, default=DOCUMENT_KEY)

    engine = StringField(required=True, choices=(
        'DozorClassic',
        'DozorLite',
        'DozorClassicPrequel',
        'DozorLitePrequel'
    ))  # TODO
    city = StringField(required=True)
    code_formats = ListField(ListField(StringField(), required=True))
    subscribers = ListField(IntField(required=True), default=[])
    is_restricted = BooleanField(required=True, default=False)

    login = StringField()
    password = StringField()
    pincode = StringField()
    game_id = StringField()
    league = StringField()
    session = StringField()
    level_number = IntField()
    hint_number = IntField()
    solved_spoilers = ListField(IntField(required=True), default=[])
    pinned_level_number = IntField()

    @staticmethod
    def create(**kwargs):
        game = Game(**kwargs)
        game.save()
        return game

    @staticmethod
    def get():
        game = r.pget(DOCUMENT_KEY)
        if game is not None:
            return game

        game = Game.objects(key=DOCUMENT_KEY).first()
        if game is None:
            return

        game._cache()
        return game

    def save(self):
        super().save()
        self._cache()

    def modify(self, **kwargs):
        super().modify(**kwargs)
        self._cache()

    def delete(self):
        super().delete()
        r.delete(DOCUMENT_KEY)

    def _cache(self):
        r.pset(DOCUMENT_KEY, self)
