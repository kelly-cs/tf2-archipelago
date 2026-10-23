# La première session

## Ce qu'un joueur voit

Huit secondes après avoir rejoint, chaque joueur reçoit ceci dans le chat :

```text
[AP] This server runs an Archipelago randomizer.
[AP] The run locks the classes and the weapon slots until it finds them. All players share the unlocks.
[AP] Mission: mvm_decoy. Each wave you clear is a check.
[AP] Unlocked classes: scout, medic
[AP] Unlocked slots: primary
[AP] Type !ap to speak to the multiworld. Examples: !ap hint Class: Scout and !ap missing.
```

Les deux lignes de déblocage montrent l'état de la partie à ce moment. Un
joueur qui arrive tard voit ce que l'équipe a déjà trouvé.

## Pendant une vague

- Les bots remplissent RED jusqu'à six quand la vague commence. Voir
  [Les bots de votre équipe](defender-bots.md).
- Le menu des classes refuse une classe verrouillée, avec une ligne dans le
  chat.
- Les emplacements d'arme verrouillés restent vides à chaque apparition et à
  chaque réapprovisionnement.
- La station d'amélioration affiche les bonus de votre équipement.
  `!ap_buffs` les affiche de nouveau.
- Chaque vague réussie écrit `[AP] Wave 3 cleared.` dans le chat.
- Chaque objet reçu écrit `[AP] Unlocked: Class: Pyro` ou
  `[AP] The run received 200 credits for 4 player(s).`
- Les autres joueurs du multiworld parlent dans le même chat.
- Tout ce qui va mal est écrit en rouge.

## Qui fait quoi

- **L'hébergeur** se connecte comme tout le monde. Il fait aussi tourner le
  lanceur, change de mission et lit les journaux.
- **Les joueurs** réussissent des vagues. Ils n'ont rien à configurer. Leurs
  commandes sont `!ap` et `!apchat`. Voir [Commandes de chat](chat-commands.md).

## Le premier check

La première vague réussie prouve que toute la chaîne marche. Guettez trois
choses, dans cet ordre :

1. `[AP] Wave 1 cleared.` dans le chat du jeu. Le plugin a vu la vague.
2. `check recorded` dans le journal du lanceur. Le check est sur le disque.
3. `tf2 sent <objet> to <quelqu'un>` sur la page de la room. Le multiworld
   l'a reçu. Le [tracker de campagne](tracker.md) montre le check en case
   verte.

Si l'étape 1 n'arrive pas, voir [Dépannage](../operate/troubleshooting.md). Si
l'étape 1 arrive et pas l'étape 2, le chat le dit en rouge.

## Quelle mission se joue

C'est la partie qui décide, pas la rotation des cartes.

- Le serveur démarre sur la **Start mission** des réglages.
- Si la mission chargée ne fait pas partie de la partie, le serveur passe à
  la première mission débloquée et non réussie. Il fait pareil quand la
  partie n'a pas débloqué la mission chargée.
- Quand l'équipe réussit une mission, le serveur charge la mission débloquée
  suivante après 30 secondes. Quand l'équipe a réussi toutes les missions
  débloquées, il en rejoue une jusqu'à ce qu'un ticket en ouvre une autre.
- L'hébergeur change de mission avec **Play** dans le tableau des missions de
  l'onglet **Play**, ou avec `!mission` dans le chat.

## Finir la soirée

Appuyez sur **Stop**, ou `make down` avec Docker. La partie reste sur le
disque. Le démarrage suivant continue la même partie, avec les mêmes checks
et les mêmes déblocages. Une partie peut attendre une semaine.

## Regarder ce que le serveur fait

Le plugin écrit chaque événement du jeu dans la console et dans le journal
SourceMod. Pour voir les mêmes lignes dans le chat du jeu, tapez ceci dans le
champ rcon sous le journal de l'onglet **Play** :

```text
tf2ap_debug 2
```

Avec Docker, utilisez `make rcon`, ou le mot de passe de console de `.env`
dans la console de développement du jeu :

```text
rcon_password votre-mot-de-passe-de-console
rcon tf2ap_debug 2
```

Suite : [Commandes de chat](chat-commands.md).
