# User Guide
Once the tool is running you have both the listplayers validation functionality and a minimal command console.

## Player validation
Players are automatically validated in the background when you run listplayers.
The output table shows from left to right:
1. The local player number
2. Their PlayFab ID
3. The date when the account was created
4. Platform information (see below)
5. Their current Display Name
6. A list of their known aliases

Suspicious players will be marked yellow and have a kick and ban command that can be used to remove them from the current server.
Every admin sees these commands, but that doesn't mean they are necessarily banned across all servers.
It's up to the admins to decide how to handle them.

Wanted players are marked in red, meaning they have an outstanding ban.
Admins should use the provided command to ban them immediately.
If you think someone is banned that shouldn't be, then you can open a ticket on the SAK discord.

Note that even after a player has been banned or kicked they might still show up in the player validation.
This is because even players who have left are still contained in the servers listplayers table.
This might cause banned players to be reported multiple times, even when already gone from the server.

### Platform information
This property of players is not 100% accurate and should be treated as an estimate.
There currently are 3 different possible platform types:
1. `G` is probably GamePass or some other console platform. These players can't use chat (without cheats).
2. `X` is probably a genuine Xbox. It's unsure if these accounts are created via GamePass or via One-Time purchase.
3. No platform likely means Steam or Epic

It is currently unclear in which category PSN players might show up.
If you see a player where the estimate does not match their platform please let me know by opening an issue.

## Player Actions
There are quick commands that can be used to manage player records.
Most commands use the local player number instead of having to copy/paste their PlayFab IDs.

### Ban command
This command adds a player that is currently in the lobby to the global wanted board.
The reasons can technically be any single word, but most of the time it's recommended to use one of the predefined reasons.
This way the ban time and ban message is automatically added to the global ban list.
This command copies an in-game `banbyid` command into your clipboard.
Admins should use that command to ban the player, but may also put a custom ban reason or ban time.
```
ban <player-number> <reasons...>
// Example:
ban 22 cheating ffa player_impersonation
```

You can also use the `banbyid` command which will work the same as the `ban` command but using a PlayFab ID.
This can be used to ban a player that is not currently in the lobby.
```
banbyid <playfab-id> <reasons...>
// Example:
banbyid EAE0E3E2F35692CE cheating harassment player_impersonation
```

### Trust command
Due to free Account exploits and free Accounts via Epic it can happen that a lot of players are marked suspicious by default.
You can use the trust command to make them not suspicious.
Only use this when you know that this account is trustworthy, for example the alt account of a known player.
```
trust <player-number>
// Example:
trust 22
```

### Unban command
When an account was banned mistakenly you can use the `unbanbyid` command.
That will remove them from the global ban list and return an in-game command to your clipboard.
Note that unbanning an account requires a server restart to take effect.
```
unbanbyid <playfab-id>
// Example:
unbanbyid 1512247D9C9C2634
```

### Kick command
This command is a convenience function that formats a kick command into your clipboard, so you don't need to copy/paste PlayFab IDs.
It does not have an effect on the global ban list or any other admins running this tool.
```
kick <player-number>
// Example:
kick 22
```

## A note on Fonts
A lot of chivalry players use special unicode character in their names.
Depending on your version of Windows and Powershell/Terminal these characters are not printed correctly by default.
They will show up as question marks or empty boxes instead when they are missing with your current font.
To see these characters choose a different Font for your Terminal that can display these characters.

