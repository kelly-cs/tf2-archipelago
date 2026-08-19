# Raw window screenshots

The pictures of the launcher's window in the README come from here. `make
shadows` takes each `.png` in this directory and writes the one beside it in
`docs/images/`, on transparent margins and over a drop shadow.

Nothing on a Linux machine can draw that window: walk is a Win32 binding, so
the capture has to come from a machine running Windows.

## Taking one

Windows 11, Snipping Tool, **Window** mode. It hands back the window with its
rounded corners and nothing outside them, which is what the shadow follows.
`Alt`+`PrintScreen` also works and gives square corners.

Then:

- Shoot the window at its normal size. The picture is scaled down in the
  README, and a maximised window reads as a wall of empty grey.
- Have something in the log. A window with an empty log says nothing about
  what the launcher does.
- Check what is in the shot. The room address, the rcon password and the
  server password are all on screen in Settings, and they are yours.

## The names

The README expects these, and `make shadows` keeps the name:

| File | What it shows |
| --- | --- |
| `launcher-main.png` | The main window: the log, Start and Stop, the join line |
| `launcher-settings-player.png` | Settings, the Player options tab |
| `launcher-settings-server.png` | Settings, the Game server tab |
