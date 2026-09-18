# Fast map downloads with Tailscale

FastDL lets a joining player download a community map over HTTPS. The game
server's own transfer is slow and sometimes restarts the download forever. Tailscale Funnel publishes those files without a web server to run
and without a port to forward on the router.

- Only the server runs Tailscale. Players use a public HTTPS address and
  install nothing.
- This changes only the map downloads. **Who can reach it** stays what it was.
- Funnel needs MagicDNS, HTTPS certificates and Funnel permission on your
  tailnet. Tailscale's [Funnel page](https://tailscale.com/kb/1223/funnel)
  explains them.
- Funnel is public and has bandwidth limits. Test the largest map before an
  event.

## Windows

1. [Install Tailscale](https://tailscale.com/download/windows) and sign in
   from its tray icon.
2. In the launcher, open **Settings**, then **Networking**.
3. Press **Set up / check Tailscale Funnel**. If a browser page asks you to
   approve Funnel, approve it, then press the button again.
4. Tick **Tailscale FastDL**. Save, then press **Start**.
5. Look for `public Tailscale Funnel FastDL ready` in the log.

The check is a one-time step. The launcher remembers the setting and recreates
the route at every start. **Stop** removes the route and nothing else.

## Linux

1. Install and sign in to Tailscale with its
   [Linux instructions](https://tailscale.com/download/linux).
2. Run the check:

```sh
./tf2ap-linux-amd64 -setup-funnel
```

If it prints an approval URL, open it in any browser, approve Funnel, and run
the command again. This works over SSH.

If Tailscale answers `Access denied: serve config denied`, let your user
manage Funnel, then run the check again without `sudo`:

```sh
sudo tailscale set --operator=$USER
```

Then turn it on:

- With a desktop: **Settings**, then **Networking**, tick **Tailscale FastDL**,
  save.
- Without: `./tf2ap-linux-amd64 -configure` and answer yes to the Funnel
  question.

A service can force it for one run:

```sh
TAILSCALE_FASTDL=1 FASTDL_PORT=27080 ./tf2ap-linux-amd64 -console
```

At every start the launcher checks that Tailscale is connected and applies
this route for the life of the server:

```text
https://server-name.example-tailnet.ts.net/tf  ->  http://127.0.0.1:27080/tf
```

If the login or the Funnel approval expired, the launcher stops before the
game server starts and prints what to do.

### Route cleanup

The launcher owns two routes and touches nothing else in Tailscale:

- `/tf2ap-funnel-setup`, created by `-setup-funnel` and removed as soon as
  the check succeeds.
- `/tf`, created at **Start** and removed at **Stop**, **Restart**, **Quit**,
  Ctrl+C, and normal exit.

A forced kill or a power loss can leave `/tf` behind. Remove it with:

```sh
tailscale funnel --https=443 --set-path=/tf off
```

## Docker

The stack includes the official `tailscale/tailscale` container beside the
Caddy FastDL server. You do not install Tailscale on the host.

1. Set these in `.env`, or tick **Tailscale FastDL** on the admin page:

```ini
TAILSCALE_FASTDL=1
TAILSCALE_HOSTNAME=tf2-fastdl
FASTDL_BIND=127.0.0.1
```

2. Apply and open the admin page:

```sh
docker compose up -d --force-recreate
```

3. On **Settings**, then **Networking**, press **Set up / check Funnel**. The
   first press gives a Tailscale sign-in link. Sign in, come back, press
   again. If the tailnet never used Funnel, the second press gives an approval
   link. Approve, press once more.
4. Wait for the ready message. The game server log then prints:

```text
[AP] using Tailscale Funnel FastDL at https://tf2-fastdl.example.ts.net/tf
```

The game server waits for the route before it starts. If Tailscale cannot
sign in, the game server waits, and the admin page stays available. Inspect
the container with:

```sh
docker compose logs tailscale-fastdl
```

The `tailscale_fastdl_state` volume keeps the device identity. `make clean`
deletes it, and the next start needs a new sign-in.

## What friends do

Nothing. They join with the normal connect line. The game reads the download
address from the server and fetches each missing file from it.

## What becomes public

The server publishes only `maps`, `materials`, `models`, `sound`, `particles`
and `resource`. Configuration files, plugins and passwords are not. The address
lists no files.
