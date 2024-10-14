import os

from .commands import *
from .events import *
from .last_command_storage_interface import LastCommandStorageInterface
from .request_interface import RequestInterface
from .statements import *
from .telegram import Bot

bot = Bot(
    os.environ['TELEGRAM_API_TOKEN'],
    int(os.environ['TELEGRAM_ADMIN_ID']),
    RequestInterface(),
    LastCommandStorageInterface(),
    command_classes=(
        NumbersToLettersCommand,
        LettersToNumbersCommand,
        IntersectionCommand,
        MorseCommand,
        HintCommand,
        AnagramCommand,
        MaskCommand,
        LinkCommand,
        QuestionCommand,
        NotesCommand,
        RatingCommand,
        ClearRatingCommand,
        TopCommand,
        MaxwellCommand,
        RomkaCommand,
        CancelCommand,
        HelpCommand,
        GoCommand,
        StartCommand,
        PermissionCommand,
        CoordinatesCommand,
        ChatIdCommand,
        UserIdCommand,
        AvatarCommand,
        GameCommand,
        CodeFormatsCommand,
        RestrictCommand,
        BruteForceCommand,
        # BroadcastCommand,
        SubscribeCommand,
        PinLevelCommand,
        UnpinLevelCommand,
        AddSpotCommand,
        DeleteSpotCommand,
        GameConfigCommand,
        ChatsCommand,
        AllowCommand,
        ForbidCommand,
        DropCommand,
        WriteCommand,
        ConfigCommand,
        TestCommand
    ),
    event_classes=(
        LeaveModeEvent,
    ),
    statement_classes=(
        InformationStatement,
        EnterPinnedLevelCodeStatement,
        EnterCodeStatement
    )
)
