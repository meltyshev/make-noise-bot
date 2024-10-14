import re
import requests

from .game_engine import EnterCodeResult, GameEngine, BETWEEN_TAGS_REGEX, REQUEST_TIMEOUT, REQUEST_TIMEOUT_MESSAGE, UNKNOWN_ERROR_MESSAGE, USER_AGENT

from urllib.parse import parse_qs, urlparse

STATUSES = {
    52: '\u26a0\ufe0f Код к приквелу не принят. Ваша команда уже отправила этот код к приквелу.',
    53: '\u274c Код к приквелу не принят. Вы ввели неверный код.',
    54: '\u2705 Код к приквелу принят.',
    55: '\u26a0\ufe0f Код к приквелу не принят. Закончилось отведенное время на прием кода.',
    56: '\u26a0\ufe0f Код к приквелу не принят. Вы исчерпали попытки для ввода кода приквела.'
}

ACCEPTED_STATUS_CODES = (54,)

SECTORS_REGEX = re.compile(r'(?<=<strong id=orang>Код сложности:).*?(?=</strong>)')
SECTOR_ROWS_REGEX = re.compile(r'(?<=<br>).*?(?=<br>)')


def _get_status_code(location):
    try:
        return int(parse_qs(urlparse(location).query)['err'][0])
    except:
        pass


def _parse_hazard(hazard):
    if hazard == 'null':
        hazard = 'N'

    return hazard


class DozorLitePrequel(GameEngine):
    name = 'DozorLitePrequel'

    @staticmethod
    def obtain_session(config):
        url = 'http://classic.dzzzr.ru/%s/API/login.php' % config.city
        try:
            response = requests.get(
                url,
                params={
                    'login': config.login,
                    'password': config.password
                },
                auth=('', ''),
                timeout=REQUEST_TIMEOUT
            )
        except requests.exceptions.RequestException:
            return

        if response.status_code != 200:
            return

        data = response.json()
        if int(data['code']) != 2:
            return

        return data['userToken']

    @staticmethod
    def start(config):
        session = DozorLitePrequel.obtain_session(config)
        if session is None:
            return

        return GameEngine.start(
            DozorLitePrequel,
            config,
            login=config.login,
            password=config.password,
            game_id=config.game_id,
            league=config.league,
            session=session
        )

    @staticmethod
    def get():
        return GameEngine.get(DozorLitePrequel)

    def __init__(self, game):
        super().__init__(game)
        self.link = 'http://lite.dzzzr.ru/%s/?league=%s' % (self.game.city, self.game.league)
        self._headers={
            'Cookie': 'dozorSiteSession=%s' % self.game.session,
            'Referer': self.link,
            'User-Agent': USER_AGENT
        }

    def enter_code(self, code):
        try:
            response = requests.post(
                self.link,
                data={
                    'action': 'prequel_code_new',
                    'league': self.game.league,
                    'game': self.game.game_id,
                    'cod': code.encode('windows-1251')
                },
                headers=self._headers,
                timeout=REQUEST_TIMEOUT,
                allow_redirects=False
            )
        except requests.exceptions.RequestException:
            return EnterCodeResult(REQUEST_TIMEOUT_MESSAGE)

        if response.status_code == 302:
            status_code = _get_status_code(response.headers['Location'])

            return EnterCodeResult(
                STATUSES.get(status_code, UNKNOWN_ERROR_MESSAGE),
                status_code=status_code,
                is_accepted=status_code in ACCEPTED_STATUS_CODES
            )

        # elif response.status_code == 401:
            # return INVALID_PINCODE_MESSAGE

        return EnterCodeResult(UNKNOWN_ERROR_MESSAGE)

    def load_game_data(self):
        try:
            response = requests.get(
                self.link,
                headers=self._headers,
                timeout=REQUEST_TIMEOUT,
                allow_redirects=False
            )
        except requests.exceptions.RequestException:
            return False

        if response.status_code != 200:
            return False

        self._game_data = response.content.decode('windows-1251') # TODO: get part
        return True

    def get_level_number(self):
        return 0

    def get_progress(self):
        pass

    def get_question(self):
        pass

    def get_notes(self):
        pass

    def get_sectors(self):
        if self._game_data is None:
            raise Exception('Game data does not loaded')

        sectors_match = SECTORS_REGEX.search(self._game_data)
        if sectors_match is None:
            return

        counter = 1
        sectors = []

        for sector_row in SECTOR_ROWS_REGEX.findall(sectors_match.group()):
            sector_row = sector_row.rsplit(': ', 1)

            codes = []
            for sector_row_code in sector_row[1].split(', '):
                is_entered = sector_row_code.startswith('<')
                if is_entered:
                    hazard = BETWEEN_TAGS_REGEX.search(sector_row_code).group()
                else:
                    hazard = sector_row_code

                codes.append({
                    'number': counter,
                    'hazard': _parse_hazard(hazard),
                    'is_entered': is_entered
                })
                counter += 1

            sectors.append({
                'name': sector_row[0],
                'codes': codes
            })

        return sectors

    def get_hint(self):
        return None, None

    def get_solved_spoilers(self):
        pass
