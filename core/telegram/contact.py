class Contact:
    @staticmethod
    def optional(data):
        return Contact(data) if data is not None else None

    def __init__(self, data):
        self.phone_number = data['phone_number']
        self.first_name = data['first_name']
        self.last_name = data.get('last_name')
        self.user_id = data.get('user_id')
