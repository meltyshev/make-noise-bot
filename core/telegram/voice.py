class Voice:
    @staticmethod
    def optional(data):
        return Voice(data) if data is not None else None

    def __init__(self, data):
        self.file_id = data['file_id']
        self.duration = data['duration']
        self.mime_type = data.get('mime_type')
        self.file_size = data.get('file_size')
