class File:
    def __init__(self, data, file_base_url):
        self.file_id = data['file_id']
        self.file_size = data.get('file_size')
        self.file_url = '%s/%s' % (
            file_base_url,
            data['file_path']
        ) if 'file_path' else None
