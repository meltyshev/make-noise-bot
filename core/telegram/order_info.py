from .shipping_address import ShippingAddress


class OrderInfo:
    @staticmethod
    def optional(data):
        return OrderInfo(data) if data is not None else None

    def __init__(self, data):
        self.name = data.get('name')
        self.phone_number = data.get('phone_number')
        self.email = data.get('email')
        self.shipping_address = ShippingAddress.optional(
            data.get('shipping_address')
        )
