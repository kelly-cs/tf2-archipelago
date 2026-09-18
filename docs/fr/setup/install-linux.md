# Installer sur Linux

Un seul fichier. C'est le même programme que le lanceur Windows, avec la même
interface dans le navigateur et les mêmes réglages. Il marche en SSH.

## 1. Télécharger et lancer

```sh
curl -fsSLO https://github.com/m-this/tf2-archipelago/releases/latest/download/tf2ap-linux-amd64
chmod +x tf2ap-linux-amd64
./tf2ap-linux-amd64
```

Un navigateur s'ouvre sur le lanceur. Sur une machine sans bureau, ajoutez
`-no-browser` et ouvrez l'adresse affichée depuis où vous êtes, ou redirigez
le port en SSH.

### Bibliothèques 32 bits

Le serveur dédié TF2 est un programme 32 bits. Sur Debian et Ubuntu :

```sh
sudo dpkg --add-architecture i386
sudo apt update
sudo apt install lib32gcc-s1 lib32stdc++6 libcurl3t64-gnutls:i386
```

Fedora appelle la bibliothèque C `glibc.i686`, Arch l'appelle `lib32-glibc`.
Quand une bibliothèque manque, SteamCMD ou le serveur affiche son nom.

## 2. Appuyer sur Start

Le premier démarrage installe SteamCMD, le serveur dédié TF2, Metamod:Source,
SourceMod, le plugin et les bots. Il télécharge environ 14 Go. Tous les
démarrages suivants prennent quelques secondes.

L'écran est celui décrit dans [Installer sur Windows](install-windows.md#lécran).

![L'onglet Play](../../images/launcher-session.png)

## 3. Créer la session

1. Installez l'[application Archipelago](https://github.com/ArchipelagoMW/Archipelago/releases).
   Le lanceur cherche `ArchipelagoGenerate` dans `PATH`, une application
   extraite dans `~/Applications/Archipelago`, `~/.local/opt/Archipelago`,
   `~/Archipelago`, `~/Downloads/Archipelago`, `/opt/Archipelago`,
   `/usr/local/lib/Archipelago` et `/ap`, et un `Archipelago*.AppImage` dans
   `~/Applications`, `~/.local/bin`, `~/Downloads`, `/opt` et
   `/usr/local/bin`. Ailleurs, réglez **Archipelago app** dans **Settings**,
   puis **Player options**.
2. Ouvrez **Settings**, puis **Player options**. Choisissez les
   [options de la partie](shape-of-the-run.md).
3. Appuyez sur **Generate seed**.
4. Envoyez le résultat sur [archipelago.gg/uploads](https://archipelago.gg/uploads)
   et créez une room.
5. Mettez l'adresse de la room dans **Settings**, puis **Archipelago room**.
   Enregistrez et appuyez sur **Restart**.

Sans bureau, `./tf2ap-linux-amd64 -yaml tf2.yaml` écrit le fichier joueur.
Générez la seed avec l'application Archipelago sur n'importe quelle machine.
Voir [Créer la session](create-the-session.md).

## 4. Inviter vos amis

La ligne **Join** montre la ligne de connexion. Par défaut, seul le réseau
local atteint le serveur. Voir [Inviter vos amis](invite-your-friends.md).

## Le lancer comme un service

- `-console` affiche le journal et ne dessine rien, ce qui convient à
  `systemd` ou à une session `screen`. Ctrl+C l'arrête.
- `-configure` modifie chaque réglage dans le terminal.
- `-setup-funnel` vérifie Tailscale Funnel. Voir
  [Téléchargement rapide des cartes avec Tailscale](tailscale-fastdl.md).

## Référence

### Ligne de commande

| Commande | Ce qu'elle fait |
| --- | --- |
| `tf2ap-linux-amd64` | Installer ce qui manque, puis lancer |
| `tf2ap-linux-amd64 -room <hôte:port>` | Régler d'abord l'adresse de la room |
| `tf2ap-linux-amd64 -no-browser` | Afficher l'adresse au lieu d'ouvrir un navigateur |
| `tf2ap-linux-amd64 -addr hôte:port` | Servir sur une adresse fixe |
| `tf2ap-linux-amd64 -console` | Afficher le journal et rien d'autre |
| `tf2ap-linux-amd64 -configure` | Modifier chaque réglage dans le terminal, puis quitter |
| `tf2ap-linux-amd64 -setup-funnel` | Vérifier Tailscale Funnel et afficher une URL d'approbation si besoin |
| `tf2ap-linux-amd64 -install` | Installer ou réparer le serveur, puis quitter |
| `tf2ap-linux-amd64 -status` | Afficher les réglages et l'état de l'installation |
| `tf2ap-linux-amd64 -yaml <chemin>` | Écrire le fichier joueur Archipelago, puis quitter |
| `tf2ap-linux-amd64 -env` | Lister les variables d'environnement lues, puis quitter |
| `tf2ap-linux-amd64 -version` | Afficher la version et les versions des outils |

### Variables d'environnement

Chaque réglage lit aussi une variable d'environnement, avec les noms donnés
dans [Les options de la partie](shape-of-the-run.md). Une variable l'emporte
sur les réglages enregistrés :

```sh
AP_ROOM=archipelago.gg:12345 SRCDS_BOT_TEAM_SIZE=4 ./tf2ap-linux-amd64
```

### Où il range les choses

| Chemin | Contenu |
| --- | --- |
| `~/tf2-archipelago/` | Les fichiers de jeu, SourceMod et SteamCMD |
| `~/tf2-archipelago/tf2.yaml` | Le fichier joueur |
| `~/tf2-archipelago/bridge-state/` | Les checks et les déblocages de la partie |
| `~/.config/tf2ap/config.json` | Vos réglages |

`TF2AP_INSTALL_ROOT` déplace les trois premiers sur un autre disque.

Suite : [Créer la session](create-the-session.md).
