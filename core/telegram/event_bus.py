class EventBus:
    def __init__(self, bot, event_classes):
        self._bot = bot
        self._event_classes = event_classes

    def handle_event(self, message):
        for event_class in self._event_classes:
            if event_class.satisfies(message):
                event_class(self._bot, message).handle()
                return True

        return False
