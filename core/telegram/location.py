class Location:
    @staticmethod
    def optional(data):
        return Location(data) if data is not None else None

    def __init__(self, data):
        self.longitude = data['longitude']
        self.latitude = data['latitude']
