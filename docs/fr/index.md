# Mann vs Archipelago

Ce projet transforme un serveur Team Fortress 2 Mann vs Machine en
randomizer. Les classes, les emplacements d'arme et les missions
commencent verrouillés. L'équipe les débloque en réussissant des vagues.
Tout le monde sur le serveur partage les mêmes déblocages.

Le serveur remplit aussi l'équipe RED avec des bots, donc deux joueurs
gagnent une partie que Valve a calibrée pour six.

## Commencer

Sur Windows, téléchargez `tf2ap.exe` depuis la
[dernière version](https://github.com/m-this/tf2-archipelago/releases/latest)
et lancez-le. Il demande l'adresse de votre room Archipelago et installe le
reste.

Avec Docker :

```sh
cp deploy/.env.example .env   # puis réglez SRCDS_RCONPW
make seed                     # envoyez le fichier sur archipelago.gg, ouvrez
                               # une room, puis réglez AP_HOST et AP_PORT
make up
make logs
```

Le premier démarrage télécharge environ 14 Go de fichiers de jeu.

[Installer sur Windows](setup/install-windows.md) et
[Installer avec Docker](setup/install.md) couvrent les deux en détail.

## Lire le livre dans cet ordre

Ce livre s'adresse à l'hébergeur. Il suppose que vous connaissez Mann vs
Machine et que vous n'avez jamais utilisé un randomizer. Il définit chaque
mot avant de l'utiliser.

1. [Archipelago pour les joueurs MvM](archipelago-for-mvm-players.md) — le
   vocabulaire. À lire en premier.
2. [Ce que le randomizer change](what-the-randomizer-changes.md) — ce qui
   diffère d'un serveur MvM normal.
3. [Prérequis](setup/requirements.md) — ce qu'il faut à la machine.
4. [La forme de la partie](setup/shape-of-the-run.md) — la longueur et la
   difficulté d'une soirée.
5. [Créer la session](setup/create-the-session.md) — fabrique la partie et
   la met sur `archipelago.gg`.
6. [Installer sur Windows](setup/install-windows.md) ou
   [Installer avec Docker](setup/install.md) — fait tourner le serveur.
7. [Inviter vos amis](setup/invite-your-friends.md) ouvre le serveur.
   [Les bots de votre équipe](play/defender-bots.md) dit qui remplit les
   places vides.
