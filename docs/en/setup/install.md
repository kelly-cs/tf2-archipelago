# Install with Docker

This is the Docker path. It works on any operating system with Docker. On
Windows, [Install on Windows](install-windows.md) is easier: one exe, no
Docker.

There are two ways to run it:

- [From a clone of the repository](#from-a-clone), with `make`.
- [From two downloaded files](#without-the-repository), with `docker compose`
  and the published images.

Both give the same admin page as the launcher, at `http://127.0.0.1:8477`.

## From a clone

Run everything from the root of the repository.

### 1. Write the configuration file

```sh
cp deploy/.env.example .env
```

`.env` is the one file you edit. Git ignores it. Every setting in it is
described in [Run options](shape-of-the-run.md).

### 2. Set the console password

Open `.env` and set `SRCDS_RCONPW`:

```sh
SRCDS_RCONPW=pick-something-long
```

The password unlocks the remote console of the game server. Only the host
needs it.

### 3. Give the stack a room

The stack needs a room address. Two choices:

- You have no room yet. Run `make seed`, upload the file it writes to
  `archipelago.gg`, and create a room. See [Create the session](create-the-session.md).
- You already generate your own multiworlds. Use the release's `.apworld` and
  point the stack at your room.

Then write the room address into `.env`:

```sh
AP_HOST=archipelago.gg
AP_PORT=12345
AP_TLS=true
```

`SRCDS_RCONPW`, `AP_HOST` and `AP_PORT` have no default. The stack refuses to
start without them, and it prints which one is missing.

To try the stack without a room, set `TF2AP_TEST_MODE=1`. The bridge then plays
a multiworld of one on this machine and ignores `AP_HOST` and `AP_PORT`.

### 4. Start the stack

```sh
make up
make logs
```

`make up` builds two images and starts the containers. `make logs` follows
their output. Ctrl-C stops following, not the stack.

The first start does this, in order:

1. Builds the plugin and the bridge. A few minutes.
2. Downloads about 14 GB of game files. This is the long part.
3. Installs the plugin. The log says `[AP] installed the plugin and ripext`.
4. Joins the room. The log says `connected to archipelago slot=tf2`.

Every later start takes seconds.

### The commands

| Command | What it does |
| --- | --- |
| `make seed` | Generate a session into `seed/`, to upload to `archipelago.gg` |
| `make up` | Start the stack |
| `make logs` | Follow the output of the services |
| `make ps` | List the containers and their state |
| `make down` | Stop the stack. Keeps the game files and the run. |
| `make restart` | `make down`, then `make up` |
| `make build` | Rebuild the images |
| `make clean` | Stop the stack and delete every volume, including the 14 GB of game files |

Use `make down` to stop. `make clean` deletes the game files.

## Without the repository

Every release attaches a `compose.yaml` that uses the published images, and an
`env.example` to go with it.

```sh
mkdir mann-vs-archipelago && cd mann-vs-archipelago
base=https://github.com/m-this/tf2-archipelago/releases/latest/download
curl -fsSLO "$base/compose.yaml"
curl -fsSL -o .env "$base/env.example"
```

Set `SRCDS_RCONPW` in `.env`. Then generate a session and start:

```sh
docker compose --profile seed run --rm seed   # writes ./seed
docker compose up -d
docker compose logs -f
```

Upload the file from `seed/`, create a room, and write the room's port into
`AP_PORT`. See [Create the session](create-the-session.md).

After every edit of `.env`, apply it with:

```sh
docker compose up -d --force-recreate
```

`docker compose up -d` alone does not restart containers it considers
unchanged.

The `compose.yaml` pins the images to the release it came from. To move to
another version, set `TF2AP_VERSION` in `.env`, then:

```sh
docker compose pull
docker compose up -d --force-recreate
```

## The admin page

The stack serves the same page as the launcher at:

```text
http://127.0.0.1:8477
```

`TF2AP_ADMIN_PORT` in `.env` changes the port. The page stays on loopback
because it can send console commands. To reach a remote server, use an SSH
port forward. Do not publish this port.

- The **Settings** tab writes to `.env`. Container settings apply after
  `docker compose up -d --force-recreate`. Seed options apply at the next
  `make seed`.
- **Stop** and **Restart** print the commands to run. The admin container has
  no Docker socket, so it cannot run them itself.
- Set **Join address** on the Game server page to the address your friends
  connect to. If Docker runs inside WSL and TF2 runs on Windows, use the WSL
  address from `hostname -I`.

## The services

| Service | What it does | Ports |
| --- | --- | --- |
| `srcds` | The Team Fortress 2 server and the plugin | `27015/udp` and `27015/tcp` |
| `bridge` | Holds the session with the room and answers the plugin | none, loopback only |
| `admin` | The admin page | `8477/tcp` on loopback |
| `fastdl` | Serves maps to joining players over HTTP | `27080/tcp` |
| `archipelago` | Hosts the session on this machine. Only with `COMPOSE_PROFILES=selfhost`. | see `deploy/compose.yml` |
| `tailscale-fastdl` | Publishes the map downloads over Tailscale Funnel. Optional. | none |

The bridge shares the network namespace of the game server. Restarting the game
server restarts the bridge too. That costs seconds, not progress: the bridge
writes every check to disk.

## Where the stack keeps things

| Volume | Holds | Delete it to |
| --- | --- | --- |
| `tf2-archipelago_tf2game` | The 14 GB of game files, SourceMod and the plugin | Download everything again |
| `tf2-archipelago_bridgestate` | The checks and the unlocks of the run | Nothing useful. The bridge rebuilds the checks from the room. |
| `tf2-archipelago_apoutput` | The session, with `COMPOSE_PROFILES=selfhost` only | [Start a new run](../operate/start-a-new-run.md) |
| `tf2-archipelago_tailscale_fastdl_state` | The Tailscale identity of the Funnel node | Sign in to Tailscale again |

The sessions are files in `seed/`. Git ignores that directory, and nothing
deletes it for you.

Next: [Create the session](create-the-session.md).
