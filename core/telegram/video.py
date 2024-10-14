from .photo_size import PhotoSize


class Video:
    @staticmethod
    def optional(data):
        return Video(data) if data is not None else None

    def __init__(self, data):
        self.file_id = data['file_id']
        self.width = data['width']
        self.height = data['height']
        self.duration = data['duration']
        self.thumb = PhotoSize.optional(data.get('thumb'))
        self.mime_type = data.get('mime_type')
        self.file_size = data.get('file_size')
