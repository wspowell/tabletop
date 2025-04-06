# tabletop
A tabletop game server for use with trusted friends.

![](https://github.com/wspowell/tabletop/blob/main/preview.png "Preview of tabletop showing token being moved around a game map in real time")

## Running the server
Run `make build` to produce a binary to run in either linux or windows. The server runs a web server on port 8080 and a websocket server on port 3000. The server is intended to be run by the game's Story Teller (ST) either on their computer or some other server they can access.

## Tokens
Each player has their own token and only they and the ST may manipulate that token. There are NPC tokens that only the ST can see and access. Tokens are placed on a grid that has a fixed cell size. Game maps are created by the ST and fitted to the game grid. Right clicking on a token sends it back to its token home.

## Configuration
All configuration is done by the ST on their own machine that is hosting the server. Everything inside the "data" directory drives the game.

### data/cache
This folder only stores temporary values related to table top state. You may delete this folder at any point, but you may lose data such as last token positions on a map.

### data/maps
These are the maps the ST wants their player characters to explore. Each png filename is the map name (only png is supported). For example, map "tavern" would be "tavern.png". Each map has optional configuration that can be applied using a json file of the same name. For example, map "tavern" may have configuration stored in "tavern.json". The allowed configuration is as follows:
```
{
    "width": "1000", // Size, in pixels, of the width of the map.
    "height": "1000" // Size, in pixels, of the height of the map.
}
```
The size configuration allows you fit a map to the fixed size grid cells. This can be edited in the browser via the ST view and then updated in the map's configuration file to make permanent.

### data/npcs
These are NPC tokens the ST wants their player characters to interact with. Like maps, these must be png files. NPCs can be duplicated as many times as an ST wants on the table top map, so only one image file is required. For example, if you wish for 10 rats, then you only need one "rat.png" and you can then drag and drop as many rats as you want your player characters to face.

NPCs could even be player characters if someone cannot make it to a session. For this, and maybe other game based reasons, it could be a good idea to copy your player character tokens to this directory.

### data/users
These are player users, including the ST, that will be logging into the server. Each username is defined by the directory name. For example, a username of "red" would have a directory named "data/users/red/". Any files stored in this directory defines information for this user. Each user directory contains three files:
* info.json
* secret.txt
* token.png

info.json defines user data needed for rendering for all players. This file looks like the following:
```
{
    "username": "red", // Must match the directory name
    "isStoryTeller": false, // If true, this user is the ST. Only one user should be marked as the ST
    "displayName": "Sir Red Token of the Southern Isles", // The name that everyone sees this player as
    "externalCharacterSheetLink": "https://www.example.com/mycharactersheet" // External character sheet for sharing with everyone.
}
```
The character sheet is integrated into the UI by clicking on a players name. This opens a new browser tab so you may view that players character sheet.

secret.txt stores that user's passphrase for connecting to the server. This is stored as a simple plaintext word and is only meant as a granular check that players are connecting as the correct character tokens. DO NOT store anything actually secret in this file. Assume that all players know the contents of this file.

token.png is the image for the player token. There are several third party sites that allow for generation of character tokens. The ST will not have a player token.

## Security
There is basically none by design. This is only meant to be run with trusted friends. As such, several design decisions were made intentionally because it eased the development. It is possible, for example, for any user to assume the role of any other player via frontend code manipulation. Solving these kinds of problems are way out of scope for a hobby project such as this. If you would not trust someone, in person, to not slip a few extra dollars during Monopoly, then this is not for you.