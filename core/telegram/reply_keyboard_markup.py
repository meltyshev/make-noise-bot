class ReplyKeyboardMarkup:
    def __init__(self, keyboard, resize_keyboard=None, one_time_keyboard=None, selective=None):
        self.keyboard = keyboard
        self.resize_keyboard = resize_keyboard
        self.one_time_keyboard = one_time_keyboard
        self.selective = selective

    def to_dict(self):
        keyboard = []
        for keybord_row in self.keyboard:
            keyboard.append([keybord_button.to_dict() for keybord_button in keybord_row])

        data = {
            'keyboard': keyboard
        }
        if self.resize_keyboard is not None:
            data['resize_keyboard'] = self.resize_keyboard
        if self.one_time_keyboard is not None:
            data['one_time_keyboard'] = self.one_time_keyboard
        if self.selective is not None:
            data['selective'] = self.selective

        return data
