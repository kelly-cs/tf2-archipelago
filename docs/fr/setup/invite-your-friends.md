# Inviter vos amis

Vos amis ont besoin de trois choses :

1. Une ligne de connexion.
2. Le mot de passe du serveur, si vous en avez mis un.
3. Une phrase : les classes et les armes commencent verrouillées, réussir des
   vagues les débloque, et tout le monde partage.

Ils n'installent rien. Un client Team Fortress 2 normal suffit.

## Choisir qui atteint le serveur

Page du lanceur : **Settings**, puis **Networking**, ligne **Who can reach
it**. Dans `.env`, c'est `SRCDS_REACH`.

| Choix | `.env` | Qui peut rejoindre | Jeton de connexion | Port sur la box |
| --- | --- | --- | --- | --- |
| **Local network** | `lan` | Les gens sur le même réseau. La valeur par défaut. | Non | Non |
| **Forwarded port** | `port` | Toute personne qui a votre adresse publique. | Oui | Oui, UDP et TCP |
| **Steam relay** | `steam` | Toute personne qui a l'adresse du relais. | Oui | Non |

**Steam relay n'est pas terminé.** Aucune partie ne l'a mené jusqu'à un client
qui rejoint. Utilisez **Forwarded port** pour des amis par internet.

Un serveur sans jeton de connexion reste sur le réseau local, quel que soit
votre choix. Le journal le dit.

## Le jeton de connexion

**Forwarded port** et **Steam relay** connectent le serveur à Steam. Il faut
pour cela un Game Server Login Token.

1. Ouvrez [steamcommunity.com/dev/managegameservers](https://steamcommunity.com/dev/managegameservers).
2. Créez un jeton pour l'App ID `440`.
3. Collez-le dans **Login token** sur la page **Networking**, ou dans
   `SRCDS_TOKEN` dans `.env`.

Le jeton n'est pas un mot de passe que quelqu'un tape. Il identifie le serveur
auprès de Steam. Sans jeton, le serveur n'obtient jamais de session Steam et
refuse chaque joueur sans message utile.

## Réseau local

Rien à régler. La ligne **Join** montre l'adresse. Vos amis la tapent dans la
console de développement :

```text
connect 192.168.1.20:27015
```

Le jeu refuse les joueurs d'un autre réseau avec `LAN servers are restricted
to local clients (class C)`. Un Wi-Fi invité, un VPN et un serveur dans un
conteneur comptent comme un autre réseau.

## Port redirigé

1. Sur votre box, redirigez le port du jeu vers cette machine, en **UDP et
   TCP**. Le port par défaut est `27015`.
2. Ouvrez le même port dans le pare-feu de la machine.
3. Réglez **Join address** sur la page **Game server** avec votre adresse
   publique.
4. Donnez la ligne de connexion :

```text
connect votre.adresse.publique:27015
```

L'UDP porte le jeu. Une règle qui ne redirige que le TCP répond au navigateur
de serveurs et rejette chaque connexion.

## Relais Steam

Le serveur demande une adresse de relais à Valve et l'affiche dans le
journal :

```text
FakeIP allocation succeeded: 169.254.13.42:20232, 20233
```

Le lanceur la montre sur la ligne **Join**. Vos amis se connectent à la
première adresse. Vous ne redirigez rien, et votre propre adresse reste
cachée.

L'adresse change à chaque démarrage. Envoyez celle de cette partie.

## Le mot de passe du serveur

Réglez **Server password** sur la page **Game server**, ou `SRCDS_PW` dans
`.env`. Vide laisse entrer toute personne qui a l'adresse.

Vos amis tapent ceci avant de se connecter :

```text
password entre-amis
connect ...
```

Ne le confondez pas avec le mot de passe de la console, `SRCDS_RCONPW`.
Celui-là lance des commandes d'administration. Ne le donnez jamais.

## Rester hors de la liste publique

Un serveur avec un jeton de connexion peut apparaître dans le navigateur de
serveurs public. Mettez un mot de passe de serveur, et les inconnus qui le
trouvent ne peuvent pas entrer.

## La console de développement

La console est désactivée par défaut dans Team Fortress 2. Pour l'activer :

1. Ouvrez **Options**, puis **Clavier**, puis **Avancé**.
2. Cochez **Activer la console de développement**.

La touche qui l'ouvre est `` ` `` sur un clavier américain, et souvent `²` sur
un clavier français.

## Quand un joueur ne peut pas se connecter

Vérifiez dans cet ordre :

1. **Le jeton.** Avec **Forwarded port** ou **Steam relay** et sans jeton, le
   serveur refuse tout le monde. La console dit
   `Could not establish connection to Steam servers`.
2. **Le réseau.** Avec **Local network**, le joueur doit être sur le même
   réseau. Voir [Réseau local](#réseau-local).
3. **Le port.** Avec **Forwarded port**, la box doit rediriger le port en UDP
   et en TCP.
4. **L'adresse.** Avec **Steam relay**, l'adresse est celle de cette partie.
5. **La machine.** Un portable qui dort est un serveur éteint.

Suite : [La première session](../play/first-session.md).
