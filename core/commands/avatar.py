import io
import os

from .command import Command

from PIL import Image, ImageDraw, ImageFont

IMAGE_SIZE = 640
IMAGE_FONT = os.path.join(
    os.path.dirname(os.path.dirname(__file__)),
    'assets',
    'fonts',
    'Roboto-Bold.ttf'
)
NICKNAME_FIT_RATIO = 0.8
NICKNAME_PADDING_RATIO = 0.1


class AvatarCommand(Command):
    name = 'avatar'

    def init(self, arguments):
        if not self.is_manager():
            return

        arguments = arguments.split()
        if len(arguments) < 3:
            self.reply_with_message(
                'Надо так - /avatar <цвета_фона> <цвет_текста> <[имена]>'
            )
            return

        background_color = arguments[0]
        font_color = arguments[1]

        for nickname in arguments[2:]:
            self.reply_with_photo(
                self._create_avatar(background_color, font_color, nickname)
            )

    def _create_avatar(self, background_color, font_color, nickname):
        nickname = nickname.upper()
        nickname = nickname.replace('Ё', 'ё')
        nickname = nickname.replace('Й', 'й')

        image = Image.new('RGBA', (IMAGE_SIZE, IMAGE_SIZE), background_color)
        image_draw = ImageDraw.Draw(image)

        nickname_width, nickname_height = image_draw.textsize(
            nickname,
            ImageFont.truetype(IMAGE_FONT, IMAGE_SIZE)
        )

        font_size = int(
            IMAGE_SIZE * NICKNAME_FIT_RATIO /
            float(nickname_width) * IMAGE_SIZE
        )
        image_font = ImageFont.truetype(IMAGE_FONT, font_size)

        nickname_width, nickname_height = image_draw.textsize(
            nickname,
            image_font
        )
        nickname_padding = int(font_size * NICKNAME_PADDING_RATIO)

        nickname_left = (IMAGE_SIZE - nickname_width) / 2
        nickname_top = (IMAGE_SIZE - nickname_height) / 2 + nickname_padding

        image_draw.text(
            (nickname_left, nickname_top),
            nickname,
            font=image_font,
            fill=font_color
        )

        image_bytes = io.BytesIO()
        image.save(image_bytes, 'PNG')

        return image_bytes.getvalue()
