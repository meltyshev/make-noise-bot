from .command import Command

from core.models import Chat
from core.telegram import KeyboardButton, ReplyKeyboardMarkup, ReplyKeyboardRemove


class ChatsCommand(Command):
    name = 'chats'

    def init(self, arguments):
        if not self.is_admin():
            return

        self._reply_with_main_menu()

    def handle(self, state):
        action = state.get('action')

        '''if action == 'add':
            return self._handle_add_action(state)'''

        if 'id' not in state:
            return self._handle_main_menu()

        chat = Chat.get(state['id'])
        if chat is None:
            self.reply_with_message('Текущий чат уже удален')
            return self._reply_with_main_menu()

        if action is None:
            return self._handle_actions_menu(state, chat)

    def _handle_main_menu(self):
        '''if self.message.text == 'Добавить':
            self.set_last_command({
                'action': 'add'
            })
            return self.reply_with_message('Введите название', reply_markup=ReplyKeyboardRemove())'''

        if self.message.text == 'Закрыть':
            return self._close()

        parts = self.message.text.rsplit(' | ', 1)
        if len(parts) != 2:
            self.reply_with_message('Действие не найдено')
            return self._reply_with_main_menu()

        chat = Chat.get(int(parts[1]))
        if chat is None:
            self.reply_with_message('Команда не найдена')
            return self._reply_with_main_menu()

        self._reply_with_actions_menu(chat)

    def _handle_actions_menu(self, state, chat):
        if self.message.text == 'Назад':
            return self._reply_with_main_menu()

        if self.message.text == 'Закрыть':
            return self._close()

        if self.message.text == 'Удалить':
            chat.delete()

            self.reply_with_message('Чат удален')
            return self._reply_with_main_menu()

        self.reply_with_message('Действие не найдено')
        self._reply_with_actions_menu(team)

    '''def _handle_add_action(self, state):
        if 'name' not in state:
            if self.message.text in ('Добавить', 'Закрыть'):
                return self.reply_with_message('Такое название зарезервировано, введите другое')

            self.set_last_command({
                **state,
                'name': self.message.text
            })
            return self.reply_with_message('Введите идентификатор')

        if Chat.get(self.message.text):
            return self.reply_with_message('Такой идентификатор уже существует, введите другой')

        chat = Chat.create(self.message.text, name=state['name'])
        if chat is None:
            self.reply_with_message('Чат уже существует')
            return self._reply_with_main_menu()

        self.reply_with_message('Чат добавлен, идентификатор - %s' % chat.key)
        self._reply_with_main_menu()'''

    def _reply_with_main_menu(self):
        self.set_last_command({})
        self.reply_with_message('Выберите чат', reply_markup=ReplyKeyboardMarkup(
            [
                [KeyboardButton('%s | %s' % (chat, chat.key))] for chat in Chat.all()
            ] + [
                [KeyboardButton('Закрыть')]
            ]
        ))

    def _reply_with_actions_menu(self, chat):
        self.set_last_command({
            'id': chat.key
        })
        self.reply_with_message(
            '%s | %s\n\nТип: %s\nСтатус: %s\n\nВыберите действие' % (
                chat,
                chat.key,
                chat.type,
                chat.permission
            ),
            reply_markup=ReplyKeyboardMarkup([
                [KeyboardButton('Удалить')],
                [KeyboardButton('Назад'), KeyboardButton('Закрыть')]
            ])
        )

    def _close(self):
        self.delete_last_command()
        self.reply_with_message('Настройка чатов завершена', reply_markup=ReplyKeyboardRemove())
