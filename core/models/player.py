from core.redis import r

HASH_KEY = 'player:%s'


class Player():
    @staticmethod
    def increment(id, name):
        hash_key = HASH_KEY % id

        pipe = r.pipeline()
        pipe.hset(hash_key, 'name', name)
        pipe.hincrby(hash_key, 'total')
        pipe.execute()

    @staticmethod
    def rating():
        players = []
        for hash_key in r.keys(HASH_KEY % '*'):
            player = {}
            for key, value in r.hgetall(hash_key).items():
                player[key.decode()] = value.decode()

            if player:
                players.append(player)

        return sorted(players, key=lambda player: int(player['total']), reverse=True) # TODO: int with fallback

    @staticmethod
    def clear():
        hash_keys = r.keys(HASH_KEY % '*')
        if hash_keys:
            r.delete(*hash_keys)
