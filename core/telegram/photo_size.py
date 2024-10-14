class PhotoSize:
    @staticmethod
    def optional(data):
        return PhotoSize(data) if data is not None else None

    def __init__(self, data):
        self.file_id = data['file_id']
        self.width = data['width']
        self.height = data['height']
        self.file_size = data.get('file_size')
