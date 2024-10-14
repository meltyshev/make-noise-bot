class Invoice:
    @staticmethod
    def optional(data):
        return Invoice(data) if data is not None else None

    def __init__(self, data):
        self.title = data['title']
        self.description = data['description']
        self.start_parameter = data['start_parameter']
        self.currency = data['currency']
        self.total_amount = data['total_amount']
