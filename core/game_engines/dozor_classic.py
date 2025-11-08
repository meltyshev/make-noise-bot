import re
import requests

from .game_engine import EnterCodeResult, GameEngine, BETWEEN_TAGS_REGEX, REQUEST_TIMEOUT, REQUEST_TIMEOUT_MESSAGE, UNKNOWN_ERROR_MESSAGE, USER_AGENT

from bs4 import BeautifulSoup
from core.utils.html_to_text import html_to_text
from urllib.parse import parse_qs, urlparse

STATUSES = {
    1: 'Игра не началась.',
    2: 'Неверный PIN.',
    3: 'Авторизация пройдена успешно.',
    4: 'Не введен код.',
    5: 'Время на задание вышло. Решайте следующее задание.',
    7: '\u26a0\ufe0f Вы уже вводили этот код.',
    8: '\u2705 Код принят. Ищите следующий составной код.',
    9: '\u2705 Код принят. Выполняйте следующее задание.',
    10: 'Спасибо за игру. Игра закончена.',
    11: '\u274c Код не принят.',
    12: 'Вы вводили неверный код больше 4 раз. Прием данных от вас заблокирован на три минуты. Повторите попытку позже.',
    13: 'Движок остановлен организатором.',
    14: 'Игра вашей команды приостановлена.',
    15: 'Вам не запланировано следующее задание. Организатор видит, что вы бездействуете и назначит вам уровень в ближайшее время. Периодически обновляйте страницу. Время за задержку будет вычтено из вашего результата. Если новое задание не будет выдаваться длительное время, свяжитесь с организатором.',
    16: '\u2705 Код принят.',
    17: 'Время на отправку кода вышло.',
    21: 'Акаунт заблокирован или не активирован. Выполните инструкции, высланные вам в письме-подтверждении.',
    22: 'Авторизация пройдена успешно.',
    23: 'Неверный пароль.',
    24: 'Неизвестный пользователь.',
    25: 'Ошибка авторизации.',
    26: 'Вы уже взяли 2 перерыва. Больше вы не можете приостанавливать игру своей команды.',
    27: 'Игра вашей команды приостановлена на 15 минут по решению штаба.',
    28: 'Вы уже дважды отказывались от заданий. Больше вы не можете завершать задания досрочно.',
    29: 'Вам уже выдано следующее задание.',
    30: 'Вам не запланировано следующее задание. Чтобы отказаться от текущего, свяжитесь с организатором и попросите его назначить вам следующий уровень.',
    31: '\u274c Код к сквозному бонусному заданию не принят.',
    32: '\u2705 Код к сквозному бонусному заданию принят.',
    33: 'Ваше сообщение отправлено организатору.',
    34: '\u2705 Код к сквозному бонусному заданию принят.',
    35: '\u26a0\ufe0f Вы уже вводили этот код.',
    36: '\u2705 Код принят.',
    37: '\u2705 Код к сквозному бонусному заданию принят.',
    38: '\u274c Код не принят. Вы превысили лимит попыток неправильного ввода кода. Предыдущее задание считается невзятым. Вам выдано следующее задание.',
    39: 'Вы решили отказаться от выполнения задания. Вам выдано новое задание.',
    40: 'Это ложный код. За его нахождение ваша команда получила штраф.',
    41: 'Вы нашли все основные коды. Вы можете продолжать искать бонусные коды или перейти на следующий уровень.',
    42: 'За досрочное использование подсказки вам начислен штраф.',
    43: 'Вы ввели не обязательный основной код, задание уже считается выполненным. Ищите бонусные коды.',
    44: 'В этом задании нельзя использовать универсальный код.',
    45: 'Вы уже использовали универсальный код, повторное его использование невозможно.',
    46: 'В бонусном сквозном задании нельзя использовать универсальный код.',
    47: '\u2705 Универсальный код принят.',
    48: '\u2705 Универсальный код принят. Выполняйте следующее задание.',
    49: '\u2705 Универсальный код принят.',
    50: '\u2705 Мастер-код принят.',
    51: '\u2705 Мастер-код принят. Выполняйте следующее задание.',
    52: '\u2705 Мастер-код принят.',
    53: '\u26a0\ufe0f Данный код уже был найден. Ищите другой код.',
    54: '\u26a0\ufe0f Код к сквозному заданию не принят, так как истекло время его выполнения.',
    55: '\u2705 Код к спойлеру принят.',
    56: '\u274c Код к спойлеру не принят.',
    57: 'У вас недостаточно прав для использования данной функции.',
    58: 'Игра еще не началась.',
}

ACCEPTED_STATUS_CODES = (8, 9, 16, 36, 55)

FIX_RESPONSE_REGEX = re.compile(r'^([^\{\[]*)?')


def _get_status_code(location):
    try:
        return int(parse_qs(urlparse(location).query)['err'][0])
    except:
        pass


def _parse_hazard(hazard):
    hazard = hazard.split(':', 1)[0]
    if hazard == 'null':
        hazard = 'N'

    return hazard


class DozorClassic(GameEngine):
    name = 'DozorClassic'

    @staticmethod
    def obtain_session(config):
        url = 'https://classic.dzzzr.ru/%s/API/login.php' % config.city
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
        session = DozorClassic.obtain_session(config)
        if session is None:
            return

        return GameEngine.start(
            DozorClassic,
            config,
            login=config.login,
            password=config.password,
            session=session
        )

    @staticmethod
    def get():
        return GameEngine.get(DozorClassic)

    def __init__(self, game):
        super().__init__(game)
        self.link = 'https://classic.dzzzr.ru/%s/go/' % self.game.city
        self._game_data = None

    @property
    def pinned_level_number(self):
        return self.game.pinned_level_number

    def pin_level(self, level_number):
        self.game.pinned_level_number = level_number
        self.game.save()

    # TODO: pinned_level_number?
    def enter_code(self, code, pinned_level_number=None):
        data = {
            'action': 'entcod',
            'cod': code.encode('windows-1251')
        }
        if pinned_level_number is not None:
            data.update({
                'level': pinned_level_number,
                'skvoz': 1
            })

        try:
            response = requests.post(
                self.link,
                data=data,
                headers={
                    'Cookie': 'dozorSiteSession=%s' % self.game.session,
                    'Referer': self.link,
                    'User-Agent': USER_AGENT
                },
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
                params={
                    's': self.game.session,
                    'api': 'true'
                },
                timeout=REQUEST_TIMEOUT,
            )
        except requests.exceptions.RequestException:
            return False

        if response.status_code != 200:
            return False

        # TODO: fuck him
        import simplejson as json
        content = response.text.replace('null:{', '"null":{')
        content = FIX_RESPONSE_REGEX.sub('', content, 1)
        self._game_data = json.loads(content)

        return True

    def get_level_number(self):
        if self._game_data is None:
            raise Exception('Game data does not loaded')

        level = self._game_data['level']
        if not level:
            return

        return int(level['levelNumber'])

    def get_progress(self):
        if self._game_data is None:
            raise Exception('Game data does not loaded')

        level = self._game_data['level']
        if not level:
            return

        if int(level['neededCodes']) == 0:
            codes_needed = level['totalCodes']
        else:
            codes_needed = level['neededCodes']

        progress = ['%s/%s' % (level['codesFounded'], codes_needed)]

        if level['bonusCodesTotal']:
            progress.append('%s/%s' % (
                level['bonusCodesFounded'],
                level['bonusCodesTotal']
            ))

        progress.append(level['timeOnLevel'])

        return ' '.join(progress)

    def get_question(self):
        if self._game_data is None:
            raise Exception('Game data does not loaded')

        level = self._game_data['level']
        if not level:
            return

        return self._html_to_text(level['question'])

    def get_notes(self):
        if self._game_data is None:
            raise Exception('Game data does not loaded')

        level = self._game_data['level']
        if not level:
            return

        return self._html_to_text(level['locationComment'])

    def get_sectors(self):
        if self._game_data is None:
            raise Exception('Game data does not loaded')

        level = self._game_data['level']
        if not level:
            return

        main_counter = 1
        main_sectors = []

        bonus_counter = 1
        bonus_sectors = []

        for sector_row in level['koline'].split('<br>')[:-1]:
            sector_row = sector_row.rsplit(': ', 1)

            name = sector_row[0].lstrip()
            name = name.replace(':', ',')
            name = name.replace('  ', ' ')

            is_main_sector = name.endswith('основные коды')

            name = name.capitalize()

            codes = []
            for sector_row_code in sector_row[1].split(', '):
                if is_main_sector:
                    number = main_counter
                    main_counter += 1
                else:
                    number = bonus_counter
                    bonus_counter += 1

                is_entered = sector_row_code.startswith('<')
                if is_entered:
                    hazard = BETWEEN_TAGS_REGEX.search(sector_row_code).group()
                else:
                    hazard = sector_row_code

                codes.append({
                    'number': number,
                    'hazard': _parse_hazard(hazard),
                    'is_entered': is_entered
                })

            (main_sectors if is_main_sector else bonus_sectors).append({
                'name': name,
                'codes': codes
            })

        return main_sectors + bonus_sectors

    def get_hint(self):
        if self._game_data is None:
            raise Exception('Game data does not loaded')

        level = self._game_data['level']
        if not level:
            return None, None

        for i in reversed(range(1, 3)):
            hint = level['hint%s' % i]
            if hint:
                return i, self._html_to_text(hint)

        return None, None

    def get_solved_spoilers(self):
        if self._game_data is None:
            raise Exception('Game data does not loaded')

        level = self._game_data['level']
        if not level:
            return None

        solved_spoilers = []
        for spoiler in level['spoilers']:
            if spoiler['spoilerSolved']:
                solved_spoilers.append(int(spoiler['spoilerNumber']))

        return solved_spoilers

    def _html_to_text(self, html):
        return html_to_text(html, self.link)
