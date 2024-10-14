from .photo_size import PhotoSize


class Document:
    @staticmethod
    def optional(data):
        return Document(data) if data is not None else None

    def __init__(self, data):
        self.file_id = data['file_id']
        self.thumb = PhotoSize.optional(data.get('thumb'))
        self.file_name = data.get('file_name')
        self.mime_type = data.get('mime_type')
        self.file_size = data.get('file_size')
