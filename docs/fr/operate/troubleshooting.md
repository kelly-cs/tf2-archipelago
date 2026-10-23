# Dépannage

Trois choses peuvent aller mal :

1. Le plugin ne voit pas le jeu.
2. Le plugin n'atteint pas le bridge.
3. Le bridge n'atteint pas la room.

Cette page trouve laquelle. Si vous demandez de l'aide, envoyez d'abord le
paquet de débogage : **Settings**, puis **Debug logs**, dans le lanceur. Il
contient le journal du lanceur, la console du serveur et vos réglages, sans
mot de passe.

## Lire les journaux

- **Lanceur :** le journal en bas de l'onglet **Play**. **Filter the log** le filtre.
- **Docker :** `make logs` suit chaque service. Pour un seul service :

```sh
docker compose --project-directory . \
  --env-file deploy/env/versions.env --env-file .env \
  -f deploy/compose.yml logs -f bridge
```

Remplacez `bridge` par `srcds`, ou par `archipelago` quand la pile héberge la
session elle-même. `make ps` liste les conteneurs.

## Interroger le serveur de jeu

Tapez `sm_ap_status` dans le champ rcon sous le journal, ou
`rcon sm_ap_status` dans la console du jeu. La réponse ressemble à ceci :

```text
[AP] version 0.1.0, mvm yes, mission mvm_decoy, wave 3 of 8
[AP] events: begin_wave yes, wave_complete yes, mission_complete yes
[AP] unlocks held at sequence 6, 0 objective(s) waiting to be sent
[AP] classes: scout, medic
[AP] slots: primary
[AP] Last bridge error: ...
```

Lisez-la dans cet ordre :

- `mvm no` : le plugin ne pense pas que c'est du Mann vs Machine. Rien n'est
  signalé sur une carte qui n'est pas une carte MvM.
- `events: ... no` : votre version du jeu n'envoie pas cet événement.
  Signalez-le. Avec `wave_complete no`, le plugin surveille le compteur de
  vagues à la place.
- `unlocks NOT FETCHED` : le plugin n'a jamais eu de réponse du bridge. Tant
  qu'il n'en a pas, il ne verrouille rien.
- `N objective(s) waiting to be sent` : le bridge ne répond pas. Le plugin
  réessaie toutes les cinq secondes.
- `Last bridge error` : la dernière chose qui a mal tourné.

`sm_ap_resync` redemande le jeu de déblocages au bridge. Essayez-le en
premier quand les déblocages du chat ont l'air périmés.

## Interroger le bridge

Le bridge sert une page avec tout ce qu'il sait. Dans le lanceur, l'onglet
**Play** l'affiche. Avec Docker, la page est sur l'adresse locale, dans
l'espace réseau du serveur de jeu :

```sh
docker run --rm --network container:tf2-archipelago-srcds-1 \
  curlimages/curl:latest -s 127.0.0.1:24680/healthz
```

| Champ | Ce qu'il vous dit |
| --- | --- |
| `connected` | Si la session avec la room est ouverte en ce moment |
| `slot` | Le nom de votre serveur dans la session |
| `missions` | Les missions que la partie a piochées |
| `seed` | L'identité de la session en cours |
| `checks` | Combien de checks la partie contient |
| `items` | Combien d'objets la partie a reçus |
| `acked_seq` | Jusqu'où le plugin a confirmé avoir appliqué |
| `goal_sent` | Si la partie est terminée |
| `last_check` et `last_check_at` | Le dernier check et quand il est arrivé |
| `wave_drift` | Les missions dont le jeu conteste le nombre de vagues |
| `wave_failures` | Chaque vague perdue par l'équipe, la pire en premier |
| `last_error` | La dernière panne côté room |

`last_check` répond à « cette vague a-t-elle compté ». `wave_failures` répond
à « quelles vagues nous ont arrêtés », la question à poser avant de changer
la taille de l'équipe ou les bots.

### Les métriques

Le bridge sert aussi les mêmes nombres en métriques Prometheus sur le port
`24681`, sur l'adresse locale par défaut. `BRIDGE_METRICS_BIND` dans `.env`
l'ouvre à une autre machine.

```sh
curl -s 127.0.0.1:24681/metrics
```

| Métrique | Ce qu'elle vous dit |
| --- | --- |
| `tf2ap_session_connected` | 1 tant que la session avec la room est ouverte |
| `tf2ap_session_missions` | Combien de missions la partie a piochées |
| `tf2ap_run_checks_total`, `tf2ap_run_items_total` | Checks envoyés, objets reçus |
| `tf2ap_run_acked_seq` | Jusqu'où le plugin a confirmé avoir appliqué |
| `tf2ap_run_goal_sent` | 1 une fois la partie terminée |
| `tf2ap_run_last_check_timestamp_seconds` | Quand le dernier check est arrivé |
| `tf2ap_mission_wave_drift` | Une série par mission où le jeu et les tables ne sont pas d'accord. Aucune est le cas sain. |
| `tf2ap_wave_lost_total` | Une série par vague perdue par l'équipe |
| `tf2ap_run_info` | La seed et le slot auxquels les nombres appartiennent |
| `tf2ap_game_up` | 1 quand le serveur de jeu a répondu à une requête lors de cette collecte |
| `tf2ap_game_players`, `_bots`, `_players_human` | Qui est sur le serveur |
| `tf2ap_game_map` | La mission en cours |

Les comptes de joueurs viennent d'une requête au serveur de jeu, gardée dix
secondes. Le serveur de jeu refuse une source qui l'interroge trop souvent,
donc ne baissez pas ce délai.

## Quand la room est hors ligne

Rien n'est perdu. Le bridge écrit chaque check sur le disque avant de
répondre au serveur de jeu, et l'envoie ensuite. Le bridge se reconnecte tout
seul. Les vagues réussies continuent de compter. Les objets reçus arrivent
quand la room revient.

Un bridge qui ne se connecte jamais une seule fois, c'est un autre problème.
Vérifiez l'adresse de la room. Avec Docker, vérifiez aussi `AP_TLS` : une
room sur `archipelago.gg` demande `AP_TLS=true`, une room hébergée dans la
pile demande `AP_TLS=false`.

## Quand le bridge est hors ligne

Le plugin garde ses checks en mémoire et réessaie toutes les cinq secondes.
Le chat dit une fois que le bridge est injoignable.

Dans le lanceur, le bridge tourne à côté du serveur de jeu. **Restart**
ramène les deux. Avec Docker, le bridge redémarre avec le serveur de jeu. Les
checks sont sur le disque, et le jeu de déblocages est reconstruit à partir
d'eux.

Si le fichier d'état du bridge est perdu, la partie ne l'est pas. La room
garde la même liste de checks et l'envoie à chaque connexion. Seul
l'historique des objets est perdu.

## Récupérer un check à la main

Il reste un trou : les secondes entre une vague réussie et la prise du check
par le bridge. La file du plugin est en mémoire. Si le serveur de jeu plante
pendant que le bridge est injoignable, cette file disparaît.

Le plugin écrit chaque check deux fois dans le journal SourceMod. La première
fois quand il le met en file, la deuxième quand le bridge l'a sur le disque.

```text
objective wave_cleared mvm_decoy wave 3 (mission length 8) queued for the bridge
objective wave_cleared mvm_decoy wave 3 is on the bridge's disk
```

Une ligne `queued` sans ligne `on the bridge's disk` correspondante est un
check qui n'est jamais arrivé. Rejouez-le, sur la carte à laquelle il
appartient :

```text
sm_ap_report wave_cleared 3
```

## Docker : ne redémarrez jamais le serveur de jeu seul

Le bridge vit dans l'espace réseau du serveur de jeu. `docker compose up -d
srcds` seul laisse le bridge attaché à un espace qui n'existe plus. Il se dit
sain et n'atteint rien.

Recréez toute la pile :

```sh
make up
```

## Quand les nombres de vagues sont faux

Chaque nombre de vagues vient du wiki, et le jeu fait autorité. Un nombre
faux fait partir la fin de mission une vague trop tôt, ou jamais.

Le bridge compare la longueur que le jeu signale avec sa table et sert les
désaccords sous `wave_drift` :

```json
"wave_drift": [
  {"popfile": "mvm_decoy", "tables": 8, "observed": 7}
]
```

Une mission qui apparaît là est une ligne à corriger dans
`gamedata/missions.go`. Signalez-la. Le check compte quand même.

## Les messages du chat

| Le chat dit | Ce que cela veut dire |
| --- | --- |
| `[AP] The run did not unlock mvm_decoy. Its checks still count.` | Le serveur lance une mission dont la partie n'a pas trouvé le ticket. Un avertissement, pas un refus. |
| `[AP] The bridge speaks API version 2 and this plugin speaks 1.` | Vous avez mis à jour une moitié et pas l'autre. **Repair** dans le lanceur, ou `make build` et `make up` avec Docker. |
| `[AP] The bridge is unreachable.` | Voir [Quand le bridge est hors ligne](#quand-le-bridge-est-hors-ligne). |

## Windows : le lanceur ne démarre pas

- SmartScreen ou Defender l'a bloqué. Voir
  [Installer sur Windows](../setup/install-windows.md#windows-va-vous-avertir).
- Le dossier d'installation est plein. Le serveur de jeu a besoin d'environ
  20 Go libres.
- Autre chose utilise le port du jeu. Changez **Game port** sur la page
  **Game server**.

## Windows : Generate seed échoue

- Le lanceur ne trouve pas l'application Archipelago. Réglez
  **Archipelago app** sur la page **Player options** avec le dossier de
  l'application.
- La version d'Archipelago ne correspond pas. `tf2ap.exe -version` affiche
  celle que le lanceur fixe. Installez celle-là.
- La réserve est trop petite pour les objets. Appuyez sur
  **Check Run Selection** sur la page **Missions**. Ajoutez des missions, ou
  activez les caches de victoire. Voir
  [Les options de la partie](../setup/shape-of-the-run.md#plus-de-checks).
