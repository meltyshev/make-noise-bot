class Response:
    def __init__(self, data):
        self.ok = data['ok']
        self.description = data.get('description')
        self.result = data.get('result')
        self.error_code = data.get('error_code')
