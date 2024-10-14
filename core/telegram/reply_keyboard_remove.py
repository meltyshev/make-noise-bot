class ReplyKeyboardRemove:
    def __init__(self, selective=None):
        self.selective = selective

    def to_dict(self):
        data = {
            'remove_keyboard': True
        }
        if self.selective is not None:
            data['selective'] = self.selective

        return data
