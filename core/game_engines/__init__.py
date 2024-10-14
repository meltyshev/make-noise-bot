from .dozor_classic import DozorClassic
from .dozor_classic_prequel import DozorClassicPrequel
from .dozor_lite import DozorLite
from .dozor_lite_prequel import DozorLitePrequel

from core.models.game import Game
from core.models.game_config import GameConfig

GAME_ENGINE_CLASSES = (DozorClassic, DozorLite, DozorClassicPrequel, DozorLitePrequel)


def start_game_engine():
    config = GameConfig.get_or_create()
    for game_engine_class in GAME_ENGINE_CLASSES:
        if game_engine_class.name == config.engine:
            return game_engine_class.start(config)


def get_game_engine():
    game = Game.get()
    if game is None:
        return

    for game_engine_class in GAME_ENGINE_CLASSES:
        if game_engine_class.name == game.engine:
            return game_engine_class(game)
