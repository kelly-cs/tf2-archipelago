# Les bots de votre équipe

Team Fortress 2 équilibre chaque vague de Mann vs Machine pour six joueurs dans
l'équipe RED. À deux, les robots passent. La vague 1 d'une mission Advanced ne
se gagne pas à deux joueurs, et la partie s'arrête là.

Le serveur remplit les places vides pour vous. Rien à installer, rien à taper.

## Ce qu'ils font

Le serveur remplit RED jusqu'à six joueurs au début de la vague, et la garde
pleine ensuite. Un bot qui meurt revient en une seconde. Les bots choisissent
leur classe, se battent, et dépensent leurs crédits à la station entre les
vagues. Ils se déclarent prêts aussi, donc la vague démarre quand *vous*
appuyez sur F4.

Ils ne sont pas humains. Les Engineers construisent trop près des robots, les
Spies se font repérer tard, et un bot ne fera jamais le coup malin de votre
ami. Ils sont assez bons pour rendre une vague gagnable, et c'est leur rôle.

Un bot cède sa place quand un ami arrive. RED tient six joueurs. Quand
l'équipe est pleine de bots et qu'un joueur se connecte, un bot part. Le
joueur prend le siège. Le mod remplit l'équipe à nouveau au début de la vague
suivante.

Les bots portent les noms de bots du jeu, ceux d'un serveur Valve.

## Les baisser, ou les couper

Les réglages dans `.env` :

| Variable | Défaut | Effet |
| --- | --- | --- |
| `SRCDS_BOTS` | `1` | `0` les garde hors du terrain jusqu'à un `sm_addbots` d'un admin |
| `SRCDS_BOT_TEAM_SIZE` | `6` | Le nombre de joueurs dans RED, humains compris |
| `SRCDS_BOT_CLASS_BLACKLIST` | vide | Les classes que les bots ne jouent jamais, séparées par des virgules : `sniper,spy` |
| `SRCDS_BOT_TEAM_COMP` | vide | Les classes dont les bots remplissent RED, dans l'ordre. Voir ci-dessous. |
| `TF2AP_BOT_UPGRADES_CHAT` | `0` | `1` écrit dans le chat ce que les bots achètent à la station d'améliorations |

Baissez `SRCDS_BOT_TEAM_SIZE` pour une partie plus dure : à `4`, trois amis
reçoivent un bot. Mettez `SRCDS_BOTS=0` quand vous êtes six et que les places
vous reviennent.

Changez l'un de ces réglages sur des relevés, pas sur un souvenir. Valve règle
chaque vague pour six défenseurs, et les bots existent pour qu'une équipe plus
petite gagne. Personne n'a mesuré à quel point. `wave_failures` dans
`/healthz` nomme chaque vague perdue d'une soirée, les pires d'abord, et
`tf2ap_wave_lost_total` trace la même chose. Jouez une mission, lisez quelle
vague vous a arrêtés, puis changez un chiffre. Voir
[Dépannage](../operate/troubleshooting.md).

Les bots sont de mauvais Snipers et de mauvais Spies.
`SRCDS_BOT_CLASS_BLACKLIST=sniper,spy` les garde sur les classes qu'ils jouent
bien. Les noms de classe sont ceux du mod : `scout`, `soldier`, `pyro`,
`demoman`, `heavyweapons`, `engineer`, `medic`, `sniper`, `spy`.

Une liste noire interdit des classes. Elle ne dit pas ce qu'est l'équipe. Un
tirage dans le reste a donné trois Spies et deux Scouts sur une mission
Advanced. Une autre équipe n'avait pas d'Engineer et a perdu deux fois la
vague 1 de Quarry.

`SRCDS_BOT_TEAM_COMP=engineer,medic,heavyweapons,soldier,demoman` nomme
l'équipe à la place. L'ordre est celui dans lequel les places se remplissent.
Mettez donc en premier les classes dont vous ne pouvez pas vous passer. Les
humains prennent les places avant les bots, et les dernières entrées servent
rarement. Les noms de classe sont ceux du mod, comme pour la liste noire.

Une équipe nommée ici l'emporte sur la liste noire. Une liste plus courte que
les places libres laisse le reste au mod.

Le jeu ne permet plus d'inspecter les améliorations d'un coéquipier. Avec
`TF2AP_BOT_UPGRADES_CHAT=1`, le chat dit ce que chaque bot achète, une ligne
par achat. C'est désactivé par défaut parce qu'un bot achète beaucoup.

Tous prennent effet au chargement de la carte suivante. `make restart` est la
façon sûre.

Sur Windows, le lanceur a un onglet **Bots** pour les mêmes réglages. Six
menus, un par place, nomment l'équipe dans l'ordre. Un équipement prédéfini
par classe dit avec quelles armes un bot de cette classe apparaît. Les armes
de base sont le défaut. Voir
[Installer sur Windows](../setup/install-windows.md).

Un bot tient ses distances selon ce qu'il porte, pas selon sa classe. Un Brass
Beast se rapproche, parce qu'il ne peut plus se replacer une fois lancé ; un
Tomislav tient une ligne. Un fusil à pompe avance au lieu de tirer à portée de
minigun.

Un bot sort aussi une arme qui a encore des munitions, au lieu de marcher sur
un robot avec une arme vide. C'est ce que faisait un Heavy quand son minigun
était à sec.

## Qui les a écrits

[OfficerSpy/TF2-MvM-Defender-TFBots][mod], GPL-3.0, et cinq dépendances :
CBaseNPC, Actions, TF2Attributes, TF Econ Data et TF2Utils. Le serveur les
compile depuis la source. TF2Attributes reçoit un correctif à nous depuis
`deploy/patches/`, dont le README dit pourquoi.

Le mod lui-même vient de notre fork, [m-this/tf2-mvm-bots][fork]. Sa branche
`tf2ap` est un tag amont plus nos changements, et `DEFENDERBOTS_VERSION` nomme
un tag de cette branche.

Le comportement des bots est celui du mod. Un bot qui rentre dans un mur se
signale au dépôt d'OfficerSpy, pas à celui-ci. La liste noire de classes et le
fichier d'équipement du serveur sont à nous, sur le fork.

[mod]: https://github.com/OfficerSpy/TF2-MvM-Defender-TFBots
[fork]: https://github.com/m-this/tf2-mvm-bots

## Sur un serveur qui n'est pas cette image

Chaque version publie `tf2-defender-bots.zip`. Il porte tout l'ensemble :
plugins, extensions pour Linux et Windows, gamedata et les repères de
navigation par carte. Le zip part de `addons/`, donc une seule décompression
dans le dossier du jeu (`tf/`) suffit.

Réglez ensuite les trois convars dans `server.cfg` :

```
sm_redbots_manager_mode 2
sm_redbots_manager_defender_team_size 6
sm_redbots_manager_min_players -1
```

`mode 2` fait apparaître les bots au début de la vague. `min_players -1` compte
plus qu'il n'y paraît : la barrière du mod compte RED *avant* la vague, où un
joueur seul n'a pas encore de bots. Laissez-la active, et elle bloque le F4 qui
les fait apparaître.
