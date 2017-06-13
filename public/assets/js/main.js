const WIDTH = 1280;
const HEIGHT = 740;
const MAP_LOW_BOUND = 50;
const MAP_HIGH_BOUND = 1950;

let game = new Phaser.Game(WIDTH, HEIGHT, Phaser.CANVAS, 'area', { preload: preload, create: create, update: update, render: render });
let socket;
let players = {};
let map;
let live;
let keyboard;
let explosion;

function preload() {
    game.load.audio('explosion', '/assets/audio/explosion.mp3');
    game.load.image('bullet', '/assets/images/bullet.png');
    game.load.image('tiles', '/assets/sprites/tilesetHouse.png');

    game.load.tilemap('map', '/assets/sprites/map.json', null, Phaser.Tilemap.TILED_JSON);
    game.load.atlas('soldier', '/assets/sprites/soldier.png', '/assets/sprites/soldier.json');
    game.load.atlas('gamepad', '/assets/sprites/gamepad.png', '/assets/sprites/gamepad.json');
}

function create() {
    socket = io.connect(window.location.host, {path: "/ws/", transports: ['websocket']});

    game.physics.startSystem(Phaser.Physics.ARCADE);
    game.time.advancedTiming = true;
    game.time.desiredFps = 60;
    game.time.slowMotion = 0;

    // клавиатура
    keyboard = game.input.keyboard.createCursorKeys();

    //звуки
    explosion = game.add.audio('explosion');

    //карта
    initTileMap();

    // геймпад
    initVirtualGamepad();

    //получаем имя игрока
    let savedName = window.localStorage.getItem("player_name");
    if (!savedName) savedName = "guest";

    let playerName = prompt("Please enter your name", savedName);
    if (!playerName) playerName = "";
    window.localStorage.setItem("player_name", playerName);

    socket.emit("join_new_player", playerName);

    //вызываем выстрелы
    game.input.onDown.add(function() {
        socket.emit("shots_fired", socket.id);
    });

    //ввзываем выстрелы
    socket.on('player_fire_add', function(id) {
        if (live && id in players) {
            explosion.play();
            players[id].weapon.fire();
        }
    });

    socket.on('world_update', function(data) {
        data = JSON.parse(data);
        let dataPlayers = data.players;
        for (let playerId in dataPlayers) {
            if (playerId in players) {
                players[playerId].player.visible = dataPlayers[playerId].isAlive;
                players[playerId].text.visible = dataPlayers[playerId].isAlive;
                if (players[playerId].debugText !== undefined) {
                    players[playerId].debugText.visible = dataPlayers[playerId].isAlive;
                }

                if (dataPlayers[playerId].isAlive) {
                    if (playerId === socket.id) {
                        let lastCommand = undefined;
                        if (data.commands !== undefined && data.commands !== null && data.commands.length > 0) {
                            lastCommand = data.commands[data.commands.length-1];
                        }

                        checkPredictions(players[playerId], dataPlayers[playerId], lastCommand);
                    } else {
                        updatePlayerRotation(players[playerId], dataPlayers[playerId]);
                        updatePlayerPosition(players[playerId], dataPlayers[playerId]);
                    }

                } else {
                    if (playerId === socket.id && live) {
                        live = false;
                        let text = game.add.text(WIDTH / 2, HEIGHT / 2, "You lose!", {font: "32px Arial", fill: "#ffffff", align: "center"});
                        text.fixedToCamera = true;
                        text.anchor.setTo(.5, .5);
                    }
                }
            } else {
                if (dataPlayers[playerId].isAlive) {
                    addPlayer(dataPlayers[playerId]);

                    if (playerId === socket.id) {
                        game.camera.follow(players[socket.id].player);
                        live = true;
                    }
                }
            }
        }

        for (let playerId in players) {
            if (!(playerId in dataPlayers)) {
                updateKilledPlayer(playerId)
            }
        }
    });
}

function updateKilledPlayer(playerId) {
    players[playerId].player.kill();
    players[playerId].text.destroy();
    delete players[playerId];
}

function update() {
    if (live === true) {
        let gamePlayer = players[socket.id];
        let player = gamePlayer.player;
        let newRotation = fixRotation(game.physics.arcade.angleToPointer(player));
        if (fixRotation(player.rotation) !== newRotation) {
            player.rotation = newRotation;
            notifyPlayerRotated(gamePlayer);
        }

        setCollisions();
        characterController();

        //for debug mode
        let debugText = gamePlayer.debugText;
        if (debugText !== undefined) {
            debugText.setText("Commands in History = " + gamePlayer.executedCommands.length);
            debugText.x = Math.floor(player.x);
            debugText.y = Math.floor(player.y - 55);
        }
    }

    for (let id in players) {
        players[id].text.x = Math.floor(players[id].player.x);
        players[id].text.y = Math.floor(players[id].player.y - 35);
    }
}

function fixRotation(rotation) {
    return Math.round(rotation * 10000) / 10000
}

function bulletHitHandler(player, bullet) {
    socket.emit("player_killed", player.id);

    bullet.destroy();
}

function setCollisions() {
    for (let x in players) {
        for (let y in players) {
            if (x !== y) {
                game.physics.arcade.collide(players[x].weapon.bullets, players[y].player, bulletHitHandler, null, this);
            }
        }
    }
}

function characterController() {
    let gamePlayer = players[socket.id];
    let player = gamePlayer.player;
    if (game.input.keyboard.isDown(Phaser.Keyboard.A) || keyboard.left.isDown) {
        changePlayerPosition(player, "x", -2);
        notifyPlayerMoved(gamePlayer, "A");
    }
    if (game.input.keyboard.isDown(Phaser.Keyboard.D) || keyboard.right.isDown) {
        changePlayerPosition(player, "x", 2);
        notifyPlayerMoved(gamePlayer, "D");
    }
    if (game.input.keyboard.isDown(Phaser.Keyboard.W) || keyboard.up.isDown) {
        changePlayerPosition(player, "y", -2);
        notifyPlayerMoved(gamePlayer, "W");
    }
    if (game.input.keyboard.isDown(Phaser.Keyboard.S) || keyboard.down.isDown) {
        changePlayerPosition(player, "y", 2);
        notifyPlayerMoved(gamePlayer, "S");
    }

    checkBounds(player);
    checkMove(player);
}

function changePlayerPosition(player, field, delta) {
    player[field] += delta;
    checkBounds(player);
}

function checkBounds(player) {
    if (player.x < MAP_LOW_BOUND) player.x = MAP_LOW_BOUND;
    if (player.y < MAP_LOW_BOUND) player.y = MAP_LOW_BOUND;
    if (player.x > MAP_HIGH_BOUND) player.y = MAP_HIGH_BOUND;
    if (player.y > MAP_HIGH_BOUND) player.y = MAP_HIGH_BOUND;
}

function checkMove(player) {
    if(Math.abs(player.body.velocity.x) > 0 || Math.abs(player.body.velocity.y) > 0) {
        player.play('move');
    } else {
        player.play('idle');
    }
}

function render() {
    game.debug.cameraInfo(game.camera, 32, 32);
}

function initVirtualGamepad() {
    let gamepad = game.plugins.add(Phaser.Plugin.VirtualGamepad)
    this.joystick = gamepad.addJoystick(90, game.height - 90, 0.75, 'gamepad');
    gamepad.addButton(game.width - 90, game.height - 90, 0.75, 'gamepad');
}

function initTileMap() {
    let map = game.add.tilemap('map');
    this.map = map;

    map.addTilesetImage('tilesetHouse', 'tiles');
    map.createLayer('Base');

    let collisionLayer = map.createLayer('Collision');
    this.collisionLayer = collisionLayer;

    collisionLayer.visible = false;

    map.setCollisionByExclusion([], true, this.collisionLayer);

    collisionLayer.resizeWorld();

    map.createLayer('Foreground');
}

function addPlayer(playerObj) {
    let text = game.add.text(0, 0, playerObj.name, {font: '14px Arial', fill: '#ffffff'});
    let weapon = game.add.weapon(30, 'bullet');
    let player = game.add.sprite(playerObj.x, playerObj.y, 'soldier');
    player.anchor.setTo(0.5, 0.5);
    player.scale.setTo(0.25, 0.25);
    player.animations.add('idle', [0 ,1 ,2 ,3 ,4 ,5 ,6 ,7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19], 30, true);
    player.animations.add('move', [20 ,21 ,22 ,23 ,24 ,25 ,26 ,27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38], 40, true);
    player.play('move');

    player.rotation = playerObj.rotation;

    // let body = game.make.sprite(0, 0, 'survivor_move');
    // body.anchor.setTo(0.5, 0.5);
    // body.animations.add('walk');
    // body.animations.play('walk', 15, true);
    //
    // player.addChild(body);

    game.physics.arcade.enable(player);
    player.smoothed = false;
    player.body.collideWorldBounds = true;
    player.id = playerObj.id;

    text.anchor.set(0.5);

    weapon.bulletKillType = Phaser.Weapon.KILL_WORLD_BOUNDS;
    weapon.bulletSpeed = 600;
    weapon.fireRate = 100;
    weapon.trackSprite(player, 25, 14, true);

    players[playerObj.id] = { player, weapon, text };

    //temporary added for debug purposes
    if (playerObj.id === socket.id) {
        let debugText = game.add.text(0, 0, playerObj.name, {font: '14px Arial', fill: "#ffffff"});
        debugText.anchor.set(0.5);
        players[playerObj.id].debugText = debugText
    }
}
