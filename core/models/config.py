from core.redis import r
from mongoengine import BooleanField, Document, IntField, ListField, StringField

DOCUMENT_KEY = 'config'


class Config(Document):
    key = StringField(required=True, unique=True, default=DOCUMENT_KEY)

    managers = ListField(IntField(required=True), default=[])
    is_leave_mode = BooleanField(required=True, default=False)

    @staticmethod
    def create():
        config = Config()
        config.save()
        return config

    @staticmethod
    def get():
        config = r.pget(DOCUMENT_KEY)
        if config is not None:
            return config

        config = Config.objects(key=DOCUMENT_KEY).first()
        if config is None:
            return

        config._cache()
        return config

    @staticmethod
    def get_or_create():
        config = Config.get()
        if config is None:
            try:
                config = Config.create()
            except NotUniqueError:
                return Config.get_or_create()

        return config

    @staticmethod
    def reset():
        config = Config.get()
        if config is not None:
            config.delete()
            config = Config.create()

        return config

    def save(self):
        super().save()
        self._cache()

    def delete(self):
        super().delete()
        r.delete(DOCUMENT_KEY)

    def _cache(self):
        r.pset(DOCUMENT_KEY, self)
