class LastCommand:
    def __init__(self, name, state):
        self.name = name
        self.state = state

    def to_dict(self):
        data = {
            'name': self.name
        }
        if self.state is not None:
            data['state'] = self.state

        return data

    @staticmethod
    def from_dict(data):
        return LastCommand(data['name'], data.get('state'))
