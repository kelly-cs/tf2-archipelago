# Invite your friends

Your friends need three things:

1. A connect line.
2. The server password, if you set one.
3. One sentence: the classes and the weapons start locked, clearing waves
   unlocks them, and everybody shares.

They install nothing. A normal Team Fortress 2 client is enough.

## Choose who can reach the server

Launcher page: **Settings**, then **Networking**, row **Who can reach it**.
In `.env`, it is `SRCDS_REACH`.

| Choice | `.env` | Who can join | Needs a login token | Needs a port on the router |
| --- | --- | --- | --- | --- |
| **Local network** | `lan` | People on the same network. The default. | No | No |
| **Forwarded port** | `port` | Anybody with your public address. | Yes | Yes, UDP and TCP |
| **Steam relay** | `steam` | Anybody with the relay address. | Yes | No |

**Steam relay is not finished.** No run has taken it all the way to a client
that joined. Use **Forwarded port** for friends over the internet.

A server with no login token stays on the local network, whatever you chose.
The log says so.

## The login token

**Forwarded port** and **Steam relay** log the server in to Steam. That needs
a Game Server Login Token.

1. Open [steamcommunity.com/dev/managegameservers](https://steamcommunity.com/dev/managegameservers).
2. Create a token for App ID `440`.
3. Paste it into **Login token** on the **Networking** page, or into
   `SRCDS_TOKEN` in `.env`.

The token is not a password anybody types. It identifies the server to Steam.
Without one, the server never gets a Steam session and refuses every player
with no useful message.

## Local network

Nothing to set up. The **Join** line shows the address. Your friends type it in
the developer console:

```text
connect 192.168.1.20:27015
```

The game refuses players from another network with `LAN servers are restricted
to local clients (class C)`. A guest Wi-Fi, a VPN and a server in a container
count as another network.

## Forwarded port

1. On your router, forward the game port to this machine, on **UDP and TCP**.
   The default port is `27015`.
2. Open the same port in the machine's firewall.
3. Set **Join address** on the **Game server** page to your public address.
4. Give out the connect line:

```text
connect your.public.address:27015
```

UDP carries the game. A rule that forwards only TCP answers the server
browser and drops every join.

## Steam relay

The server asks Valve for a relay address and prints it in the log:

```text
FakeIP allocation succeeded: 169.254.13.42:20232, 20233
```

The launcher shows it on the **Join** line. Your friends connect to the first
address. You forward nothing, and your own address stays hidden.

The address changes at every start. Send the one from this run.

## The server password

Set **Server password** on the **Game server** page, or `SRCDS_PW` in `.env`.
Empty lets anybody with the address join.

Your friends type this before they connect:

```text
password friends-only
connect ...
```

Do not confuse it with the console password, `SRCDS_RCONPW`. That one runs
admin commands. Never give it out.

## Staying off the public list

A server with a login token can appear in the public server browser. Set a
server password, and strangers who find it cannot get in.

## The developer console

The console is off by default in Team Fortress 2. To turn it on:

1. Open **Options**, then **Keyboard**, then **Advanced**.
2. Tick **Enable developer console**.

The key that opens it is `` ` ``, left of the `1` key on a US keyboard.

## When a player cannot connect

Check these in order:

1. **The token.** With **Forwarded port** or **Steam relay** and no token, the
   server refuses everybody. The console says
   `Could not establish connection to Steam servers`.
2. **The network.** With **Local network**, the player has to be on the same
   network. See [Local network](#local-network).
3. **The port.** With **Forwarded port**, the router has to forward the port on UDP
   as well as TCP.
4. **The address.** With **Steam relay**, the address is the one from this
   run.
5. **The machine.** A laptop that sleeps is a server that is down.

Next: [The first session](../play/first-session.md).
