# Les bots de votre équipe

Team Fortress 2 calibre chaque vague de Mann vs Machine pour six joueurs sur
RED. À deux, les robots passent. Le serveur remplit donc les places vides
avec des bots. Rien à installer, rien à taper.

## Ce qu'ils font

- Ils rejoignent RED quand une vague commence et restent jusqu'à la fin.
- Ils choisissent leurs classes, se battent, et achètent leurs propres
  améliorations entre les vagues.
- Ils se mettent prêts eux-mêmes, donc une vague commence quand **vous**
  appuyez sur F4.
- Un bot qui meurt revient en moins d'une seconde.
- Quand un ami rejoint une équipe pleine, un bot part et l'ami prend la place.

Ils ne sont pas humains. Ils repèrent les espions tard, et ils ne font jamais
le coup malin de votre ami. Ils rendent une vague gagnable, et c'est leur
rôle.

## Les réglages

Page du lanceur : **Settings**, puis **Bots**.

| Lanceur | `.env` | Défaut | Ce que le réglage fait |
| --- | --- | --- | --- |
| Fill RED with bots | `SRCDS_BOTS` | activé | Désactivé garde les bots hors du terrain jusqu'à ce qu'un admin lance `!addbots`. |
| Fill RED to | `SRCDS_BOT_TEAM_SIZE` | `6` | Jusqu'à combien de joueurs les bots remplissent RED, humains compris. |
| Un menu par place | `SRCDS_BOT_TEAM_COMP` | `engineer,medic,heavyweapons,soldier,demoman` | Les classes que les bots jouent, dans l'ordre où les places se remplissent. |
| Classes, une case par classe | `SRCDS_BOT_CLASS_BLACKLIST` | vide | Les classes où le mod pioche quand une place n'est pas nommée. Décochez celles que les bots ne jouent jamais. |
| Un équipement par classe | `SRCDS_BOT_LOADOUTS` | vide | Ce qu'un bot de chaque classe porte. Vide donne les armes de base. |
| Say what they buy | `TF2AP_BOT_UPGRADES_CHAT` | désactivé | Écrire dans le chat chaque amélioration qu'un bot achète. |
| Cosmetic items | `SRCDS_BOT_HATS` | activé | Un chapeau au hasard sur chaque bot. |
| Unusual effects | `SRCDS_BOT_HAT_EFFECTS` | désactivé | Un effet unusual au hasard sur ce chapeau. |

Chaque réglage de cette page s'applique au prochain chargement de carte.
**Restart** est le moyen sûr.

### Moins de bots, ou aucun

- Baissez **Fill RED to** pour une partie plus dure. À `4`, trois amis ont un
  bot.
- Désactivez **Fill RED with bots** quand vous jouez à six.

Changez ces réglages d'après un relevé, pas un souvenir. `wave_failures` sur
la page de santé du bridge nomme chaque vague perdue par l'équipe, la pire en
premier. Voir [Dépannage](../operate/troubleshooting.md#interroger-le-bridge).

### L'équipe

Les humains prennent les places avant les bots. Mettez donc en premier les
classes dont vous ne pouvez pas vous passer. Une équipe plus courte que les
places vides laisse le reste au mod, qui pioche dans les classes autorisées.

Les bots sont de mauvais snipers et de mauvais espions. Les noms de classe
sont ceux du mod : `scout`, `soldier`, `pyro`, `demoman`, `heavyweapons`,
`engineer`, `medic`, `sniper`, `spy`.

### Les équipements

Les préréglages couvrent les équipements courants. Pour créer le vôtre :

1. Ouvrez **Settings**, puis **Loadouts**.
2. Choisissez une classe, puis une arme par emplacement.
3. Tapez un nom et appuyez sur **Save**.

L'équipement apparaît ensuite dans les menus d'armes de cette classe sur la
page **Bots**. Supprimez un équipement, et toute place qui le nommait joue
avec les armes de base.

### L'apparence

Un bot tire un chapeau que sa classe peut porter et le garde pour la mission.
Le chapeau distingue donc un Heavy d'un autre. Les effets unusual sont
désactivés par défaut, parce que six effets de particules à l'écran pendant
toute une vague, c'est beaucoup. Ni l'un ni l'autre ne change la façon dont un
bot joue.

## Changer l'équipe en cours de mission

1. Ouvrez l'onglet **Bots** du lanceur.
2. Réglez les places.
3. Appuyez sur **Apply to the running server**.

Le mod ne remplace que les places dont la classe a changé. La vague continue,
et les bots gardent l'argent gagné. Dans le jeu, `!ap bots` ouvre la même
équipe en menu pour un admin.

Un changement fait pendant une vague s'applique à la pause suivante. Un bot
retiré pendant une vague perd ses constructions.

## Une équipe réduite

Valve règle chaque vague pour six défenseurs. Deux réglages aident une équipe
qui en manque :

- **Weapon buffs**, sur la page **Rewards**, rendent l'équipe plus forte. Ils
  sont activés par défaut. Voir
  [Les options de la partie](../setup/shape-of-the-run.md#récompenses).
- **Robot health (%)**, sur la page **Balancing**, met à l'échelle la santé
  de chaque robot. À 50, un test a réussi trois vagues sur huit là où la même
  équipe n'en avait réussi aucune.

## Comment ils jouent

- Un bot garde ses distances selon ce qu'il porte. Un Brass Beast s'approche,
  un Tomislav tient un couloir.
- Un bot passe à une arme qui a encore des munitions plutôt que de marcher
  vers un robot avec une arme vide.
- Un Engineer s'installe près de la trappe, pas à la porte d'apparition des
  robots. `sm_redbots_manager_engineer_nest_depth` dit jusqu'où sur le chemin
  de la bombe il peut construire, en fraction du chemin. La valeur par défaut
  est `0.4`.
- À la station d'amélioration, un bot achète d'abord des dégâts, pour l'arme
  qu'il a en main. Un Medic achète du soin, un Engineer achète la sentry. Les
  résistances viennent en dernier.

## Qui les a écrits

Les bots sont les [MvM Defender TFBots d'OfficerSpy][mod], en GPL-3.0, plus
cinq dépendances : CBaseNPC, Actions, TF2Attributes, TF Econ Data et TF2Utils.
Le mod vient de [m-this/tf2-mvm-bots-go][fork]. Ce dépôt écrit les décisions
des bots en Go et génère le SourcePawn à partir de là.

Signalez un bot qui marche dans un mur à ce dépôt, pas à celui-ci.

[mod]: https://github.com/OfficerSpy/TF2-MvM-Defender-TFBots
[fork]: https://github.com/m-this/tf2-mvm-bots-go

## Sur un serveur qui n'est pas celui-ci

Chaque version joint `tf2-defender-bots.zip`. Il contient les plugins, les
extensions pour Linux et Windows, les gamedata et les indices de navigation.
Le zip part de `addons/`, donc une seule décompression dans `tf/` l'installe.

Réglez ensuite ceci dans `server.cfg`. Le lanceur et l'image Docker le font
pour vous :

```text
sm_redbots_manager_mode 2
sm_redbots_manager_defender_team_size 6
sm_redbots_manager_min_players -1
```

`mode 2` fait apparaître les bots quand une vague commence. `min_players -1`
compte. La barrière de mise en prêt du mod vaut 3 par défaut. Elle compte RED
avant la vague, où un joueur seul n'a pas encore de bots. Laissée en place,
elle bloque le F4 qui les fait apparaître.

Suite : [Dépannage](../operate/troubleshooting.md).
