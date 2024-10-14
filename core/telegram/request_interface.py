class RequestInterface:
    def __init__(self):
        pass

    def get(self, url):
        raise NotImplementedError

    def post(self, url, data):
        raise NotImplementedError
