from .photo_size import PhotoSize


class VideoNote:
    @staticmethod
    def optional(data):
        return VideoNote(data) if data is not None else None

    def __init__(self, data):
        self.file_id = data['file_id']
        self.length = data['length']
        self.duration = data['duration']
        self.thumb = PhotoSize.optional(data.get('thumb'))
        self.file_size = data.get('file_size')
