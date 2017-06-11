function checkPredictions(gamePlayer, dataPlayer, dataCommand) {
    if (!checkCommandAlreadyExecuted(gamePlayer, dataCommand)) {
        updatePlayerRotation(gamePlayer, dataPlayer);
        updatePlayerPosition(gamePlayer, dataPlayer);
    }

    gamePlayer.lastServerCommand = dataCommand.when
}

function checkCommandAlreadyExecuted(gamePlayer, dataCommand) {
    if (gamePlayer.lastServerCommand !== undefined &&
        gamePlayer.lastServerCommand === dataCommand.when) {
        return true;
    } else {
        let gamePlayerCommands = gamePlayer.executedCommands;
        for (let i = 0; i < gamePlayerCommands.length; i++) {
            if ((gamePlayerCommands[i].what === dataCommand.what &&
                isResultEqual(gamePlayerCommands[i].result, dataCommand.result))) {
                gamePlayerCommands.splice(0, i + 1);
                return true;
            }
        }
        return false;
    }
}

function isResultEqual(result1, result2) {
    return result1.x === result2.x && result1.y === result2.y && fixRotation(result1.rotation) === fixRotation(result2.rotation)
}

function updatePlayerRotation(gamePlayer, dataPlayer) {
    if (gamePlayer.rotationTween !== undefined) {
        gamePlayer.rotationTween.stop();
    }

    let player = gamePlayer.player;
    let delta = getShortestAngle(Phaser.Math.radToDeg(dataPlayer.rotation), player.angle);
    if (Math.abs(delta) <= 5) {
        player.rotation = Number(dataPlayer.rotation)
    } else {
        let degrees = player.angle + delta;
        gamePlayer.rotationTween = game.add.tween(player).to({angle: degrees}, Math.abs(delta), Phaser.Easing.Linear.None);
        gamePlayer.rotationTween.start()
    }
}

function getShortestAngle(angle1, angle2) {
    let difference = angle2 - angle1;
    let times = Math.floor((difference - (-180)) / 360);

    return (difference - (times * 360)) * -1;
}

function updatePlayerPosition(gamePlayer, dataPlayer) {
    if (gamePlayer.movementTween !== undefined) {
        gamePlayer.movementTween.stop();
    }

    let dataPlayerX = Number(dataPlayer.x);
    let dataPlayerY = Number(dataPlayer.y);
    let player = gamePlayer.player;
    let deltaX = dataPlayerX - player.x;
    let deltaY = dataPlayerY - player.y;
    let distance = Math.sqrt(Math.pow(deltaX, 2) + Math.pow(deltaY, 2));
    if (distance <= 4) {
        player.x = dataPlayerX;
        player.y = dataPlayerY;
    } else {
        gamePlayer.movementTween = game.add.tween(player).to({x: dataPlayerX, y: dataPlayerY}, 5 * distance , Phaser.Easing.Linear.None);
        gamePlayer.movementTween.start()
    }
}