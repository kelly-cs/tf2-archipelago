# Prérequis

## Choisir une façon de lancer le serveur

| Façon | Pour qui | Page |
| --- | --- | --- |
| **Lanceur Windows** | La plupart des gens. Un seul exe, rien d'autre. | [Installer sur Windows](install-windows.md) |
| **Lanceur Linux** | Le même programme, sur une machine Linux ou en SSH. | [Installer sur Linux](install-linux.md) |
| **Docker** | Une machine qui fait déjà tourner des piles Docker. | [Installer avec Docker](install.md) |

Les trois font tourner le même logiciel et ont les mêmes réglages.

## La machine

| Quoi | Ce qu'il faut |
| --- | --- |
| Disque | Environ 20 Go libres. Le serveur de jeu fait environ 14 Go et se télécharge une fois. |
| Mémoire | 4 Go. |
| Processeur | Deux cœurs. |
| Réseau | Un accès sortant vers `archipelago.gg`. Rien à ouvrir sur la box, sauf si vous choisissez la route du port redirigé. |

## Ce qu'il faut aussi à l'hébergeur

- L'[application Archipelago](https://github.com/ArchipelagoMW/Archipelago/releases)
  officielle, pour générer la seed. Voir [Créer la session](create-the-session.md).
- Un jeton de connexion de serveur Steam, si des amis rejoignent par
  internet. Voir [Inviter vos amis](invite-your-friends.md). Jouer sur le
  réseau local n'en demande pas.

## Ce dont vous n'avez pas besoin

- Pas de compte Steam pour le serveur.
- Pas de Team Fortress 2 installé sur la machine. Le serveur télécharge ses
  propres fichiers.
- Pas de compte sur `archipelago.gg`.
- Rien pour les joueurs. Un client Team Fortress 2 normal suffit.

## Note de sécurité

Le serveur de jeu est un gros programme en C++ qui lit le trafic réseau de
toute personne qui connaît l'adresse. Faites-le tourner sur une machine où
c'est acceptable.

Suite : [Installer sur Windows](install-windows.md), [Installer sur Linux](install-linux.md)
ou [Installer avec Docker](install.md).
