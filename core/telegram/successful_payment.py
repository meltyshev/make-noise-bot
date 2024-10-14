from .order_info import OrderInfo


class SuccessfulPayment:
    @staticmethod
    def optional(data):
        return SuccessfulPayment(data) if data is not None else None

    def __init__(self, data):
        self.currency = data['currency']
        self.total_amount = data['total_amount']
        self.invoice_payload = data['invoice_payload']
        self.shipping_option_id = data.get('shipping_option_id')
        self.order_info = OrderInfo.optional(data.get('order_info'))
        self.telegram_payment_charge_id = data['telegram_payment_charge_id']
        self.provider_payment_charge_id = data['provider_payment_charge_id']
