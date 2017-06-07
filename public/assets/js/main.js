const width = window.innerWidth;
const height = window.innerHeight;
const mapSize = 2000;

let game = new Phaser.Game(width, height, Phaser.CANVAS, 'area', { preload: preload, create: create, update: update, render: render });
let socket;
let players = {};
let map;
let live;
let keyboard;
let explosion;

let soldier;

function preload() {
    game.load.audio('explosion', '/assets/audio/explosion.mp3');
    game.load.image('unit', '/assets/images/unit.png');
    game.load.image('bullet', '/assets/images/bullet.png');
    game.load.image('killer', '/assets/images/killers.png');
    game.load.image('earth', '/assets/images/scorched_earth.png');

    game.load.atlasJSONHash('survivor_feet_walk', '/assets/sprites/survivor_feet_walk.png', '/assets/sprites/survivor_feet_walk.json');
    game.load.atlasJSONHash('survivor_move', '/assets/sprites/survivor_move.png', '/assets/sprites/survivor_move.json');
}

function create() {
    socket = io.connect(window.location.host, {path: "/ws/", transports: ['websocket']});

    game.physics.startSystem(Phaser.Physics.ARCADE);
    game.time.advancedTiming = true;
    game.time.desiredFps = 60;
    game.time.slowMotion = 0;

    game.add.tileSprite(0, 0, mapSize, mapSize, 'earth');
    game.world.setBounds(0, 0, mapSize, mapSize);
    game.stage.backgroundColor = "#242424";

    // клавиатура
    keyboard = game.input.keyboard.createCursorKeys();

    //звуки
    explosion = game.add.audio('explosion');

    //получаем имя игрока
    let playerName = prompt("Please enter your name", "guest");
    socket.emit("join_new_player", playerName);

    //создаем игроков
    socket.on("add_players", function(data) {
        data = JSON.parse(data);
        for (let playerId in data) {
            if (!(playerId in players) && data[playerId].isAlive) {
                addPlayer(playerId, data[playerId].x, data[playerId].y, data[playerId].name);
            }
        }

        game.camera.follow(players[socket.id].player);
        live = true;
    });

    //вращение вокруг по событию от сервера
    socket.on("player_rotation_update", function(data) {
        data = JSON.parse(data);
        players[data.id].player.rotation = data.rotation;
    });

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

    //смерть от выстрелов
    socket.on('clean_dead_player', function(victimId) {
        if (victimId === socket.id) {
            live = false;
            let text = game.add.text(width / 2, height / 2, "You lose!", {font: "32px Arial", fill: "#ffffff", align: "center"});
            text.fixedToCamera = true;
            text.anchor.setTo(.5, .5);
        }

        if (victimId in players) {
            players[victimId].player.kill();
            players[victimId].text.destroy();
            delete players[victimId];
        }
    });

    //убираем отключившихся игроков
    socket.on('player_disconnect', function(id) {
        if (id in players) {
            players[id].player.kill();
            players[id].text.destroy();
            delete players[id];
        }
    });

    socket.on('world_update', function(data) {
        for (let playerId in data) {
            if (playerId in players) {
                players[playerId].player.x = data[playerId].x;
                players[playerId].player.y = data[playerId].y;
            }
        }
    });

}

function update() {
    if (live === true) {
        players[socket.id].player.rotation = game.physics.arcade.angleToPointer(players[socket.id].player);
        socket.emit("player_rotation", String(players[socket.id].player.rotation));
        setCollisions();
        characterController();
    }

    for (let id in players) {
        players[id].text.x = Math.floor(players[id].player.x);
        players[id].text.y = Math.floor(players[id].player.y - 25);
    }
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
    if (game.input.keyboard.isDown(Phaser.Keyboard.A) || keyboard.left.isDown) {
        socket.emit("player_move", "A");
    }
    if (game.input.keyboard.isDown(Phaser.Keyboard.D) || keyboard.right.isDown) {
        socket.emit("player_move", "D");
    }
    if (game.input.keyboard.isDown(Phaser.Keyboard.W) || keyboard.up.isDown) {
        socket.emit("player_move", "W");
    }
    if (game.input.keyboard.isDown(Phaser.Keyboard.S) || keyboard.down.isDown) {
        socket.emit("player_move", "S");
    }
}

function render() {
    game.debug.cameraInfo(game.camera, 32, 32);
}

function addPlayer(playerId, x, y, name) {
    let text = game.add.text(0, 0, name, {font: '14px Arial', fill: '#ffffff'});
    let weapon = game.add.weapon(30, 'bullet');
    let player = game.add.sprite(x, y, 'survivor_feet_walk');
    player.anchor.setTo(0.5, 0.5);
    player.scale.setTo(0.25, 0.25);

    player.animations.add('walk');
    player.animations.play('walk', 15, true);

    let body = game.make.sprite(0, 0, 'survivor_move');
    body.anchor.setTo(0.5, 0.5);
    body.animations.add('walk');
    body.animations.play('walk', 15, true);

    player.addChild(body);

    game.physics.arcade.enable(player);
    player.smoothed = false;
    player.body.collideWorldBounds = true;
    player.id = playerId;

    text.anchor.set(0.5);

    weapon.bulletKillType = Phaser.Weapon.KILL_WORLD_BOUNDS;
    weapon.bulletSpeed = 600;
    weapon.fireRate = 100;
    weapon.trackSprite(player, 25, 14, true);

    players[playerId] = { player, weapon, text };
}
