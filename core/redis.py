import os
import pickle

from redis import StrictRedis


class PickledRedis(StrictRedis):
    def pset(self, key, value, *args, **kwargs):
        return self.set(key, pickle.dumps(value), *args, **kwargs)    

    def pget(self, *args, **kwargs):
        data = self.get(*args, **kwargs)
        if data is None:
            return data

        return pickle.loads(data)


r = PickledRedis.from_url(os.environ['REDIS_URL'], db=int(os.environ['REDIS_DB']))
