from .mask_position import MaskPosition
from .photo_size import PhotoSize


class Sticker:
    @staticmethod
    def optional(data):
        return Sticker(data) if data is not None else None

    def __init__(self, data):
        self.file_id = data['file_id']
        self.width = data['width']
        self.height = data['height']
        self.thumb = PhotoSize.optional(data.get('thumb'))
        self.emoji = data.get('emoji')
        self.set_name = data.get('set_name')
        self.mask_position = MaskPosition.optional(data.get('mask_position'))
        self.file_size = data.get('file_size')
