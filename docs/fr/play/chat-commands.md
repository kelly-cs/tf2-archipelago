# Commandes de chat

Vous tapez tout ce qui est sur cette page dans le chat normal de Team
Fortress 2. Il n'y a pas de client à installer ni de deuxième fenêtre à garder
ouverte.

## Pour les joueurs

| Tapez ceci | Ce que le serveur fait |
| --- | --- |
| `!ap` | Afficher l'aide |
| `!ap status` | Afficher la mission, la vague, les classes et les emplacements débloqués, et si le bridge est connecté |
| `!mission` | Lister les missions de la partie : celle qui se joue, les réussies, les verrouillées |
| `!ap missing` | Lister les checks que personne n'a encore trouvés |
| `!ap checked` | Lister les checks déjà trouvés |
| `!ap remaining` | Lister ce qui reste, si la room l'autorise avant la fin |
| `!ap players` | Lister les joueurs du multiworld |
| `!ap hint Class: Scout` | Demander où est un objet |
| `!ap hint_location Doe's Doom Wave 3` | Demander ce qu'un check contient |
| `!ap options` | Afficher les options de la session |
| `!ap help` | Afficher l'aide de la room |
| `!apchat bien joué` | Parler aux autres joueurs du multiworld |
| `!ap_buffs` | Afficher les bonus de votre équipement |
| `!ap unlock mission` | Test mode seulement. Donner le ticket de mission suivant. |

Le serveur de jeu répond lui-même à `!ap status` et `!mission`, donc ces
commandes marchent quand la room ne répond pas. Toute autre commande `!ap`
part à la room, qui répond dans le chat. Le chat d'équipe marche comme le chat
général.

Les indices coûtent des points d'indice, que la session gagne avec les
checks. Tapez le nom entier de l'objet, avec son préfixe. `!ap hint Scout` ne
trouve pas `Class: Scout`. La room répond avec le nom qu'elle pense que vous
vouliez, donc le deuxième essai marche.

## Les commandes refusées

La liste ci-dessus est la liste entière. Le serveur refuse toute autre
commande :

```text
[AP] That multiworld command cannot be sent from the game. It cannot be undone.
```

Chaque commande autorisée ne fait que lire. Les commandes absentes changent la
partie, et rien ne les annule. `!release`, par exemple, donne chaque objet
restant de ce serveur aux autres joueurs. Une ligne d'un seul joueur termine
la partie pour tout le monde.

## Les autres refus

| Le chat dit | Pourquoi |
| --- | --- |
| `Wait a moment before speaking to the multiworld again.` | Un joueur peut parler une fois toutes les trois secondes |
| `Too much is going to the multiworld. Wait a moment.` | Cinq lignes d'un coup pour tout le serveur, puis une toutes les trois secondes |
| `That line is too long for the multiworld.` | Une ligne fait au plus 300 caractères |
| `The bridge has no connection to the multiworld. It refused your line.` | La room est injoignable pour l'instant |

Une ligne refusée n'est jamais mise en attente.

## Pour l'admin

Un admin est un identifiant Steam dans **Admins by Steam id**, sur la page de
réglages **Game server**. Dans `.env`, c'est `SRCDS_ADMIN_STEAMIDS`. Les deux
formes marchent : l'identifiant à 17 chiffres de l'URL du profil, ou
`STEAM_0:1:...`.

| Tapez ceci | Ce que le serveur fait |
| --- | --- |
| `!mission 3` | Passer à la troisième mission de la liste |
| `!mission mvm_decoy_intermediate` | Passer à une mission par son nom de fichier |
| `!ap bots` | Ouvrir l'équipe de bots en menu. Choisissez une place, puis une classe. |

Le serveur dit non à un joueur qui n'est pas admin. Il refuse à tout le monde
une mission que la partie n'a pas débloquée.

Un changement de bots fait pendant une vague s'applique à la pause suivante,
parce qu'un bot retiré pendant une vague perd ses constructions.

## Pour l'hébergeur

Les mêmes commandes, et quelques autres, se lancent depuis la console
distante. Dans le lanceur, tapez-les dans le champ rcon sous le journal de
l'onglet **Play**. Avec Docker, utilisez `make rcon`.

| Commande | Ce qu'elle fait |
| --- | --- |
| `sm_ap_status` | Afficher la mission, la vague, les événements de jeu que le serveur envoie, les déblocages, les missions et la dernière erreur |
| `sm_ap_mission` | Lister les missions de la partie. Avec un argument, passer à l'une d'elles |
| `sm_ap_resync` | Redemander le jeu de déblocages au bridge |
| `sm_ap_buffs` | Afficher les bonus de l'équipement actuel |

La console est le serveur lui-même, donc elle atteint chaque commande, quoi
que dise **Admins by Steam id**.

### Commandes de débogage et de test

Ces commandes demandent l'accès admin racine. Les bonus de test durent
jusqu'au rechargement du plugin, de la carte ou de l'état de la partie.

| Commande | Ce qu'elle fait |
| --- | --- |
| `sm_ap_buff_test <1-80\|clé-effet\|all> [niveaux]` | Ajouter des effets à votre arme active. Exemple : `sm_ap_buff_test projectile-count 3` |
| `sm_ap_buff_give <cible> <1-80\|clé-effet\|all> [niveaux]` | Ajouter des effets à l'arme active d'un autre joueur de RED |
| `sm_ap_buff_slot <cible> <1\|2\|3\|primary\|secondary\|melee> <1-80\|clé-effet\|all> [niveaux]` | Ajouter des effets à un objet équipé, par emplacement |
| `sm_ap_projectile_debug on` | Activer le diagnostic des projectiles |
| `sm_ap_projectile_debug` | Afficher les 24 dernières lignes de diagnostic |
| `sm_ap_projectile_debug off` | Désactiver le diagnostic des projectiles |
| `sm_ap_unlock_override <on\|off>` | Autoriser toutes les classes et tous les emplacements pour l'instant, ou remettre les verrous de la partie |
| `sm_ap_bundle [crédits]` | Payer un Cash Bundle de test, 200 crédits par défaut |
| `sm_ap_report wave_cleared [vague]` | Signaler une vague réussie à la main |
| `sm_ap_report mission_cleared` | Signaler une mission réussie à la main |
| `sm_ap_report death` | Signaler une vague perdue à la main |

Dans le chat, enlevez le préfixe `sm_` : `!ap_buff_test projectile-count 3`.
Depuis la console du jeu, envoyez la commande au serveur avec
`cmd sm_ap_buff_test projectile-count 3`.

`sm_ap_report` sans numéro de vague prend la vague en cours. Signaler deux
fois le même check ne fait rien : le bridge identifie un check par sa place.

## Les variables de console

Réglez-en une pour la session avec `tf2ap_debug 2` dans le champ de commande.
Pour la garder entre les redémarrages, modifiez
`cfg/sourcemod/tf2_archipelago.cfg` dans les fichiers de jeu.

| Variable | Défaut | Ce qu'elle fait |
| --- | --- | --- |
| `tf2ap_announce` | `1` | Écrire les vagues réussies et les objets reçus dans le chat |
| `tf2ap_chat` | `1` | Écrire ce que dit le reste du multiworld dans le chat |
| `tf2ap_debug` | `1` | `0` n'écrit rien. `1` écrit chaque appel au bridge et chaque événement de jeu dans la console et le journal SourceMod. `2` les écrit aussi dans le chat. |
| `tf2ap_bridge_url` | `http://127.0.0.1:24680` | Où est le bridge. Ne le changez pas. |
| `tf2ap_start_mission` | vide | La mission où le serveur démarre. Le lanceur l'écrit depuis **Start mission**. |
| `tf2ap_next_mission_delay` | `30` | Les secondes entre une mission réussie et la suivante. `0` laisse faire la rotation du jeu. |
| `tf2ap_bot_upgrades_chat` | `0` | Écrire dans le chat ce que les bots achètent à la station d'amélioration |
| `tf2ap_bots_wait_for_players` | `1` | Garder les bots non prêts tant que chaque joueur de RED n'est pas prêt |
| `tf2ap_bots_backfill` | `1` | Remettre un bot sur RED quand un joueur part entre deux vagues |

Les erreurs arrivent dans le chat quoi que dise `tf2ap_announce`.

Suite : [Les bots de votre équipe](defender-bots.md).
