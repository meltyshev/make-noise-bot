from .location import Location


class Venue:
    @staticmethod
    def optional(data):
        return Venue(data) if data is not None else None

    def __init__(self, data):
        self.location = Location(data['location'])
        self.title = data['title']
        self.address = data['address']
        self.foursquare_id = data.get('foursquare_id')
