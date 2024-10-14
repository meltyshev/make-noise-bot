class KeyboardButton:
    def __init__(self, text, request_contact=None, request_location=None):
        self.text = text
        self.request_contact = request_contact
        self.request_location = request_location

    def to_dict(self):
        data = {
            'text': self.text
        }
        if self.request_contact is not None:
            data['request_contact'] = self.request_contact
        if self.request_location is not None:
            data['request_location'] = self.request_location

        return data
