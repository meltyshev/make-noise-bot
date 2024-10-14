import os
import traceback

from core.bot import bot
from flask import Flask, render_template, request, url_for
from mongoengine import connect

app = Flask(__name__)
connect(host=os.environ['MONGODB_URL'])


@app.route('/%s' % bot.api_token, methods=('POST',))
def webhook():
    try:
        bot.handle(request.get_json())
    except:  # debug
        bot.send_message_to_admin(traceback.format_exc())

    return 'ok'


@app.route('/set-webhook')
def set_webhook():
    bot.set_webhook(url_for('webhook', _external=True))

    return 'ok'


@app.route('/')
def index():
    return 'ok'
