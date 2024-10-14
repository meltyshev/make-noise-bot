import os

from apscheduler.schedulers.blocking import BlockingScheduler
# from core import heroku
from core.bot import bot
from core.game_engines import get_game_engine
from core.telegram import ClientError
from mongoengine import connect

UPDATE_INTERVAL = int(os.environ['UPDATE_INTERVAL'])

scheduler = BlockingScheduler()
connect(host=os.environ['MONGODB_URL'])


def diff(first, second):
    second_set = set(second)

    return [item for item in first if item not in second_set]


def send_message(bot, game_engine, subscriber, message, parse_mode=None):
    try:
        bot.send_message(subscriber, message, parse_mode=parse_mode)
    except ClientError as e:
        if e.response.error_code == 403:
            game_engine.unsubscribe(subscriber)


def broadcast(bot, game_engine, message, parse_mode=None):
    for subscriber in game_engine.subscribers:
        send_message(bot, game_engine, subscriber, message, parse_mode)


@scheduler.scheduled_job('interval', seconds=UPDATE_INTERVAL)
def update():
    game_engine = get_game_engine()
    if game_engine is None:
        '''
        clock = heroku.get_clock()
        if clock.quantity == 1:
            clock.scale(0)
        '''

        return

    if not game_engine.subscribers:
        return

    if not game_engine.load_game_data():
        return

    level_number = game_engine.get_level_number()
    if level_number != game_engine.level_number:
        game_engine.set_level_number(level_number)
        if level_number is not None:
            broadcast(bot, game_engine, 'АП!')

            try:
                question = game_engine.get_question()
            except:
                pass
            else:
                if question:
                    send_message(
                        bot,
                        game_engine,
                        game_engine.subscribers[0],
                        question,
                        parse_mode='HTML'
                    )

            try:
                notes = game_engine.get_notes()
            except:
                pass
            else:
                if notes:
                    send_message(
                        bot,
                        game_engine,
                        game_engine.subscribers[0],
                        notes,
                        parse_mode='HTML'
                    )

    hint = game_engine.get_hint()
    hint_number = hint[0]

    if hint_number != game_engine.hint_number:
        game_engine.set_hint_number(hint_number)
        if hint_number is not None:
            broadcast(bot, game_engine, 'Подсказка %s:\n\n%s' % (
                hint_number,
                hint[1]
            ), parse_mode='HTML')

    solved_spoilers = game_engine.get_solved_spoilers() or []
    current_solved_spoilers = game_engine.solved_spoilers
    new_solved_spoilers = diff(solved_spoilers, current_solved_spoilers)

    if new_solved_spoilers:
        game_engine.set_solved_spoilers(solved_spoilers)
        for spoiler_number in new_solved_spoilers:
            broadcast(bot, game_engine, 'Спойлер %s - АП!' % spoiler_number)


scheduler.start()
