from core.redis import r
from mongoengine import Document, IntField, StringField

DOCUMENT_KEY = 'spot:%s'


class Spot(Document):
    key = StringField(required=True, unique=True)

    location = StringField(required=True)

    @staticmethod
    def create(**kwargs):
        spot = Spot(**kwargs)
        spot.save()
        return spot

    @staticmethod
    def get(name):
        spot = r.pget(DOCUMENT_KEY % name)
        if spot is not None:
            return spot

        spot = Spot.objects(key=name).first()
        if spot is None:
            return

        spot._cache()
        return spot

    def save(self):
        super().save()
        self._cache()

    def delete(self):
        super().delete()
        r.delete(DOCUMENT_KEY % self.key)

    def _cache(self):
        r.pset(DOCUMENT_KEY % self.key, self)
