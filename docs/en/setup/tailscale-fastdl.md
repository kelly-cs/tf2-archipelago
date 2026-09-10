# Fast map downloads with Tailscale

FastDL lets Team Fortress 2 download community maps over HTTPS instead of
squeezing them through the game-server connection. Tailscale Funnel publishes
those files without a webserver for you to maintain or an HTTP port to forward
at the router. Only the server runs Tailscale. Players use an ordinary public
HTTPS address and do not install it.

This changes only map downloads. Keep choosing **local network**, **Steam
relay**, or **forwarded port** for the game server exactly as before.

Funnel requires MagicDNS, HTTPS certificates and Funnel permission on the
tailnet. Tailscale's [Funnel setup page](https://tailscale.com/kb/1223/funnel)
explains those prerequisites. Funnel is public and has bandwidth limits, so
test the largest map before relying on it for an event.

## Windows launcher

1. [Install Tailscale](https://tailscale.com/download/windows), open its tray
   icon and sign in.
2. Open the launcher's **Settings**, then **Networking**.
3. Click **Set up / check Tailscale Funnel**. Approve Funnel in the browser if
   asked, then click the check button again.
4. Turn on **Publish downloads with Tailscale Funnel**, save and press
   **Start**.
5. Look for `public Tailscale Funnel FastDL ready` in the launcher log.

The check is a one-time authorization step. The launcher remembers the setting
and recreates the `/tf` route automatically on every later start. Stop removes
that route, without signing the server out of Tailscale or changing any other
Funnel and Serve routes on the machine.

## Native Linux

Install and sign in to Tailscale on the server using its
[Linux instructions](https://tailscale.com/download/linux), then check Funnel:

```sh
./tf2ap-linux-amd64 -setup-funnel
```

If it prints an approval URL, open that URL in any browser, approve Funnel and
run the command again. This works over SSH; it does not depend on `xdg-open`.
It is the preferred setup method because the launcher also removes its temporary
authorization-check route once the check finishes.

If Tailscale answers `Access denied: serve config denied`, allow your normal
user to manage Funnel with this one-time command:

```sh
sudo tailscale set --operator=$USER
```

Then rerun `./tf2ap-linux-amd64 -setup-funnel` without `sudo`. The launcher
itself does not need to run as root.

To perform the same authorization probe manually, quote the complete static
text target because it contains spaces:

```sh
tailscale funnel --yes --bg --https=443 \
  --set-path=/tf2ap-funnel-setup \
  "text:TF2 Archipelago Funnel setup"
```

Without the quotes, the shell passes `text:TF2`, `Archipelago`, `Funnel`, and
`setup` as four targets, and Tailscale reports `invalid number of arguments
(4)`. Remove the temporary manual route after Funnel is authorized:

```sh
tailscale funnel --https=443 --set-path=/tf2ap-funnel-setup off
```

Choose either interface:

- In the terminal interface, press `,`, open **Networking**, run **Set up /
  check Funnel**, enable **Tailscale FastDL**, and save.
- For a plain terminal, run `./tf2ap-linux-amd64 -configure` and answer yes to
  **Publish map downloads with Tailscale Funnel**.

Then start normally. A service can override the saved setting explicitly:

```sh
TAILSCALE_FASTDL=1 FASTDL_PORT=27080 ./tf2ap-linux-amd64 -console
```

Environment variables apply only to that invocation; put them in the systemd
service environment if that is how the launcher starts. `-status` reports
whether the saved or overridden FastDL uses Funnel.

On every start the launcher checks that Tailscale is connected, discovers its
MagicDNS name and applies this route for the lifetime of the server:

```text
https://server-name.example-tailnet.ts.net/tf
    -> http://127.0.0.1:27080/tf
```

The local HTTP listener is loopback-only. If login or Funnel authorization has
expired, startup stops before SRCDS starts and prints the repair or approval
instructions.

## Route cleanup

The launcher owns two narrowly scoped Funnel routes and leaves other Tailscale
Serve and Funnel configuration alone:

- `-setup-funnel` briefly creates `/tf2ap-funnel-setup` to check authorization,
  then removes it immediately when the check succeeds.
- A running server uses `/tf`. The launcher removes it on **Stop**, **Restart**,
  **Quit**, Ctrl+C or SIGTERM, normal server exit, and a later startup failure
  after the route was created. Restart creates a fresh route for the new run.

Cleanup has its own five-second timeout, so it still runs after the server's
shutdown signal cancels the run. A forced process kill, power loss, or Tailscale
failure can prevent any program from cleaning up. The launcher reports a failed
cleanup in its log; remove a stale route manually with:

```sh
tailscale funnel --https=443 --set-path=/tf off
```

This command removes only the launcher's `/tf` route. It does not disable
Tailscale, sign the machine out, or erase unrelated routes. The saved setting
remains enabled, so the launcher recreates `/tf` on the next Start.

## Docker Compose

The Compose stack uses the existing Caddy FastDL server plus the official
`tailscale/tailscale` sidecar. Caddy alone can read the read-only TF2 game
volume. Tailscale shares only Caddy's network namespace and proxies its
loopback HTTP port.

First enable Funnel for the tailnet and create an auth key in the Tailscale
[Keys page](https://login.tailscale.com/admin/settings/keys). Then set these
values in `.env`:

```ini
COMPOSE_PROFILES=tailscale-fastdl
TAILSCALE_FASTDL=1
TAILSCALE_AUTHKEY=tskey-auth-your-key-here
TAILSCALE_HOSTNAME=tf2-fastdl
FASTDL_BIND=127.0.0.1
```

If `COMPOSE_PROFILES` already contains `selfhost`, use a comma-separated list:

```ini
COMPOSE_PROFILES=selfhost,tailscale-fastdl
```

Start the stack and follow its first login:

```sh
make up
make logs
```

Compose waits for the sidecar to be connected with the `/tf` Funnel active
before starting the game server. The srcds log then prints a line like:

```text
[AP] using Tailscale Funnel FastDL at https://tf2-fastdl.example.ts.net/tf
```

The `tailscale_fastdl_state` volume preserves the device identity and login.
`TS_AUTH_ONCE` prevents needless reauthentication, so after the first
successful start you can remove `TAILSCALE_AUTHKEY` from `.env`. Normal
`make down`, `make up`, image upgrades and host restarts keep working. `make
clean` deliberately deletes every volume, including this identity, and the
next start therefore needs a new auth key.

If the sidecar cannot authenticate or apply Funnel, it stays unhealthy and
Compose does not start SRCDS with an empty download URL. Inspect it with:

```sh
docker compose logs tailscale-fastdl
docker compose exec tailscale-fastdl tailscale funnel status
```

## What friends do

Nothing Tailscale-specific. They join TF2 through the normal address or Steam
relay printed by the launcher. TF2 reads `sv_downloadurl` and fetches each
required asset from the public HTTPS Funnel.

Opening the `/tf` address is only a health check; it intentionally does not
list files. TF2 requests exact paths such as `/tf/maps/example.bsp`.

## What becomes public

Only `maps`, `materials`, `models`, `sound`, `particles`, and `resource` are
served. The launcher and Caddy configurations both exclude `cfg`, SourceMod
plugins, passwords, directory listings, writes, and the rest of the server
installation.
