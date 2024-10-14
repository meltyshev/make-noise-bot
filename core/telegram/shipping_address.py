class ShippingAddress:
    @staticmethod
    def optional(data):
        return ShippingAddress(data) if data is not None else None

    def __init__(self, data):
        self.country_code = data['country_code']
        self.state = data['state']
        self.city = data['city']
        self.street_line1 = data['street_line1']
        self.street_line2 = data['street_line2']
        self.post_code = data['post_code']
