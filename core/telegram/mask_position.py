class MaskPosition:
    @staticmethod
    def optional(data):
        return MaskPosition(data) if data is not None else None

    def __init__(self, data):
        self.point = data['point']
        self.x_shift = data['x_shift']
        self.y_shift = data['y_shift']
        self.scale = data['scale']
