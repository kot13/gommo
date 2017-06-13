const COMMAND_PLAYER_MOVE = "player_move";
const COMMAND_PLAYER_ROTATION = "player_rotation";

function notifyPlayerMoved(gamePlayer, direction) {
    socket.emit(COMMAND_PLAYER_MOVE, direction);
    addMotionCommand(gamePlayer, COMMAND_PLAYER_MOVE + "_" + direction)
}

function notifyPlayerRotated(gamePlayer) {
    socket.emit(COMMAND_PLAYER_ROTATION, String(gamePlayer.player.rotation));
    addMotionCommand(gamePlayer, COMMAND_PLAYER_ROTATION);
}

function addMotionCommand(gamePlayer, what) {
    let player = gamePlayer.player;
    addPlayerCommand(gamePlayer, what, {
        x: player.x,
        y: player.y,
        rotation: player.rotation
    })
}

function addPlayerCommand(gamePlayer, what, result) {
    getExecutedCommands(gamePlayer).push({
        what: what,
        result: result
    })
}

function getExecutedCommands(gamePlayer) {
    if (gamePlayer.executedCommands === undefined) {
        gamePlayer.executedCommands = []
    }
    return gamePlayer.executedCommands
}