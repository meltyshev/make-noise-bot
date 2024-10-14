import heroku3
import os


def get_clock():
    heroku = heroku3.from_key(os.environ['HEROKU_API_KEY'])
    return heroku.app(os.environ['HEROKU_APP_NAME']).process_formation()['clock']
