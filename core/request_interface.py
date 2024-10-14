import requests

from .telegram import RequestInterface as IRequestInterface, Response


class RequestInterface(IRequestInterface):
    def get(self, url):
        return self._get_response(requests.get(url))

    def post(self, url, data, files={}):
        if files:
            response = requests.post(url, data=data, files=files)
        else:
            response = requests.post(url, json=data)

        return self._get_response(response)

    def _get_response(self, response):
        return Response(response.json())
