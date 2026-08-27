# Installer sur Windows

La façon la plus simple de faire tourner un serveur Mann vs Archipelago. Un
seul fichier. Pas de Docker, pas de clone, pas de compilateur.

Téléchargez `tf2ap.exe` depuis la
[dernière version](https://github.com/m-this/tf2-archipelago/releases/latest)
et lancez-le.

## Windows va vous avertir

SmartScreen bloque le premier lancement. Cliquez sur **Informations
complémentaires**, puis sur **Exécuter quand même**. Defender met parfois le
lanceur en quarantaine.

C'est un faux positif. Le lanceur extrait des archives dans votre dossier Team
Fortress 2. Il y écrit les DLL de Metamod et de SourceMod, télécharge un
serveur de jeu et le démarre. Un installeur fait cela. Un virus aussi. Le
lanceur n'a pas encore de signature, donc l'antivirus ne voit pas la
différence.

Le projet a demandé une signature gratuite à SignPath. L'avertissement reste
jusqu'à son arrivée.

Vérifiez le lanceur vous-même :

- Chaque version publie `SHA256SUMS`. Lancez
  `Get-FileHash tf2ap.exe -Algorithm SHA256` et comparez les deux valeurs.
- Chaque version renvoie vers son propre rapport VirusTotal.
- `make launcher` le reconstruit à partir du code du dépôt.
- `gh attestation verify tf2ap.exe --repo m-this/tf2-archipelago` nomme le
  commit et le workflow qui ont produit le fichier que vous avez.

## Ce qui se passe

Une fenêtre s'ouvre et demande l'adresse de votre room Archipelago. Puis
elle installe tout : SteamCMD, le serveur dédié TF2, SourceMod, le plugin, et
les bots qui remplissent votre équipe. Le serveur de jeu pèse environ 14 Go,
donc le premier démarrage prend du temps. Chaque démarrage suivant prend
quelques secondes.

La fenêtre contient :

- **Start**, **Stop**, **Restart**. Un voyant à côté passe au rouge, à
  l'orange puis au vert.
- **Join**, sous les boutons : les adresses où vos amis se connectent.
  **Copy** en met une dans le presse-papiers.
- Un onglet **Log** et une zone **rcon**, pour quand quelque chose semble
  aller de travers.
- Un onglet **Session** : état de la connexion, checks, items, et les
  missions de la partie. **Play this mission** charge celle que vous
  choisissez.
- Un onglet **Bot Switcher** : ce que chaque place de RED joue et ce qu'elle
  porte. **Appliquer au serveur en cours** donne une nouvelle équipe sans
  terminer la mission.
- **Settings**, pour la room, les missions, les bots, qui peut rejoindre et la
  forme de la partie.

Fermer la fenêtre arrête le serveur. Vos réponses sont enregistrées pour la
prochaine fois.

## Ce qu'il vous faut

| Élément | Ce qu'il faut |
| --- | --- |
| Windows | 10 ou 11, 64 bits |
| Disque | Environ 20 Go libres |
| Mémoire | 4 Go pour six joueurs |
| Processeur | Deux cœurs |
| Réseau | Rien, pour des amis sur le même réseau ou via Steam. Un seul port à ouvrir si vous choisissez cette voie. |

Pas de Docker, pas de client Steam, pas de compte Steam pour le serveur
lui-même.

## La session Archipelago

Le lanceur fait tourner le serveur TF2. La session multiworld est séparée.
Mann vs Machine ne fait pas partie des jeux livrés avec Archipelago, donc le
générateur de seed reste dans l'app officielle.

1. Installez l'app officielle
   [Archipelago](https://github.com/ArchipelagoMW/Archipelago/releases). Le
   lanceur la trouve seul aux emplacements habituels.
2. Dans le lanceur, ouvrez **Settings**, réglez les options du joueur, et
   cliquez sur **Generate seed**. Il écrit le fichier joueur et ouvre le
   dossier avec l'archive générée.
3. Envoyez cette archive sur
   [archipelago.gg/uploads](https://archipelago.gg/uploads) pour ouvrir une
   room, et collez l'adresse de la room (par exemple `archipelago.gg:12345`)
   dans le lanceur.

Si le lanceur ne trouve pas l'app Archipelago, **Generate seed** le dit.
Indiquez son dossier dans **Settings → Player options → Archipelago app**.

Voir [Créer la session](create-the-session.md) pour le détail complet, y
compris héberger la session vous-même.

## Inviter des amis

Vos amis se connectent depuis la console du jeu :

```
connect adresse.de.votre.serveur:27015
```

La ligne **Join** sous les boutons montre les adresses à donner. Voir
[Inviter vos amis](invite-your-friends.md) pour toucher des gens hors de
votre réseau.

## Les bots de votre équipe

Valve calibre chaque vague pour six joueurs. Le serveur remplit les places
vides de RED avec des bots qui jouent : ils choisissent une classe,
combattent, et achètent leurs propres améliorations. L'onglet **Bots** les
désactive, réduit l'équipe pour une partie plus dure, ou choisit quelles
classes ils jouent. Voir
[Les bots de votre équipe](../play/defender-bots.md).

## L'essayer sans Archipelago

**Test mode**, dans Settings, fait tourner un multiworld d'une seule
personne sur votre machine. Pas de room, pas de seed, rien ne sort de
l'ordinateur. Utilisez-le pour essayer le serveur ou vérifier quelque chose
avant une vraie partie.

## Besoin d'aide

**Debug logs**, dans Settings, rassemble tout ce qui est utile dans un seul
fichier : le journal du lanceur, la console du serveur, vos réglages. Aucun
mot de passe. Envoyez-le à qui vous aide.

**Repair** réinstalle les mods sans toucher aux fichiers du jeu ni à votre
partie.

**Reset settings** remet chaque réponse à ce qu'a une installation neuve, pour
une partie dont les réglages ont dérivé quelque part que vous ne voyez pas. Ça
garde les fichiers du jeu et leur emplacement : rien n'est retéléchargé.

Voir [Dépannage](../operate/troubleshooting.md) pour le reste, et
[Installer avec Docker](install.md) si vous préférez faire tourner ceci sur
Linux.

---

## Référence

### Commandes

Lancez-les depuis un terminal. Sinon, la fenêtre s'ouvre seule.

| Commande | Ce qu'elle fait |
| --- | --- |
| `tf2ap.exe` | Ouvre la fenêtre |
| `tf2ap.exe -room <hôte:port>` | Règle l'adresse, puis ouvre la fenêtre |
| `tf2ap.exe -console` | Tourne dans le terminal, sans fenêtre |
| `tf2ap.exe -configure` | Édite tous les réglages dans le terminal, puis quitte |
| `tf2ap.exe -install` | Installe ou répare le serveur, puis quitte |
| `tf2ap.exe -status` | Affiche les réglages et l'état de l'installation |
| `tf2ap.exe -yaml <chemin>` | Écrit le fichier joueur Archipelago, puis quitte |
| `tf2ap.exe -env` | Liste les variables d'environnement, puis quitte |
| `tf2ap.exe -version` | Affiche la version et les versions des outils |

### Les réglages par l'environnement

Chaque réglage lit aussi une variable d'environnement, sous le nom qu'emploie
`deploy/.env.example`. Une variable l'emporte sur le fichier pour cette
exécution :

```bat
set AP_ROOM=archipelago.gg:12345
set SRCDS_BOT_TEAM_SIZE=4
tf2ap.exe
```

`tf2ap.exe -env` affiche chaque nom lu et marque ceux déjà réglés.

### Où le lanceur range ses fichiers

| Chemin | Contenu |
| --- | --- |
| `%USERPROFILE%\tf2-archipelago\` | Les fichiers du jeu, SourceMod et SteamCMD |
| `%USERPROFILE%\tf2-archipelago\tf2.yaml` | Le fichier joueur |
| `%USERPROFILE%\tf2-archipelago\bridge-state\` | Les checks et les déblocages |
| `%APPDATA%\tf2ap\config.json` | Vos réglages |
| `%LOCALAPPDATA%\Programs\Archipelago\` | L'app Archipelago, si installée là |

`TF2AP_INSTALL_ROOT` déplace les trois premiers, pour un second disque.
