# Installer avec Docker

C'est la voie Docker. Elle marche sur tout système avec Docker. Sur Windows,
[Installer sur Windows](install-windows.md) est plus simple : un exe, pas de
Docker.

Il y a deux façons de la lancer :

- [Depuis un clone du dépôt](#depuis-un-clone), avec `make`.
- [Depuis deux fichiers téléchargés](#sans-le-dépôt), avec `docker compose`
  et les images publiées.

Les deux donnent la même page d'administration que le lanceur, sur
`http://127.0.0.1:8477`.

## Depuis un clone

Lancez tout depuis la racine du dépôt.

### 1. Écrire le fichier de configuration

```sh
cp deploy/.env.example .env
```

`.env` est le seul fichier que vous modifiez. Git l'ignore.
[Les options de la partie](shape-of-the-run.md) décrit chaque réglage qu'il
contient.

### 2. Régler le mot de passe de la console

Ouvrez `.env` et réglez `SRCDS_RCONPW` :

```sh
SRCDS_RCONPW=choisissez-quelque-chose-de-long
```

Ce mot de passe ouvre la console distante du serveur de jeu. Seul l'hébergeur
en a besoin.

### 3. Donner une room à la pile

La pile a besoin d'une adresse de room. Deux cas :

- Vous n'avez pas encore de room. Lancez `make seed`, envoyez le fichier écrit
  sur `archipelago.gg` et créez une room. Voir [Créer la session](create-the-session.md).
- Vous générez déjà vos propres multiworlds. Utilisez le `.apworld` de la
  version et pointez la pile vers votre room.

Écrivez ensuite l'adresse de la room dans `.env` :

```sh
AP_HOST=archipelago.gg
AP_PORT=12345
AP_TLS=true
```

`SRCDS_RCONPW`, `AP_HOST` et `AP_PORT` n'ont pas de valeur par défaut. La pile
refuse de démarrer sans eux, et elle dit lequel manque.

Pour essayer la pile sans room, réglez `TF2AP_TEST_MODE=1`. Le bridge joue
alors un multiworld d'un seul joueur sur cette machine et ignore `AP_HOST` et
`AP_PORT`.

### 4. Démarrer la pile

```sh
make up
make logs
```

`make up` construit deux images et démarre les conteneurs. `make logs` suit
leur sortie. Ctrl-C arrête le suivi, pas la pile.

Le premier démarrage fait ceci, dans l'ordre :

1. Il compile le plugin et le bridge. Quelques minutes.
2. Il télécharge environ 14 Go de fichiers de jeu. C'est la partie longue.
3. Il installe le plugin. Le journal dit `[AP] installed the plugin and ripext`.
4. Il rejoint la room. Le journal dit `connected to archipelago slot=tf2`.

Tous les démarrages suivants prennent quelques secondes.

### Les commandes

| Commande | Ce qu'elle fait |
| --- | --- |
| `make seed` | Générer une session dans `seed/`, à envoyer sur `archipelago.gg` |
| `make up` | Démarrer la pile |
| `make logs` | Suivre la sortie des services |
| `make ps` | Lister les conteneurs et leur état |
| `make down` | Arrêter la pile. Garde les fichiers de jeu et la partie. |
| `make restart` | `make down`, puis `make up` |
| `make build` | Reconstruire les images |
| `make clean` | Arrêter la pile et supprimer chaque volume, y compris les 14 Go de fichiers de jeu |

Utilisez `make down` pour arrêter. `make clean` supprime les fichiers de jeu.

## Sans le dépôt

Chaque version joint un `compose.yaml` qui utilise les images publiées, et un
`env.example` qui va avec.

```sh
mkdir mann-vs-archipelago && cd mann-vs-archipelago
base=https://github.com/m-this/tf2-archipelago/releases/latest/download
curl -fsSLO "$base/compose.yaml"
curl -fsSL -o .env "$base/env.example"
```

Réglez `SRCDS_RCONPW` dans `.env`. Générez ensuite une session et démarrez :

```sh
docker compose --profile seed run --rm seed   # écrit ./seed
docker compose up -d
docker compose logs -f
```

Envoyez le fichier de `seed/`, créez une room, et écrivez le port de la room
dans `AP_PORT`. Voir [Créer la session](create-the-session.md).

Après chaque modification de `.env`, appliquez-la avec :

```sh
docker compose up -d --force-recreate
```

`docker compose up -d` seul ne redémarre pas les conteneurs qu'il considère
inchangés.

Le `compose.yaml` fixe les images à la version d'où il vient. Pour passer à
une autre version, réglez `TF2AP_VERSION` dans `.env`, puis :

```sh
docker compose pull
docker compose up -d --force-recreate
```

## La page d'administration

La pile sert la même page que le lanceur sur :

```text
http://127.0.0.1:8477
```

`TF2AP_ADMIN_PORT` dans `.env` change le port. La page reste sur l'adresse
locale parce qu'elle peut envoyer des commandes de console. Pour atteindre un
serveur distant, utilisez une redirection de port SSH. Ne publiez pas ce port.

- L'onglet **Settings** écrit dans `.env`. Les réglages des conteneurs
  s'appliquent après `docker compose up -d --force-recreate`. Les options de
  la seed s'appliquent au prochain `make seed`.
- **Stop** et **Restart** affichent les commandes à lancer. Le conteneur
  d'administration n'a pas de socket Docker, donc il ne peut pas les lancer
  lui-même.
- Réglez **Join address** sur la page Game server avec l'adresse à laquelle
  vos amis se connectent. Si Docker tourne dans WSL et TF2 sur Windows,
  utilisez l'adresse WSL donnée par `hostname -I`.

## Les services

| Service | Ce qu'il fait | Ports |
| --- | --- | --- |
| `srcds` | Le serveur Team Fortress 2 et le plugin | `27015/udp` et `27015/tcp` |
| `bridge` | Tient la session avec la room et répond au plugin | aucun, adresse locale seulement |
| `admin` | La page d'administration | `8477/tcp` sur l'adresse locale |
| `fastdl` | Sert les cartes aux joueurs qui rejoignent, en HTTP | `27080/tcp` |
| `archipelago` | Héberge la session sur cette machine. Seulement avec `COMPOSE_PROFILES=selfhost`. | voir `deploy/compose.yml` |
| `tailscale-fastdl` | Publie le téléchargement des cartes par Tailscale Funnel. Optionnel. | aucun |

Le bridge partage l'espace réseau du serveur de jeu. Redémarrer le serveur de
jeu redémarre aussi le bridge. Cela coûte quelques secondes, pas de
progression : le bridge écrit chaque check sur le disque.

## Où la pile range les choses

| Volume | Contenu | Le supprimer pour |
| --- | --- | --- |
| `tf2-archipelago_tf2game` | Les 14 Go de fichiers de jeu, SourceMod et le plugin | Tout retélécharger |
| `tf2-archipelago_bridgestate` | Les checks et les déblocages de la partie | Rien d'utile. Le bridge reconstruit les checks depuis la room. |
| `tf2-archipelago_apoutput` | La session, avec `COMPOSE_PROFILES=selfhost` seulement | [Démarrer une nouvelle partie](../operate/start-a-new-run.md) |
| `tf2-archipelago_tailscale_fastdl_state` | L'identité Tailscale du nœud Funnel | Se reconnecter à Tailscale |

Les sessions sont des fichiers dans `seed/`. Git ignore ce dossier, et rien ne
le supprime à votre place.

Suite : [Créer la session](create-the-session.md).
