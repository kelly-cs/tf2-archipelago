# Mann vs Archipelago

Ce projet transforme un serveur Team Fortress 2 Mann vs Machine en randomizer
[Archipelago](https://archipelago.gg).

- Les classes, les emplacements d'arme et les missions commencent verrouillés.
- Chaque vague que votre équipe réussit est un check. Les checks débloquent
  des choses, chez vous ou dans la partie d'un autre joueur.
- Tout le monde sur le serveur partage les mêmes déblocages.
- Des bots remplissent les places vides de l'équipe RED. Deux personnes
  peuvent donc jouer des vagues que Valve a calibrées pour six.
- Vos amis n'installent rien. Ils rejoignent avec un client Team Fortress 2
  normal.

## Les trois étapes

Chaque partie suit les mêmes trois étapes. Le livre les suit dans l'ordre.

1. **Installer le serveur.** Un seul fichier sur
   [Windows](setup/install-windows.md) ou [Linux](setup/install-linux.md), ou
   une pile [Docker](setup/install.md). Le premier démarrage télécharge
   environ 14 Go de fichiers de jeu.
2. **Créer la session Archipelago.** Choisissez les
   [options de la partie](setup/shape-of-the-run.md), générez une seed,
   envoyez-la sur `archipelago.gg`, puis donnez l'adresse de la room au
   serveur. Voir [Créer la session](setup/create-the-session.md).
3. **Inviter vos amis.** Donnez-leur une ligne de connexion. Voir
   [Inviter vos amis](setup/invite-your-friends.md).

Lisez ensuite [La première session](play/first-session.md) pour savoir à quoi
ressemble la première soirée. [Le lanceur, onglet par onglet](setup/the-launcher.md)
explique chaque écran, et [le tracker de campagne](play/tracker.md) montre la
partie à vos joueurs.

## Démarrage rapide sur Windows

1. Téléchargez `tf2ap.exe` depuis la
   [dernière version](https://github.com/m-this/tf2-archipelago/releases/latest).
2. Lancez-le. Windows affiche un avertissement. Cliquez sur
   **Informations complémentaires**, puis sur **Exécuter quand même**.
   C'est un faux positif. Voir [Installer sur Windows](setup/install-windows.md).
3. Appuyez sur **Start**. Attendez la fin du téléchargement.
4. Installez l'[application Archipelago](https://github.com/ArchipelagoMW/Archipelago/releases).
5. Dans le lanceur, ouvrez **Settings**, puis **Player options**, et appuyez
   sur **Generate seed**.
6. Envoyez le fichier généré sur
   [archipelago.gg/uploads](https://archipelago.gg/uploads) et créez une room.
7. Collez l'adresse de la room dans **Settings**, puis **Archipelago room**,
   et appuyez sur **Restart**.
8. Envoyez à vos amis la ligne de connexion affichée sous les boutons.

## Nouveau sur Archipelago ?

Lisez d'abord [Les mots d'Archipelago, pour les joueurs MvM](archipelago-for-mvm-players.md).
C'est une seule page. Archipelago et Mann vs Machine emploient les mêmes mots
pour des choses différentes, et le reste du livre suppose que vous savez
lesquelles.

## Où trouver de l'aide

- [Dépannage](operate/troubleshooting.md) trouve quelle partie est en panne.
- **Debug logs**, dans les Settings du lanceur, écrit un seul fichier avec
  tout ce dont une personne qui vous aide a besoin. Envoyez ce fichier quand
  vous demandez de l'aide.
- Les problèmes se signalent sur
  [GitHub](https://github.com/m-this/tf2-archipelago/issues).
