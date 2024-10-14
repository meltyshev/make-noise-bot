from core.redis import r
from mongoengine import Document, IntField, ListField, StringField

DOCUMENT_KEY = 'game-config'


class GameConfig(Document):
    key = StringField(required=True, unique=True, default=DOCUMENT_KEY)

    engine = StringField(required=True, default='DozorClassic', choices=(
        'DozorClassic',
        'DozorLite',
        'DozorClassicPrequel',
        'DozorLitePrequel'
    ))  # TODO
    city = StringField(required=True, default='e-burg')
    login = StringField(required=True, default='-')
    password = StringField(required=True, default='-')
    pincode = StringField(required=True, default='-')
    game_id = StringField(required=True, default='-')
    league = StringField(required=True, default='-') # TODO: int?
    code_formats = ListField(ListField(StringField(), required=True), default=[['dr', 'др', '--']])
    subscribers = ListField(IntField(required=True), default=[])

    @staticmethod
    def create():
        config = GameConfig()
        config.save()
        return config

    @staticmethod
    def get():
        config = r.pget(DOCUMENT_KEY)
        if config is not None:
            return config

        config = GameConfig.objects(key=DOCUMENT_KEY).first()
        if config is None:
            return

        config._cache()
        return config

    @staticmethod
    def get_or_create():
        config = GameConfig.get()
        if config is None:
            try:
                config = GameConfig.create()
            except NotUniqueError:
                return GameConfig.get_or_create()

        return config

    @staticmethod
    def reset():
        config = GameConfig.get()
        if config is not None:
            config.delete()
            config = GameConfig.create()

        return config

    def save(self):
        super().save()
        self._cache()

    def delete(self):
        super().delete()
        r.delete(DOCUMENT_KEY)

    def _cache(self):
        r.pset(DOCUMENT_KEY, self)
