# Visual campaign tracker

`index.html` is a static, shared tracker for a TF2 Mann vs Machine
Archipelago room. It reads Archipelago's public tracker APIs, so viewers do not
need the room password and the page does not need its own server application.

## Try it locally

Serve the repository root so the page can also read the generated mission and
weapon catalogues:

```sh
python3 -m http.server 8000
```

Open <http://localhost:8000/tracker/> and either:

- paste the room URL from `archipelago.gg`;
- paste its standard tracker URL or compact tracker ID; or
- choose **View sample run** to explore the interface without a room.

Select a TF2 slot when a multiworld contains more than one. Click a class icon
to see the weapon buffs that class can equip, their individual stack counts,
and the combined level for each weapon. A live tracker refreshes once a minute;
**Refresh** updates it immediately.

## Publish it as a static page

There is no build step. Publish these paths under the same site root:

```text
tracker/index.html
apworld/tf2_mvm/data/missions.json
apworld/tf2_mvm/data/weapon_classes.json
```

That layout works from a repository-root GitHub Pages deployment. The page
falls back to the catalogues on this repository's `main` branch when the local
copies are unavailable. TF2 class and item icons load from the Official Team
Fortress Wiki.

Rooms must have tracking enabled. Seeds made with an older apworld do not put
their precollected inventory in public slot data, so randomly selected starting
classes can appear locked. Generate a new seed with the apworld from this branch
to include that inventory.

## How it gets its data

The page calls Archipelago's public room-status, static-tracker, slot-data,
tracker-state, and datapackage endpoints. It does not connect to the TF2 server
or the bridge and it never sends a room password.

Mission metadata comes from `missions.json`. `weapon_classes.json` is generated
from the same Go weapon catalogue used by the launcher and plugin, which keeps
the class popup aligned with weapons each class can actually equip. Regenerate
both catalogues after changing gamedata:

```sh
go generate ./gamedata
```
