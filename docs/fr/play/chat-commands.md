# Commandes de chat

Tout ici se tape dans le chat normal de Team Fortress 2. Il n'y a aucun
client à installer, aucune seconde fenêtre à garder ouverte.

## Pour les joueurs

| Tapez ceci | Ce que le serveur fait |
| --- | --- |
| `!ap` | Affiche l'aide |
| `!ap status` | Affiche l'état de la partie. La mission et la vague. Les classes et les emplacements que la partie possède. Combien de missions la partie a débloquées et terminées. Si le bridge est connecté au serveur randomizer |
| `!mission` | Liste les missions de la partie, et marque laquelle est jouée, lesquelles sont terminées et lesquelles sont encore verrouillées |
| `!ap missing` | Liste les checks que personne n'a encore trouvées |
| `!ap checked` | Liste les checks déjà trouvées |
| `!ap remaining` | Liste ce qui reste. Le serveur randomizer décide s'il répond avant la fin de la partie. |
| `!ap players` | Liste les participants de la session |
| `!ap hint Class: Scout` | Demande où se trouve un item |
| `!ap hint_location Doe's Doom Wave 3` | Demande ce que contient un endroit |
| `!ap options` | Affiche les réglages de la session |
| `!ap help` | Affiche l'aide du serveur randomizer |
| `!ap license` | Affiche la licence du serveur randomizer |
| `!apchat nice one` | Parle aux autres joueurs de la session |

Le serveur répond lui-même à `!ap status` et à `!mission`, donc ces deux
commandes fonctionnent quand le serveur randomizer ne répond pas. Toute
autre commande `!ap` part au serveur randomizer, qui affiche sa réponse
dans le chat. `!apchat` envoie du texte brut. Les lignes des autres
joueurs arrivent dans le même chat. Le chat d'équipe fonctionne comme le
chat général.

Demander où se trouve un item coûte des points d'indice, que la session
gagne grâce aux checks. Utilisez les noms d'items de
[Archipelago pour les joueurs MvM](../archipelago-for-mvm-players.md).

Le nom entier, préfixe compris. `!ap hint Scout` ne trouve pas
`Class: Scout` : le serveur randomizer compare sur le nom entier et refuse
tout ce qui ressemble à moins de 75 %, et le préfixe fait plus de la
moitié de ce nom. Il répond avec le nom qu'il pense que vous vouliez dire,
donc le second essai fonctionne.

## Les commandes refusées

La liste ci-dessus est la liste complète. Toute autre commande est
refusée avec :

```
[AP] That multiworld command cannot be sent from the game. It cannot be undone.
```

Chaque commande de la liste ne fait que lire. Celles qui manquent changent
la partie, et rien dans une session randomisée ne les annule. `!release`
donne les items restants de ce serveur aux autres participants. Une seule
ligne tapée par un joueur terminerait la partie pour tout le monde, et un
serveur de jeu n'a ni comptes ni moyen de savoir qui devrait être autorisé.

La règle est donc une liste de ce qui est permis, tenue sur le bridge. Il
y a une seule liste plutôt qu'une par composant.

## Les autres refus

| Le chat dit | Pourquoi |
| --- | --- |
| `Wait a moment before speaking to the multiworld again.` | Un joueur peut parler une fois toutes les trois secondes |
| `Too much is going to the multiworld. Wait a moment.` | Cinq lignes d'un coup pour tout le serveur, puis une toutes les trois secondes |
| `That line is too long for the multiworld.` | Une ligne fait au plus 300 caractères |
| `The bridge has no connection to the multiworld. It refused your line.` | Le serveur randomizer est injoignable en ce moment |

Une ligne n'est jamais mise en attente. Un message qui arrive dix minutes
en retard est pire qu'un message refusé pendant que le joueur lisait
encore le chat.

## Pour l'admin

Un admin est un identifiant Steam dans `SRCDS_ADMIN_STEAMIDS`. Réglez-le
avant le premier démarrage et il fonctionne dans le chat normal, comme
n'importe quelle autre commande :

| Tapez ceci | Ce que le serveur fait |
| --- | --- |
| `!mission 3` | Passe à la troisième mission de la liste |
| `!mission mvm_decoy_intermediate` | Passe par le nom du fichier de mission |

Un joueur qui n'est pas admin reçoit un refus plutôt qu'un silence. Une
mission que la partie n'a pas débloquée est refusée à tout le monde : son
ticket est quelque part dans le multiworld.

Changer de mission change la mission, et la carte avec elle quand les deux
diffèrent.

## Quelle mission se joue

La partie décide, pas la rotation des cartes :

- Le serveur démarre sur `tf2ap_start_mission`, ou sur la mission propre à
  la carte si la valeur est vide.
- Si la mission chargée ne fait pas partie de la partie, le serveur passe à
  la première mission débloquée qui n'est pas terminée. Il fait de même si la
  partie n'a pas débloqué la mission chargée.
- Quand l'équipe termine une mission, le serveur annonce la mission
  suivante et la charge `tf2ap_next_mission_delay` secondes plus tard.
  C'est la première mission débloquée et non terminée, dans l'ordre du
  tirage. Quand toutes les missions débloquées sont terminées, le serveur
  en rejoue une jusqu'à ce qu'un ticket en ouvre une autre.

## Pour l'hébergeur

Les mêmes commandes, et quelques autres, se lancent depuis la console
distante. Connectez-vous au serveur, ouvrez la console développeur, et
tapez :

```
rcon_password votre-SRCDS_RCONPW
rcon sm_ap_status
```

| Commande | Ce qu'elle fait |
| --- | --- |
| `sm_ap_status` | Affiche la mission, la vague, quels événements de jeu existent, les déblocages, les missions, l'état du relais de chat, la profondeur de la file et la dernière erreur |
| `sm_ap_mission` | Liste les missions de la partie. Avec un argument, change de mission |
| `sm_ap_resync` | Redemande l'ensemble des déblocages au bridge |
| `sm_ap_report wave_cleared 3` | Rapporte une vague réussie à la main |
| `sm_ap_report mission_cleared` | Rapporte une mission réussie à la main |

La console n'est pas un joueur, elle n'est donc pas admin non plus : rcon
tourne comme le serveur lui-même et atteint chaque commande ci-dessus quoi
que dise `SRCDS_ADMIN_STEAMIDS`.

`sm_ap_report` sans numéro de vague utilise la vague sur laquelle le jeu
se trouve. Rapporter la même check deux fois n'est pas un problème : le
bridge identifie une check par l'endroit auquel elle appartient, donc le
second rapport ne fait rien.

## Les réglages

Ce sont des variables de console. Réglez-en une pour la session avec
`rcon tf2ap_debug 1`. Pour la garder entre les redémarrages, modifiez
`cfg/sourcemod/tf2_archipelago.cfg` dans les fichiers du jeu. La stack
n'écrase jamais ce fichier une fois qu'il existe.

| Variable | Défaut | Ce qu'elle fait |
| --- | --- | --- |
| `tf2ap_announce` | `1` | Écrit les vagues réussies et les items reçus dans le chat |
| `tf2ap_chat` | `1` | Écrit ce que dit le reste de la session dans le chat |
| `tf2ap_debug` | `0` | Écrit chaque appel au bridge et chaque événement de jeu dans le chat et la console |
| `tf2ap_bridge_url` | `http://127.0.0.1:24680` | Où se trouve le bridge. Loopback uniquement. Ne le changez pas. |
| `tf2ap_start_mission` | vide | La mission de départ du serveur, par nom de popfile. Le lanceur et l'image l'écrivent depuis leurs réglages. |
| `tf2ap_next_mission_delay` | `30` | Secondes entre la fin d'une mission et la suivante. `0` laisse faire la rotation du jeu. |
| `tf2ap_bot_upgrades_chat` | `0` | Écrit dans le chat ce que les bots défenseurs achètent à la station d'améliorations. Une ligne par achat. |

Les erreurs atteignent le chat quel que soit le réglage de
`tf2ap_announce`. Un échec que personne ne voit se retrouve reproché au
jeu.
