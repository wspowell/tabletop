//////////////////////////////////////////////////////////////////////////////////////////
//////// Login
//////////////////////////////////////////////////////////////////////////////////////////

function User(username, displayName, isStoryTeller, externalCharacterSheetLink) {
    this.username = username
    this.displayName = displayName
     this.isStoryTeller = isStoryTeller
     this.externalCharacterSheetLink = externalCharacterSheetLink
}

let loggedInPlayer = null;
let keepAliveJob = null;

function showLogin() {
    username = "";

    const login = document.getElementById("login");
    const pageLayout = document.getElementById("page-layout");

    // Swap visibility of login and map.
    login.style.display = "block";
    pageLayout.style.display = "none";
}

function setLoginError(errorMessage) {
    username = "";

    const loginError = document.getElementById("loginError");
    loginError.innerText = errorMessage;
}

function login() {
    const username = document.getElementById("loginUsername");
    const secret = document.getElementById("loginSecret");

    sendMessage({
        type: "login",
        data: {
            username: username.value,
            secret: secret.value
        }
    });
}

function completeLogin(user) {
    loggedInPlayer = new User(
        user.username, 
        user.displayName, 
        user.isStoryTeller, 
        user.externalCharacterSheetLink
    );

    const login = document.getElementById("login");
    const pageLayout = document.getElementById("page-layout");

    // Swap visibility of login and map.
    login.style.display = "none";
    pageLayout.style.display = "grid";

    if (!loggedInPlayer.isStoryTeller) {
        const stContainer = document.getElementById("st-container");
        stContainer.style.display = "none";
        pageLayout.style.gridTemplateRows = "1fr 5fr";
    }

    keepAlive();
    keepAliveJob = setInterval(keepAlive, 10000);

    console.log("player successfully logged in", loggedInPlayer.username);
}

function keepAlive() {
    if (!loggedInPlayer) {
        clearInterval(keepAliveJob);
        return;
    }

    sendMessage({
        type: "keepAlive",
        data: {}
    });

    return;
}


//////////////////////////////////////////////////////////////////////////////////////////
//////// Tokens
//////////////////////////////////////////////////////////////////////////////////////////

function sendTokensHome(shouldSendUpdate) {
    document.querySelectorAll('.token').forEach(function(token) {
        const mapName = document.getElementById("map-grid").dataset.mapName;
        const tokenId = token.id;

        if (token.dataset.npc == "true") {
            if (token.dataset.instance == "true") {
                if (shouldSendUpdate) {
                    deleteToken(tokenId, mapName);
                    sendMessage({
                        type: "deleteToken",
                        data: {
                            id: npcToken.id,
                            mapName: mapName,
                        }
                    });
                } else {
                    token.remove();
                }
            }
        } else {
            const tokenHome = document.getElementById(token.dataset.homeId);
            tokenHome.appendChild(token);

            if (shouldSendUpdate) {
                updateTokenStateCache(tokenId, token.dataset.tokenName, mapName, {
                    x: 0,
                    y: 0
                }, true);

                sendMessage({
                    type: "tokenPosition",
                    data: {
                        id: tokenId,
                        tokenName: token.dataset.tokenName,
                        mapName: mapName,
                        position: {
                            x: 0,
                            y: 0,
                        },
                        isHome: true,
                    }
                });
            }
        }
    });
}

function loadPlayer(user) {
    console.log("loading player", user.username);

    const username = user.username;
    const displayName = user.displayName;
    const externalCharacterSheetLink = user.externalCharacterSheetLink;


    const header = document.getElementById("header");

    const player = document.createElement("div");
    if (user.isStoryTeller) {
        player.style.borderRight = "4px double #42404d";
        player.style.paddingRight = "20px";
    }
    player.id = "player_"+username;
    player.classList.add("player");

    const playerName = document.createElement("div");
    playerName.id = "player-name-"+username;
    playerName.classList.add("player-name");
    playerName.innerText = displayName;
    if (externalCharacterSheetLink != "") {
        console.log("linking", username, "to external character sheet at", externalCharacterSheetLink)
        playerName.style.cursor = "pointer"
        playerName.onclick = function() {
            window.open(externalCharacterSheetLink, '_blank');
            return false;
        }
    }
    player.appendChild(playerName);

    // ST does not have a player token.
    if (!user.isStoryTeller) {
        // Add player token and token home.

        const playerTokenHome = document.createElement("div");
        playerTokenHome.id = "player_token_home_"+username;
        playerTokenHome.classList.add("token-home");
        playerTokenHome.style.width = gridCellSizePx+"px";
        playerTokenHome.style.height = gridCellSizePx+"px";
        // Apply drag and drop functionality.
        playerTokenHome.ondragover = function(event) {
            event.preventDefault();
        };
        playerTokenHome.ondrop = function(event) {
            event.preventDefault();
            const droppingUsername = event.dataTransfer.getData("username");
            if (username != droppingUsername) {
                return;
            }

            const tokenId = event.dataTransfer.getData("tokenId");
            const token = document.getElementById(tokenId);


            const aPosition = getPositionAtCenter(token);
            const bPosition = getPositionAtCenter(event.target);
            token.style.top = (aPosition.y - bPosition.y)+"px";
            token.style.left = (aPosition.x - bPosition.x)+"px";

            event.target.appendChild(token);

            setTimeout(function() {
                token.style.top = "0px";
                token.style.left = "0px";
            }, 1);

            const mapName = document.getElementById("map-grid").dataset.mapName;

            // Update the local cache so we can put them back when we move between maps.
            updateTokenStateCache(tokenId, token.dataset.tokenName, mapName, {
                x: 0,
                y: 0
            }, true);

            sendMessage({
                type: "tokenPosition",
                data: {
                    id: tokenId,
                    tokenName: token.dataset.tokenName,
                    mapName: mapName,
                    position: {
                        x: 0,
                        y: 0,
                    },
                    isHome: true,
                }
            });
        };
        player.appendChild(playerTokenHome);

        const playerToken = document.createElement("div");
        playerToken.id = "token_"+username;
        playerToken.classList.add("token");
        playerToken.style.backgroundImage = "url('/data/users/"+username+"/token')";
        playerToken.style.backgroundSize = gridCellSizePx+"px "+gridCellSizePx+"px";
        playerToken.style.width = gridCellSizePx+"px";
        playerToken.style.height = gridCellSizePx+"px";
        playerToken.dataset.homeId = playerTokenHome.id;
        playerToken.dataset.tokenName = username;
        playerToken.oncontextmenu = function(event) {
            event.preventDefault();

            // Send home.
            playerTokenHome.appendChild(document.getElementById(event.target.id));
        };
        // Apply drag and drop functionality.
        playerToken.draggable = true;
        playerToken.ondragstart = function(event) {
            // Only allow token owners or ST to move tokens.
            if (!loggedInPlayer.isStoryTeller && loggedInPlayer.username != username) {
                return false;
            }

            event.dataTransfer.setData("tokenId", event.target.id);
            event.dataTransfer.setData("username", username);
        };
        playerToken.ondragover = function(event) {
            event.preventDefault();
            return false;
        };
        playerToken.ondrop = function(event) {
            event.preventDefault();
            return false;
        }
        playerTokenHome.appendChild(playerToken);

        header.appendChild(player);
        addHealthBar(username, player);
    } else {
        const stToken = document.createElement("div");
        stToken.id = "st-token";
        stToken.style.backgroundImage = "url('/data/users/"+username+"/token')";
        stToken.style.backgroundSize = gridCellSizePx+"px "+gridCellSizePx+"px";
        stToken.style.width = gridCellSizePx+"px";
        stToken.style.height = gridCellSizePx+"px";

        player.appendChild(stToken);

        // Always put the ST first in the players list.
        header.insertBefore(player, header.firstChild);
    }
}

function addHealthBar(username, token) {
    const healthBar = document.createElement("div");
    healthBar.classList.add("health-bar");
    
    const healthValue = document.createElement("div");
    healthValue.classList.add("health-value");
    const healthValueCurrent = document.createElement("input");
    healthValueCurrent.id = "current-health-"+username;
    healthValueCurrent.type = "text";
    healthValueCurrent.classList.add("health-input");
    healthValueCurrent.value = 0;
    healthValueCurrent.dataset.username = username;
    healthValueCurrent.onfocus = modifyHealthValue;
    const healthDivider = document.createElement("div");
    healthDivider.textContent = "/";
    const healthValueMax = document.createElement("input");
    healthValueMax.id = "max-health-"+username;
    healthValueMax.type = "text";
    healthValueMax.classList.add("health-input");
    healthValueMax.value = 0;
    healthValueMax.dataset.username = username;
    healthValueMax.onfocus = modifyHealthValue;
    healthValue.appendChild(healthValueCurrent);
    healthValue.appendChild(healthDivider);
    healthValue.appendChild(healthValueMax);

    healthBar.appendChild(healthValue);

    token.appendChild(healthBar);

    window.addEventListener("updateStateHealth", function(event) {
        if (event.detail.playerHealth && event.detail.playerHealth[username]) {
            const health = event.detail.playerHealth[username];
            healthValueCurrent.value = health.currentHealth;
            healthValueMax.value = health.maxHealth;
        }
        applyHealthEffect(username);
    });
}

function emitHealthStateUpdate(playerHealth) {
    const event = new CustomEvent("updateStateHealth", {
        detail: {
            playerHealth: playerHealth,
        }
    });
    window.dispatchEvent(event);
}

function modifyHealthValue(event) {
    const originalValue = event.target.value;
    event.target.value = "";

    event.target.onkeydown = function (e) {
        e.target.style.width = e.target.value.length+2 + "ch"; // Make the input as wide as the text.
        if (e.code === "Enter") { // Checks whether the pressed key is "Enter".
            event.target.blur();
        }
    };
    event.target.onblur = function (e) {
        const username = event.target.dataset.username;

        if (event.target.value.startsWith("+") || event.target.value.startsWith("-")) {
            const newValue = Number(originalValue) + Number(event.target.value);
            console.log("health altered", originalValue, event.target.value, "=", newValue);
            event.target.value = newValue;

            const playerHealth = {};
            playerHealth[username] = {
                currentHealth: Number(document.getElementById("current-health-"+username).value),
                maxHealth: Number(document.getElementById("max-health-"+username).value),
            };
            sendMessage({
                type: "playerHealth",
                data: {
                    playerHealth: playerHealth
                }
            });
        } else if (event.target.value === "") {
            event.target.value = Number(originalValue);
        } else {
            console.log("health overridden", originalValue, "->", event.target.value);
            event.target.value = Number(event.target.value);

            const playerHealth = {};
            playerHealth[username] = {
                currentHealth: Number(document.getElementById("current-health-"+username).value),
                maxHealth: Number(document.getElementById("max-health-"+username).value),
            };
            sendMessage({
                type: "playerHealth",
                data: {
                    playerHealth: playerHealth
                }
            });
        }

        applyHealthEffect(username);
    };
}

function applyHealthEffect(username) {
    const currentHealth = document.getElementById("current-health-"+username);
    const maxHealth = document.getElementById("max-health-"+username);
    const playerName = document.getElementById("player-name-"+username);

    currentHealth.classList.remove("health-bloodied");
    currentHealth.classList.remove("health-beefy");
    playerName.classList.remove("health-unconscious");

    if (Number(currentHealth.value) < (Number(maxHealth.value)/2.0)) {
        currentHealth.classList.add("health-bloodied");

        if (Number(currentHealth.value) <= 0) {
            playerName.classList.add("health-unconscious");
        }
    } else if (Number(currentHealth.value) > Number(maxHealth.value)) {
        currentHealth.classList.add("health-beefy");
    }

    // Make the input as wide as the text.
    currentHealth.style.width = currentHealth.value.length + "ch"; 
    maxHealth.style.width = maxHealth.value.length + "ch";
}


function unloadPlayer(username) {
    console.log("unloading player", username)

    const player = document.getElementById("player_"+username);
    const playerToken = document.getElementById("token_"+username);
    
    // Might not have one if the player is the ST.
    if (playerToken) {
        playerToken.remove();
    }
    player.remove();
}

function setTokenPosition(id, tokenName, mapName, position, isHome) {
    if (isHome) {
        console.log("token", id, "moved to home on map", mapName)
    } else {
        console.log("token", id, "moved to", position, "on map", mapName)
    }

    const isNpc = id.startsWith("npc_");

    let token = null;
    if (isNpc) {
        token = document.getElementById(id);
        if (!token) {
            // Token does not already exists so we need to create one from the token bank.
            token = createNpcToken(tokenName, true);
            token.id = id;
        }
    } else {
        token = document.getElementById(id);
        if (!token) { return } // Token may not be loaded into the page (ex setting from token cache for offline user).
    }

    updateTokenStateCache(id, token.dataset.tokenName, mapName, position, isHome);


    
    let target;
    if (isHome) {
        // NPCs are always listed in their token home.
        if (!isNpc) {
            const tokenHome = document.getElementById(token.dataset.homeId);
            // tokenHome.appendChild(token);
            target = tokenHome;
        }
    } else {
        const gridCell = document.getElementById(position.x+","+position.y);
        target = gridCell;
    }


    const aPosition = getPositionAtCenter(token);
    const bPosition = getPositionAtCenter(target);
    token.style.top = (aPosition.y - bPosition.y)+"px";
    token.style.left = (aPosition.x - bPosition.x)+"px";

    target.appendChild(token);

    setTimeout(function() {
        token.style.top = "0px";
        token.style.left = "0px";
    }, 1);
}

let tokenStateCache = {};

function updateTokenStateCache(id, tokenName, mapName, position, isHome) {
    // Update the local cache so we can put them back when we move between maps.

    if (!tokenStateCache) {
        tokenStateCache = {};
    }

    if (!tokenStateCache[mapName]) {
        tokenStateCache[mapName] = {};
    }

    if (!tokenStateCache[mapName][id]) {
        tokenStateCache[mapName][id] = {};
    }

    tokenStateCache[mapName][id].tokenName = tokenName;
    tokenStateCache[mapName][id].position = position;
    tokenStateCache[mapName][id].isHome = isHome;
}

function setTokenBank(tokenNames) {
    const tokenBank = document.getElementById("token-bank");

    for (let i = 0; i < tokenNames.length; i++) {
        const tokenName = tokenNames[i];

        console.log("adding npc token", tokenName);

        const npcTokenHome = document.createElement("div");
        npcTokenHome.id = "npc_token_home_"+tokenName;
        npcTokenHome.classList.add("token-home");
        npcTokenHome.style.width = gridCellSizePx+"px";
        npcTokenHome.style.height = gridCellSizePx+"px";
        
        const npcToken = createNpcToken(tokenName, false);
        npcTokenHome.appendChild(npcToken);

        tokenBank.appendChild(npcTokenHome);
    }
}

let nextNpcTokenId = 1;
function createNpcToken(tokenName, isInstance) {
    console.log("creating npc token", tokenName, "- is instance", isInstance);
    
    const npcToken = document.createElement("img");
    npcToken.id = "npc_"+tokenName;
    if (isInstance) {
        npcToken.id += "_"+nextNpcTokenId;
        nextNpcTokenId++;
    }
    npcToken.classList.add("token");
    npcToken.src = "/data/npcs/"+tokenName;
    npcToken.style.width = gridCellSizePx+"px";
    npcToken.style.height = gridCellSizePx+"px";
    npcToken.dataset.npc = "true";
    npcToken.dataset.instance = ""+isInstance;
    npcToken.dataset.tokenName = tokenName;
    npcToken.oncontextmenu = function(event) {
        event.preventDefault();

        // Only allow token owners or ST to move tokens.
        if (!loggedInPlayer.isStoryTeller) {
            return false;
        }

        // Only allow deletion of instanced tokens.
        if (npcToken.dataset.instance == "false") {
            return false;
        }

        // Send home.
        // There can be multiple so just delete them since the token bank always has a copy.
        const mapName = document.getElementById("map-grid").dataset.mapName;
        deleteToken(event.target.id, mapName);

        sendMessage({
            type: "deleteToken",
            data: {
                id: npcToken.id,
                mapName: mapName,
            }
        });
    };
    // Apply drag and drop functionality.
    npcToken.draggable = true;
    npcToken.ondragstart = function(event) {
        // Only allow token owners or ST to move tokens.
        if (!loggedInPlayer.isStoryTeller) {
            return false;
        }

        event.dataTransfer.setData("tokenId", event.target.id);
        event.dataTransfer.setData("npc", "true");
    };
    npcToken.ondragover = function(event) {
        event.preventDefault();
        return false;
    };
    npcToken.ondrop = function(event) {
        event.preventDefault();
        return false;
    }

    return npcToken;
}

function deleteToken(tokenId, mapName) {
    document.getElementById(tokenId).remove();

    if (!tokenStateCache) {
        tokenStateCache = {};
    }

    if (!tokenStateCache[mapName]) {
        tokenStateCache[mapName] = {};
    }

    if (!tokenStateCache[mapName][tokenId]) {
        tokenStateCache[mapName][tokenId] = {};
    }

    delete tokenStateCache[mapName][tokenId];
}

//////////////////////////////////////////////////////////////////////////////////////////
//////// Map Grid
//////////////////////////////////////////////////////////////////////////////////////////

const gridCellSizePx = 50;

function setMapList(mapNames) {
    console.log("map list", mapNames);

    const mapList = document.getElementById("map-list");

    for (let i = 0; i < mapNames.length; i++) {
        const mapName = document.createElement("li");
        mapName.innerText = mapNames[i];
        mapName.onclick = function () {
            sendMessage({
                type: "mapChange",
                data: {
                    mapName: mapNames[i]
                }
            })
        };
        mapList.appendChild(mapName);
    }
}

function changeMap(mapName, mapData) {
    console.log("changing map", mapName, mapData);
    sendTokensHome(false);
    updateMap(mapName, mapData);
}

// Setup the map-grid to update the image and grid anytime the map image is changed.
window.addEventListener("load", function() {
    // updateMap('tavern');
});

function updateMap(mapName, mapData) {
    const imageUrl = "/data/maps/"+mapName;

    const mapGrid = document.getElementById('map-grid');
    mapGrid.style.backgroundImage = "url('"+imageUrl+"')";
    mapGrid.dataset.mapName = mapName;

    console.log("map-grid image source", mapGrid.style.backgroundImage );
    var imageSrc = mapGrid.style.backgroundImage.replace(/url\((['"])?(.*?)\1\)/gi, '$2').split(',')[0];
    console.log("map-grid image source", imageSrc);

    var mapImage = new Image();
    mapImage.src = imageSrc;
    mapImage.onload = function() {
        let width = mapImage.width;
        let height = mapImage.height;

        if (mapData && mapData.width) {
            width = mapData.width;
        }

        if (mapData && mapData.height) {
            height = mapData.height;
        }

        // Send tokens home briefly so we do not lose them when the map regenerates.
        sendTokensHome(false);

        generateMapGrid(width, height);

        // Set token positions based on saved state.
        const mapTokens = tokenStateCache[mapName];
        if (mapTokens) {
            for (tokenId in mapTokens) {
                const tokenName = mapTokens[tokenId].tokenName;
                const position = mapTokens[tokenId].position;
                const isHome = mapTokens[tokenId].isHome;
                setTokenPosition(tokenId, tokenName, mapName, position, isHome);
            }
        }
    };
}

function regenerateMapGrid(mapWidth, mapHeight) {
    // changeMap(mapName, {
    //     width: mapWidth,
    //     height: mapHeight,
    // })
    sendMessage({
        type: "mapSize",
        data: {
            width: Number(mapWidth),
            height: Number(mapHeight),
        }
    });
}

function generateMapGrid(mapWidth, mapHeight) {
    console.log("generating map grid...");

    const mapGrid = document.getElementById('map-grid');

    // Delete any existing data.
    mapGrid.innerHTML = "";

    console.log("map size", mapWidth, "x", mapHeight);

    document.getElementById("new-grid-width").value = mapWidth;
    document.getElementById("new-grid-height").value = mapHeight;

    const numColumns = Math.trunc(mapWidth / gridCellSizePx);
    const numRows = Math.trunc(mapHeight / gridCellSizePx);

    console.log("grid size", numColumns, "x", numRows);

    mapGrid.style.width = (numColumns*gridCellSizePx)+"px";
    mapGrid.style.height = (numRows*gridCellSizePx)+"px";
    mapGrid.style.gridTemplateColumns = "repeat("+numColumns+", "+gridCellSizePx+"px)";
    mapGrid.style.backgroundSize = mapWidth+"px "+mapHeight+"px";

    for (y = 0; y < numRows; y++) {
        for (x = 0; x < numColumns; x++) {
            const gridCell = document.createElement("div");
            gridCell.id = x+","+y;
            gridCell.style.width = gridCellSizePx+"px";
            gridCell.style.height = gridCellSizePx+"px";
            gridCell.dataset.x = x;
            gridCell.dataset.y = y;
            gridCell.oncontextmenu = function() {
                return false;
            };
            // Apply Drag and drop functionality.
            gridCell.ondragover = function(event) {
                event.preventDefault();
            };
            gridCell.ondrop = function(event) {
                event.preventDefault();

                // Prevent dropping multiple tokens in the same cell.
                if (gridCell.children.length > 0) {
                    return false;
                }

                let tokenId = event.dataTransfer.getData("tokenId");
                let token = document.getElementById(tokenId);
                if (token.dataset.npc == "true") {
                    // If the token is dragging from the token bank then clone the element.
                    if (token.dataset.instance == "false") {
                        const tokenName = token.dataset.tokenName;
                        const tokenClone = createNpcToken(tokenName, true);
                        tokenId = tokenClone.id;
                        token = tokenClone;
                    } 
                }

                // Move the token into the grid cell.

                const aPosition = getPositionAtCenter(token);
                const bPosition = getPositionAtCenter(event.target);
                token.style.top = (aPosition.y - bPosition.y)+"px";
                token.style.left = (aPosition.x - bPosition.x)+"px";

                event.target.appendChild(token);

                setTimeout(function() {
                    token.style.top = "0px";
                    token.style.left = "0px";
                }, 1);

                const mapName = document.getElementById("map-grid").dataset.mapName;

                // Update the local cache so we can put them back when we move between maps.
                updateTokenStateCache(tokenId, token.dataset.tokenName, mapName, {
                    x: Number(gridCell.dataset.x),
                    y: Number(gridCell.dataset.y)
                }, false);
                
                sendMessage({
                    type: "tokenPosition",
                    data: {
                        id: tokenId,
                        tokenName: token.dataset.tokenName,
                        mapName: mapName,
                        position: {
                            x: Number(gridCell.dataset.x),
                            y: Number(gridCell.dataset.y),
                        },
                        isHome: false,
                    }
                });
            };
            mapGrid.appendChild(gridCell);
        }
    }
}


function getPositionAtCenter(element) {
    const {top, left, width, height} = element.getBoundingClientRect();
    return {
        x: left + width / 2,
        y: top + height / 2
    };
}

//////////////////////////////////////////////////////////////////////////////////////////
//////// WebSocket
//////////////////////////////////////////////////////////////////////////////////////////

const websocketUrl = "ws://"+window.location.hostname+":"+websocketPort+"/websocket";
// Create WebSocket connection.
let socket;
function resetSocket() {
    console.log("reset websocket");

    socket = new WebSocket(websocketUrl);

    // Connection opened
    socket.addEventListener("open", (event) => {
        console.log("connected to", websocketUrl);

        showLogin();
    });

    socket.addEventListener("close", (event) => {
        loggedInPlayer = null;

        showLogin();
        setLoginError("disconnected from server (please refresh to reconnect)");

        // setTimeout(function () {
        //     resetSocket();
        // }, 5000);
    });

    socket.addEventListener("error", (event) => {
        const loginError = document.getElementById("loginError");
        loginError.innerText = "disconnected from server (please refresh to reconnect)";

        loggedInPlayer = null;

        showLogin();
        setLoginError(payload.data.errorMessage);

        // setTimeout(function () {
        //     resetSocket();
        // }, 5000);
    });

    socket.addEventListener("message", (event) => {
        console.log("received message ", event.data);

        const payload = JSON.parse(event.data);

        switch (payload.type) {
            case "error":
                switch (payload.data.typeOfFailedMessage) {
                    case "login":
                        showLogin();
                        setLoginError(payload.data.errorMessage);
                        break;
                    default:
                        console.log("error:", payload.data.type, payload.data.errorMessage);
                        break;
                }
                break;
            case "loginSuccess":
                completeLogin(payload.data.user);
                break;
            case "userOnline":
                loadPlayer(payload.data.user);
                break;
            case "userOffline":
                unloadPlayer(payload.data.username);
                break;
            case "logout":
                console.log("user logged out:", payload.data.username);
                break;
            case "state":
                if (payload.data.mapTokens) {
                    tokenStateCache = payload.data.mapTokens;
                }
            
                if (payload.data.currentMap) {
                    updateMap(payload.data.currentMap, payload.data.mapData);
                }

                emitHealthStateUpdate(payload.data.playerHealth);
                break;
            case "tokenPosition":
                setTokenPosition(payload.data.id, payload.data.tokenName, payload.data.mapName, payload.data.position, payload.data.isHome);
                break;
            case "deleteToken":
                deleteToken(payload.data.id, payload.data.mapName);
                break;
            case "mapSize":
                // generateMapGrid(payload.data.width, payload.data.height);
                const mapName = document.getElementById("map-grid").dataset.mapName;
                changeMap(mapName, {
                    width: payload.data.width,
                    height: payload.data.height,
                });
                break;
            case "mapList":
                setMapList(payload.data.mapNames);
                break;
            case "tokenBank":
                setTokenBank(payload.data.tokenNames);
                break;
            case "mapChange":
                changeMap(payload.data.mapName, payload.data.mapData);
                break;
            case "playerHealth":
                emitHealthStateUpdate(payload.data.playerHealth);
                break;
            default:
                console.error("unknown message type:", payload.type);
                break;
        }

        if (payload.event === "updateTokenPosition") {
            moveTokens(payload.data.tokens);
        }
    });
}

function sendMessage(payload) {
    const message = JSON.stringify(payload);
    if (payload.type !== "keepAlive") {
        console.log("sending message", message);
    }
    socket.send(message);
}

window.addEventListener("load", function() {
    resetSocket();
});


