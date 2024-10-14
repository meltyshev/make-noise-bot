class User:
    @staticmethod
    def optional(data):
        return User(data) if data is not None else None

    def __init__(self, data):
        self.id = data['id']
        self.is_bot = data['is_bot']
        self.first_name = data.get('first_name')
        self.last_name = data.get('last_name')
        self.username = data.get('username')
        self.language_code = data.get('language_code')

    @property
    def name(self):
        return ' '.join(filter(None, (self.first_name, self.last_name)))
