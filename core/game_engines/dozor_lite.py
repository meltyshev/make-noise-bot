import re
import requests
import urllib3

from .game_engine import EnterCodeResult, GameEngine, BETWEEN_TAGS_REGEX, REQUEST_TIMEOUT, REQUEST_TIMEOUT_MESSAGE, UNKNOWN_ERROR_MESSAGE, USER_AGENT

from core.utils.html_to_text import html_to_text
from urllib.parse import parse_qs, urlparse

STATUSES = {
    1: 'Игра не началась.',
    2: 'Неверный PIN.',
    3: 'Авторизация пройдена успешно.',
    4: 'Не введен код.',
    5: 'Время на отправку кода вышло. Для получения следующего задания обновите страницу.',
    6: 'Время на отправку кода вышло. Вы прошли все уровни.',
    7: '\u26a0\ufe0f Код не принят. Вы уже ввели этот составной код. Ищите другой.',
    8: '\u2705 Код принят. Ищите следующий составной код.',
    9: '\u2705 Код принят. Выполняйте следующее задание.',
    10: 'Игра закончена. Вы прошли все уровни.',
    11: '\u274c Код не принят.',
    12: 'Вы вводили неверный код больше 4 раз. Прием данных от Вас заблокирован на три минуты. Повторите попытку позже.',
    13: 'Движок остановлен организатором.',
    14: 'Игра вашей команды приостановлена.',
    15: 'Вам не запланировано следующее задание.',
    16: '\u26a0\ufe0f Код не принят. Вы уже ввели этот код.',
    17: '\u2705 Код принят.',
    18: 'Вы не можете взять подсказку.',
    19: '\u2705 Принят финишный код.',
    20: '\u26a0\ufe0f Код не принят. Время, отведенное на игру, вышло.',
    21: 'Неверная авторизация.',
    22: 'Заявка на игру не подана или не принята.',
    23: '\u26a0\ufe0f Вы пытаетесь ввести уже принятый бонусный код.',
    24: '\u2705 Принят бонусный код.',
    25: '\u2705 Принят бонусный код. Вам начислено дополнительное бонусное время на нахождение всех бонусных кодов.',
    26: 'Это ложный код. За его нахождение ваша команда получила штраф.',
    27: '\u274c Код не принят. Вы превысили лимит попыток неправильного ввода кода. Предыдущее задание считается невзятым. Вам выдано следующее задание.',
    28: 'Вы превысили лимит попыток неправильного ввода кода. Предыдущее задание считается невзятым. Вам выдано следующее задание.',
    29: 'Это ложный код. За его нахождение ваша команда получила штраф. Вы превысили лимит попыток неправильного ввода кода. Предыдущее задание считается невзятым. Вам выдано следующее задание.',
    30: '\u26a0\ufe0f Универсальный код не принят. Вы уже ввели этот универсальный код. Ищите другой.',
    31: '\u26a0\ufe0f Мастер-код не принят. Вы уже ввели этот мастер-код. Ищите другой.',
    32: '\u2705 Универсальный код принят. Ищите следующий составной код.',
    33: '\u2705 Мастер-код принят. Ищите следующий составной код.',
    34: '\u2705 Универсальный код принят. Выполняйте следующее задание.',
    35: '\u2705 Мастер-код принят. Выполняйте следующее задание.',
    36: 'В этом задании нельзя использовать универсальный код.',
    37: 'Вы нашли все основные коды. Вы можете продолжать искать бонусные коды или перейти на следующий уровень.',
    38: 'Вы не можете отправлить основной код после ввода финишного.',
    39: '\u26a0\ufe0f Код к сквозному заданию не принят, так как истекло время его выполнения.',
    40: '\u2705 Код принят.',
    41: '\u274c Код к спойлеру не принят.',
    42: 'Подсказка выдана.',
    43: '\u26a0\ufe0f Данный код уже был найден вашей или другой командой. Ищите другой код.'
}

ACCEPTED_STATUS_CODES = (8, 9, 17, 24, 25, 40)

BETWEEN_BRACKETS_REGEX = re.compile(r'(?<=\().*?(?=\))')
LEVEL_NUMBER_REGEX = re.compile(r'(?<=<!--levelNumberBegin-->)\d+(?=<!--levelNumberEnd-->)')
QUESTION_REGEX = re.compile(r'(?<=<!--levelTextBegin-->).*?(?=<!--levelTextEnd-->)', re.DOTALL)
SECTORS_REGEX = re.compile(r'(?<=<strong>Коды сложности</strong>).*?(?=</div>)')
SECTOR_ROWS_REGEX = re.compile(r'(?<=<br>).*?(?=<br>)')
PROGRESS_REGEX = re.compile(r'(?<=\(Всего - )(\d+) ?(, для прохождения достаточно любых (\d+) ?)?, принято - (\d+)(?=\))')
TIME_ON_LEVEL_REGEX = re.compile(r'(?<=<!--timeOnLevelBegin )\d+(?= timeOnLevelEnd-->)')
HINTS_REGEX = re.compile(r'(?<=<!--LevelClue)(\d)Text-->(.*?)<!--LevelClue\d(?=TextEnd-->)', re.DOTALL)

http = urllib3.PoolManager()


def _get_status_code(location):
    try:
        return int(parse_qs(urlparse(location).query)['err'][0])
    except:
        pass


def _parse_hazard(hazard):
    if hazard == 'null':
        hazard = 'N'

    return hazard


class DozorLite(GameEngine):
    name = 'DozorLite'

    @staticmethod
    def start(config):
        return GameEngine.start(
            DozorLite,
            config,
            pincode=config.pincode
        )

    @staticmethod
    def get():
        return GameEngine.get(DozorLite)

    def __init__(self, game):
        super().__init__(game)
        self.link = 'https://lite.dzzzr.ru/%s/go/?pin=%s' % (
            self.game.city,
            self.game.pincode
        )
        self._headers = {'User-Agent': USER_AGENT}
        self._game_data = None

    def enter_code(self, code):
        try:
            response = http.request(
                'POST',
                self.link,
                fields={
                    'action': 'entcod',
                    'cod': code.encode('windows-1251')
                },
                headers=self._headers,
                timeout=REQUEST_TIMEOUT,
                redirect=False
            )
        except urllib3.exceptions.HTTPError:
            return EnterCodeResult(REQUEST_TIMEOUT_MESSAGE)

        if response.status == 302:
            status_code = _get_status_code(response.headers['Location'])

            return EnterCodeResult(
                STATUSES.get(status_code, UNKNOWN_ERROR_MESSAGE),
                status_code=status_code,
                is_accepted=status_code in ACCEPTED_STATUS_CODES
            )

        # elif response.status == 401:
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

        self._game_data = response.content.decode('windows-1251').split('<!--BonusLevels-->', 1)[0] # TODO: ...
        return True

    def get_level_number(self):
        if self._game_data is None:
            raise Exception('Game data does not loaded')

        level_number = LEVEL_NUMBER_REGEX.search(self._game_data)
        if level_number is None:
            return

        return int(level_number.group())

    def get_progress(self):
        if self._game_data is None:
            raise Exception('Game data does not loaded')

        progress = []

        progress_match = PROGRESS_REGEX.search(self._game_data)
        if progress_match is not None:
            if progress_match.group(3) is None:
                codes_needed = progress_match.group(1)
            else:
                codes_needed = progress_match.group(3)

            progress.append('%s/%s' % (progress_match.group(4), codes_needed))

        time_on_level_match = TIME_ON_LEVEL_REGEX.search(self._game_data)
        if time_on_level_match is not None:
            minutes, seconds = divmod(int(time_on_level_match.group()), 60)
            hours, minutes = divmod(minutes, 60)

            progress.append('%02d:%02d:%02d' % (hours, minutes, seconds))

        if not progress:
            return

        return ' '.join(progress)

    def get_question(self):
        if self._game_data is None:
            raise Exception('Game data does not loaded')

        question_match = QUESTION_REGEX.search(self._game_data)
        if question_match is None:
            return

        return self._html_to_text(question_match.group())

    def get_notes(self):
        pass

    def get_sectors(self):
        if self._game_data is None:
            raise Exception('Game data does not loaded')

        sectors_match = SECTORS_REGEX.search(self._game_data)
        if sectors_match is None:
            return

        main_counter = 1
        main_sectors = []

        bonus_counter = 1
        bonus_sectors = []

        for sector_row in SECTOR_ROWS_REGEX.findall(sectors_match.group()):
            sector_row = sector_row.rsplit(': ', 1)
            if len(sector_row) != 2:
                continue

            name = sector_row[0].lstrip()
            name = name.replace(':', ',')
            name = name.replace('  ', ' ')

            is_main_sector = name.endswith('основные коды')
            is_nonstandart = sector_row[1].startswith('<br />')

            name = name.capitalize()

            if is_nonstandart:
                sector_row_codes = sector_row[1].split('<br />')[1:]
            else:
                sector_row_codes = sector_row[1].split(', ')

            codes = []
            for sector_row_code in sector_row_codes:
                if is_main_sector:
                    number = main_counter
                    main_counter += 1
                else:
                    number = bonus_counter
                    bonus_counter += 1

                is_entered = sector_row_code.startswith('<')
                if is_nonstandart:
                    hazard = BETWEEN_BRACKETS_REGEX.search(sector_row_code).group()
                elif is_entered:
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

        hints = HINTS_REGEX.findall(self._game_data)
        if not hints:
            return None, None

        hint = hints[-1]
        return int(hint[0]), self._html_to_text(hint[1])

    def get_solved_spoilers(self):
        pass

    def _html_to_text(self, html):
        return html_to_text(html, self.link)
