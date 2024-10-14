from .file import File
from .message import Message
from .user import User

BASE_URL = 'https://api.telegram.org/bot%s'
FILE_BASE_URL = 'https://api.telegram.org/file/bot%s'


class ClientError(Exception):
    def __init__(self, response):
        super().__init__(response.description)
        self.response = response


class Client:
    def __init__(self, api_token, request_interface):
        self._api_token = api_token
        self._request_interface = request_interface

    def get_me(self):
        return User(self._get('getMe'))

    def set_webhook(self, url):
        self._post('setWebhook', {
            'url': url
        })

    def send_message(self, chat_id, text, parse_mode=None, disable_web_page_preview=None, disable_notification=None, reply_to_message_id=None, reply_markup=None):
        data = {
            'chat_id': chat_id,
            'text': text
        }
        if parse_mode is not None:
            data['parse_mode'] = parse_mode
        if disable_web_page_preview is not None:
            data['disable_web_page_preview'] = disable_web_page_preview
        if disable_notification is not None:
            data['disable_notification'] = disable_notification
        if reply_to_message_id is not None:
            data['reply_to_message_id'] = reply_to_message_id
        if reply_markup is not None:
            data['reply_markup'] = reply_markup.to_dict()

        return Message(self._post('sendMessage', data))

    def send_photo(self, chat_id, photo, caption=None, parse_mode=None, disable_notification=None, reply_to_message_id=None, reply_markup=None):
        data = {
            'chat_id': chat_id
        }
        files = {}

        if isinstance(photo, str):
            data['photo'] = photo
        else:
            files['photo'] = photo
        if caption is not None:
            data['caption'] = caption
        if parse_mode is not None:
            data['parse_mode'] = parse_mode
        if disable_notification is not None:
            data['disable_notification'] = disable_notification
        if reply_to_message_id is not None:
            data['reply_to_message_id'] = reply_to_message_id
        if reply_markup is not None:
            data['reply_markup'] = reply_markup.to_dict()

        return Message(self._post('sendPhoto', data, files))

    def send_audio(self, chat_id, audio, caption=None, parse_mode=None, disable_notification=None, reply_to_message_id=None, reply_markup=None):
        data = {
            'chat_id': chat_id
        }
        files = {}

        if isinstance(audio, str):
            data['audio'] = audio
        else:
            files['audio'] = audio
        if caption is not None:
            data['caption'] = caption
        if parse_mode is not None:
            data['parse_mode'] = parse_mode
        if disable_notification is not None:
            data['disable_notification'] = disable_notification
        if reply_to_message_id is not None:
            data['reply_to_message_id'] = reply_to_message_id
        if reply_markup is not None:
            data['reply_markup'] = reply_markup.to_dict()

        return Message(self._post('sendAudio', data, files))

    def send_document(self, chat_id, document, caption=None, parse_mode=None, disable_notification=None, reply_to_message_id=None, reply_markup=None):
        data = {
            'chat_id': chat_id
        }
        files = {}

        if isinstance(document, str):
            data['document'] = document
        else:
            files['document'] = document
        if caption is not None:
            data['caption'] = caption
        if parse_mode is not None:
            data['parse_mode'] = parse_mode
        if disable_notification is not None:
            data['disable_notification'] = disable_notification
        if reply_to_message_id is not None:
            data['reply_to_message_id'] = reply_to_message_id
        if reply_markup is not None:
            data['reply_markup'] = reply_markup.to_dict()

        return Message(self._post('sendDocument', data, files))

    def send_video(self, chat_id, video, caption=None, parse_mode=None, disable_notification=None, reply_to_message_id=None, reply_markup=None):
        data = {
            'chat_id': chat_id
        }
        files = {}

        if isinstance(video, str):
            data['video'] = video
        else:
            files['video'] = video
        if caption is not None:
            data['caption'] = caption
        if parse_mode is not None:
            data['parse_mode'] = parse_mode
        if disable_notification is not None:
            data['disable_notification'] = disable_notification
        if reply_to_message_id is not None:
            data['reply_to_message_id'] = reply_to_message_id
        if reply_markup is not None:
            data['reply_markup'] = reply_markup.to_dict()

        return Message(self._post('sendVideo', data, files))

    def send_voice(self, chat_id, voice, caption=None, parse_mode=None, disable_notification=None, reply_to_message_id=None, reply_markup=None):
        data = {
            'chat_id': chat_id
        }
        files = {}

        if isinstance(voice, str):
            data['voice'] = voice
        else:
            files['voice'] = voice
        if caption is not None:
            data['caption'] = caption
        if parse_mode is not None:
            data['parse_mode'] = parse_mode
        if disable_notification is not None:
            data['disable_notification'] = disable_notification
        if reply_to_message_id is not None:
            data['reply_to_message_id'] = reply_to_message_id
        if reply_markup is not None:
            data['reply_markup'] = reply_markup.to_dict()

        return Message(self._post('sendVoice', data, files))

    def send_video_note(self, chat_id, video_note, disable_notification=None, reply_to_message_id=None, reply_markup=None):
        data = {
            'chat_id': chat_id
        }
        files = {}

        if isinstance(video_note, str):
            data['video_note'] = video_note
        else:
            files['video_note'] = video_note
        if disable_notification is not None:
            data['disable_notification'] = disable_notification
        if reply_to_message_id is not None:
            data['reply_to_message_id'] = reply_to_message_id
        if reply_markup is not None:
            data['reply_markup'] = reply_markup.to_dict()

        return Message(self._post('sendVideoNote', data, files))

    def send_location(self, chat_id, latitude, longitude, live_period=None, disable_notification=None, reply_to_message_id=None, reply_markup=None):
        data = {
            'chat_id': chat_id,
            'latitude': latitude,
            'longitude': longitude
        }
        if live_period is not None:
            data['live_period'] = live_period
        if disable_notification is not None:
            data['disable_notification'] = disable_notification
        if reply_to_message_id is not None:
            data['reply_to_message_id'] = reply_to_message_id
        if reply_markup is not None:
            data['reply_markup'] = reply_markup.to_dict()

        return Message(self._post('sendLocation', data))

    def send_sticker(self, chat_id, sticker, disable_notification=None, reply_to_message_id=None, reply_markup=None):
        data = {
            'chat_id': chat_id
        }
        files = {}

        if isinstance(sticker, str):
            data['sticker'] = sticker
        else:
            files['sticker'] = sticker
        if disable_notification is not None:
            data['disable_notification'] = disable_notification
        if reply_to_message_id is not None:
            data['reply_to_message_id'] = reply_to_message_id
        if reply_markup is not None:
            data['reply_markup'] = reply_markup.to_dict()

        return Message(self._post('sendSticker', data, files))

    def get_file(self, file_id):
        return File(self._post('getFile', {
            'file_id': file_id
        }), FILE_BASE_URL % self._api_token)

    def leave_chat(self, chat_id):
        return self._post('leaveChat', {
            'chat_id': chat_id
        })

    def _get(self, method):
        return self._get_response_result(self._request_interface.get(self._make_url(method)))

    def _post(self, method, data, files={}):
        return self._get_response_result(self._request_interface.post(self._make_url(method), data, files))

    def _make_url(self, method):
        return '%s/%s' % (BASE_URL % self._api_token, method)

    def _get_response_result(self, response):
        if not response.ok:
            raise ClientError(response)

        return response.result
