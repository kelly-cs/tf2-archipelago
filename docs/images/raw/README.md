# Raw window screenshots

The pictures of the launcher's window in the README come from here. `make
shadows` takes each `.png` in this directory and writes the one beside it in
`docs/images/`, on transparent margins and over a drop shadow.

`make window-captures` fills this directory: it builds `tf2ap.exe`, runs it
under Wine on a virtual display, and photographs the two windows. The window
is Win32, so that is what it takes to draw one on a machine that is not
running Windows. `docs/window-shot.sh` says what has to be installed.

A shot taken by hand on a real Windows machine and dropped in here goes
through the same second half, and is worth taking when Wine draws something
the real thing does not. Windows 11, Snipping Tool, **Window** mode: it hands
back the window with its rounded corners and nothing outside them, which is
what the shadow follows. Check what is in frame first, because the room
address and both passwords are on screen in Settings.

## The names

The README expects these, and both halves keep the name:

| File | What it shows |
| --- | --- |
| `launcher-main.png` | The main window: the log, Start and Stop, the join line |
| `launcher-settings.png` | Settings, the Player options tab |
