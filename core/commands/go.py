import os
import re
import requests

from .command import Command

from core.models import Spot

DD_COORDINATES_REGEX = re.compile(r'(\d{1,3}\.\d+)[^\d]+(\d{1,3}\.\d+)')
DMS_COORDINATES_REGEX = re.compile(
    r'(\d{1,3})[^\d]+(\d{1,2})[^\d]+(\d{1,2}((\.\d{1,2})?))[^\d]+(\d{1,3})[^\d]+(\d{1,2})[^\d]+(\d{1,2}((\.\d{1,2})?))'
)

GOOGLE_API_KEY = os.environ['GOOGLE_API_KEY']


class GoCommand(Command):
    name = 'go'

    def init(self, arguments):
        if not self.is_allowed():
            return

        if arguments is not None:
            self.delete_last_command()
            self._reply_with_result(arguments)
        else:
            self.set_last_command()
            self.reply_with_message('Какой адрес/координаты?')

    def handle(self, state):
        self.delete_last_command()
        self._reply_with_result(self.message.text)

    def _reply_with_result(self, message):
        coordinates = self._check_and_convert_coordinates(message)
        destination = message if coordinates is None else ','.join(
            map(str, coordinates)
        )

        nearest_list = self._get_nearest_list(Spot.objects.all(), destination)
        if nearest_list:
            self.reply_with_message('\n'.join(
                ['%s - %s (%s)' % (
                    nearest['name'],
                    nearest['duration_text'],
                    nearest['distance_text']
                ) for nearest in nearest_list]
            ))

        if coordinates is None:
            coordinates = self._get_address_coordinates(destination)
        if coordinates is not None:
            self.reply_with_location(coordinates[0], coordinates[1])
        else:
            self.reply_with_message('Не могу найти место \ud83d\ude14')

    def _check_and_convert_coordinates(self, string):
        match = DD_COORDINATES_REGEX.search(string)
        if match is not None:
            groups = match.groups()
            return (float(groups[0]), float(groups[1]))

        match = DMS_COORDINATES_REGEX.search(string)
        if match is None:
            return

        groups = match.groups()
        return (self._convert_coordinate_parts(groups[0], groups[1], groups[2]),
                self._convert_coordinate_parts(groups[5], groups[6], groups[7]))

    def _convert_coordinate_parts(self, part1, part2, part3):
        return round(float(part1) + float(part2) / 60 + float(part3) / 3600, 6)

    def _get_nearest_list(self, spots, destination):
        if not spots:
            return

        response = requests.get('https://maps.googleapis.com/maps/api/distancematrix/json', params={
            'origins': '|'.join([spot.location for spot in spots]),
            'destinations': destination,
            'departure_time': 'now',
            'language': 'ru',
            'key': GOOGLE_API_KEY
        })
        if response.status_code != 200:
            return

        data = response.json()
        if data['status'] != 'OK':
            return

        nearest_list = []
        for i, spot in enumerate(spots):
            element = data['rows'][i]['elements'][0]
            if element['status'] == 'OK':
                duration = element.get(
                    'duration_in_traffic',
                    element['duration']
                )
                nearest_list.append({
                    'name': spot.key,
                    'duration': duration['value'],
                    'duration_text': duration['text'],
                    'distance_text': element['distance']['text'] + '.'
                })

        return sorted(nearest_list, key=lambda n: n['duration'])

    def _get_address_coordinates(self, address):
        response = requests.get('https://maps.googleapis.com/maps/api/geocode/json', params={
            'address': address,
            'key': GOOGLE_API_KEY
        })
        if response.status_code != 200:
            return

        data = response.json()
        if data['status'] != 'OK':
            return

        location = data['results'][0]['geometry']['location']
        return (location['lat'], location['lng'])
