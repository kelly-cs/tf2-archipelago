<p align="center">
  <a href="https://github.com/m-this/tf2-archipelago/releases/latest/download/tf2ap.exe">
    <img alt="Download tf2ap.exe for Windows" src="https://img.shields.io/badge/Download-tf2ap.exe%20for%20Windows-2ea44f?style=for-the-badge&logo=windows&logoColor=white">
  </a>
  <a href="https://github.com/m-this/tf2-archipelago/releases/latest/download/tf2ap-linux-amd64">
    <img alt="Download tf2ap-linux-amd64 for Linux" src="https://img.shields.io/badge/Download-tf2ap--linux--amd64-1b1f27?style=for-the-badge&logo=linux&logoColor=white">
  </a>
</p>

<!-- CHANGES -->

## Windows

Download **`tf2ap.exe`** above. Run it. It installs everything else.

Windows will warn you, and it is a false positive. SmartScreen calls the exe an
unrecognized app; Defender sometimes flags it. Click **More info**, then **Run
anyway**. The launcher unpacks archives into your TF2 directory, writes
SourceMod's DLLs there, downloads a game server and starts it, which is what an
installer does and also what a dropper does, and the file is not signed yet, so
a scanner has nothing to weigh against the heuristic.
[TODO item 13](https://github.com/m-this/tf2-archipelago/blob/main/TODO.md#13-sign-the-windows-exe-open)
is the fix. Check the file below instead of trusting it, or build it yourself
with `make launcher`.

[Full guide](https://m-this.github.io/tf2-archipelago/en/setup/install-windows.html)

## Linux

Download **`tf2ap-linux-amd64`**. Run it. It installs everything else.

```sh
chmod +x tf2ap-linux-amd64
./tf2ap-linux-amd64
```

[Full guide](https://m-this.github.io/tf2-archipelago/en/setup/install-linux.html)

## Docker

Download **`compose.yaml`** and **`.env.example`**:

```sh
mkdir mann-vs-archipelago && cd mann-vs-archipelago
base=https://github.com/m-this/tf2-archipelago/releases/latest/download
curl -fsSLO "$base/compose.yaml"
curl -fsSL -o .env "$base/.env.example"
```

Set `SRCDS_RCONPW` in `.env`, then:

```sh
docker compose --profile seed run --rm seed   # writes ./seed
docker compose up -d
```

Upload the file from `seed/` at
[archipelago.gg/uploads](https://archipelago.gg/uploads) and put the room's
port into `AP_PORT`.

[Full guide](https://m-this.github.io/tf2-archipelago/en/setup/install.html)

## Other assets

| File | For |
| --- | --- |
| `tf2_mvm.apworld` | The official Archipelago app. |
| `tf2_archipelago.smx` | A SourceMod server the launcher and compose file don't manage. |
| `tf2-defender-bots.zip` | The bots for that same server. Unzip into `tf/`. |
| `meta.json`, `items.json`, `missions.json` | The item and check tables, for reference. |

## Checking what you downloaded

**`SHA256SUMS`** is attached to this release. It holds the hash of every file
here, computed by the workflow that built them.

```powershell
Get-FileHash tf2ap.exe -Algorithm SHA256    # Windows
```

```sh
sha256sum -c SHA256SUMS --ignore-missing    # Linux
```

`tf2ap.exe` is also on VirusTotal. Search its SHA-256 at
[virustotal.com](https://www.virustotal.com/gui/home/search) to see what every
engine says about this exact build. Expect a small number of heuristic and
machine-learning detections and no named malware family; the section above says
why. The source of all of it is in the repository, and `make launcher` rebuilds
the exe on your own machine.
