# Create the session

An Archipelago session is a **seed** hosted in a **room**. The server plays a
room that already exists. This page makes one.

1. Choose the [run options](shape-of-the-run.md).
2. Generate the seed with the Archipelago app.
3. Upload the seed to `archipelago.gg` and create a room.
4. Give the server the room address.

Mann vs Machine is not one of the games that ship with Archipelago. So the
website cannot generate the seed for you. Your machine generates it, and the
website hosts it.

## 1. Choose the run options

The run options decide how long the run is, how hard it is, and what ends it.
The seed keeps them. To change one later, you need a new seed and a new room.

- **Launcher:** open **Settings**, then **Player options**, **Rewards**,
  **Balancing** and **Missions**.
- **Docker:** edit the `MVM_` lines in `.env`.
- **Archipelago app by hand:** edit the YAML file. See
  [With the Archipelago app](#with-the-archipelago-app).

[Run options](shape-of-the-run.md) describes every option.

## 2. Generate the seed

### With the launcher

1. Install the [Archipelago app](https://github.com/ArchipelagoMW/Archipelago/releases).
   Use the same version the launcher pins. `tf2ap.exe -version` prints it.
2. In the launcher, open **Settings**, then **Player options**.
3. Press **Generate seed**. The launcher writes `tf2.yaml`, runs the
   generator, and opens the folder with the result. The result is a `.zip`
   named like `AP_53174869021847362095.zip`.

If **Generate seed** says it cannot find the Archipelago app, set
**Archipelago app** on the same page to the app's folder.

**Check Run Selection**, on the **Missions** page, checks the pool before you
generate. It tells you whether the missions you picked hold enough checks for
the items of the run.

### With the Archipelago app

Use this to play with people in other games, or to edit the YAML by hand.

1. Download `tf2_mvm.apworld` from the
   [release](https://github.com/m-this/tf2-archipelago/releases/latest).
2. Double-click it, or copy it into the app's `custom_worlds/` folder.
3. Get a player file. Either run `tf2ap.exe -yaml tf2.yaml` in the launcher,
   or in the app's Launcher press **Generate Template Options** and take
   `Team Fortress 2 Mann vs Machine.yaml`.
4. Edit the file. [Run options](shape-of-the-run.md).
5. Put it into the app's `Players/` folder, beside the files of the other
   players.
6. Run **Generate**. The result is in the app's `output/` folder.

The `name` in the file is the slot name. The launcher's **Slot name**, in
**Settings**, then **Archipelago room**, has to match it. The default is `tf2`.

### With Docker

```sh
make seed
```

The command writes the seed into `seed/` and prints its name. Keep the files
in `seed/`. A room you lose comes back from its file.

If the upload fails because of the version, read the Archipelago version in
the footer of the website. Set `ARCHIPELAGO_VERSION` in `deploy/env/versions.env`
to it and run `make seed` again.

## 3. Upload the seed and create a room

1. Open [archipelago.gg/uploads](https://archipelago.gg/uploads).
2. Upload the `.zip`.
3. Click **Create New Room**.

The website asks for no account. The room page shows:

- the room address, like `archipelago.gg:12345`,
- a link to the tracker, where your players watch the run from a browser.

Every new room gets a new port. Anybody with the address can join the room, so
set a password on the room page if the address leaves your friends.

## 4. Give the server the room address

- **Launcher:** paste the address into **Settings**, then **Archipelago room**.
  Put the room password there too, if you set one. Save, then press
  **Restart**.
- **Docker:** write the two halves into `.env`, then `make restart`:

```sh
AP_HOST=archipelago.gg
AP_PORT=12345
AP_TLS=true
AP_PASSWORD=
```

The status line of the launcher, or the bridge log, then says `connected to archipelago`.
The room page says `tf2 (Team #1) playing Team Fortress 2 Mann vs Machine has
joined`.

## Host the session yourself

The Docker stack can host the room on your machine. Put these lines in `.env`:

```sh
COMPOSE_PROFILES=selfhost
AP_HOST=archipelago
AP_PORT=38281
AP_TLS=false
```

`make up` then starts a third container that generates the seed at its first
start and hosts it. You upload nothing.

What it costs:

- Your players get no room page and no tracker.
- A player in another game needs a second public port. `deploy/compose.yml`
  says which.

Next: [Run options](shape-of-the-run.md), or [Invite your friends](invite-your-friends.md)
if the session is ready.
