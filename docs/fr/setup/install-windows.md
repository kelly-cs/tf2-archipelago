# Installer sur Windows

Un seul fichier. Pas de Docker, pas de clone, pas de compilateur.

## 1. Télécharger et lancer

1. Téléchargez `tf2ap.exe` depuis la
   [dernière version](https://github.com/m-this/tf2-archipelago/releases/latest).
2. Double-cliquez dessus. Un onglet de navigateur s'ouvre avec le lanceur.

### Windows va vous avertir

SmartScreen bloque le premier lancement. Cliquez sur
**Informations complémentaires**, puis sur **Exécuter quand même**. Defender
met parfois le fichier en quarantaine à la place. Restaurez-le et ajoutez une
exclusion.

L'avertissement est un faux positif. Le lanceur décompresse des archives,
écrit des DLL dans un dossier de jeu, télécharge un serveur et le démarre.
C'est ce que fait un installeur, et aussi ce que fait un virus. Le fichier n'a
pas encore de signature de code, donc l'antivirus ne fait pas la différence.
Voir [Politique de signature de code](https://github.com/m-this/tf2-archipelago/blob/main/design/code-signing.md).

Pour vérifier le fichier vous-même :

- Comparez `Get-FileHash tf2ap.exe -Algorithm SHA256` avec `SHA256SUMS` sur la
  page de la version.
- Ouvrez le rapport VirusTotal lié sur la page de la version.
- Lancez `gh attestation verify tf2ap.exe --repo m-this/tf2-archipelago`.

## 2. Appuyer sur Start

Le premier démarrage installe SteamCMD, le serveur dédié TF2, SourceMod, le
plugin et les bots. Il télécharge environ 14 Go, donc il prend du temps. Tous
les démarrages suivants prennent quelques secondes.

Vous n'avez pas encore besoin d'une adresse de room. Sans adresse, le serveur
tourne, et l'onglet **Play** dit qu'il attend une room.

## 3. Créer la session

Le lanceur fait tourner le serveur de jeu. La session Archipelago est à part,
et c'est l'application Archipelago officielle qui génère la seed.

1. Installez l'[application Archipelago](https://github.com/ArchipelagoMW/Archipelago/releases).
   Le lanceur la trouve aux emplacements habituels.
2. Ouvrez **Settings**, puis **Player options**. Choisissez les
   [options de la partie](shape-of-the-run.md).
3. Appuyez sur **Generate seed**. Le lanceur écrit le fichier joueur, lance le
   générateur et ouvre le dossier qui contient le résultat.
4. Envoyez ce fichier sur [archipelago.gg/uploads](https://archipelago.gg/uploads)
   et cliquez sur **Create New Room**.
5. Copiez l'adresse de la room, du type `archipelago.gg:12345`, dans
   **Settings**, puis **Archipelago room**. Enregistrez, puis appuyez sur
   **Restart**.

[Créer la session](create-the-session.md) détaille chaque étape, et dit quoi
faire si **Generate seed** ne trouve pas l'application Archipelago.

## 4. Inviter vos amis

La ligne **Join**, sous les boutons, montre la ligne de connexion à donner.
Par défaut, seul votre réseau local atteint le serveur. Pour que des amis
rejoignent par internet, voir [Inviter vos amis](invite-your-friends.md).

## L'écran

| Onglet | Ce qu'il contient |
| --- | --- |
| **Play** | La ligne de connexion, les classes débloquées, l'équipe de bots, les missions de la partie, et le journal avec un champ rcon. **Play** sur la ligne d'une mission la charge. |
| **Unlocks** | Tout ce que le multiworld a donné à votre serveur. |
| **Bots** | L'équipe RED : la classe de chaque place et ce qu'elle porte. **Apply** change l'équipe sans terminer la mission. |
| **Settings** | Les options de la partie, la room, les missions, les bots, le réseau. |

**Start**, **Stop**, **Restart** et **Quit** sont en haut de chaque onglet.
Fermer l'onglet du navigateur laisse le serveur tourner. **Quit** l'arrête.

![Les réglages, sur la réserve de missions](../../images/launcher-settings.png)

[Le lanceur, onglet par onglet](the-launcher.md) passe chaque écran en revue.

Trois boutons dans Settings aident quand quelque chose ne va pas :

- **Debug logs** écrit un seul fichier avec le journal du lanceur, la console
  du serveur et vos réglages, sans mot de passe. Envoyez-le quand vous
  demandez de l'aide.
- **Repair** réinstalle les mods. Il garde les fichiers de jeu et la partie.
- **Reset settings** remet chaque réglage à sa valeur par défaut. Il garde les
  fichiers de jeu.

## Essayer sans Archipelago

**Test mode**, dans **Settings**, puis **Archipelago room**, fait tourner un
multiworld d'un seul joueur sur votre machine. Pas de room, pas de seed, rien
ne quitte votre ordinateur. Utilisez-le pour essayer le serveur.

## Référence

### Ligne de commande

Double-cliquer sur l'exe ouvre le navigateur. Depuis un terminal :

| Commande | Ce qu'elle fait |
| --- | --- |
| `tf2ap.exe` | Servir l'interface et ouvrir un navigateur dessus |
| `tf2ap.exe -room <hôte:port>` | Régler d'abord l'adresse de la room |
| `tf2ap.exe -no-browser` | Afficher l'adresse au lieu d'ouvrir un navigateur |
| `tf2ap.exe -addr 127.0.0.1:8080` | Servir sur une adresse fixe |
| `tf2ap.exe -console` | Afficher le journal et rien d'autre |
| `tf2ap.exe -configure` | Modifier chaque réglage dans le terminal, puis quitter |
| `tf2ap.exe -install` | Installer ou réparer le serveur, puis quitter |
| `tf2ap.exe -status` | Afficher les réglages et l'état de l'installation |
| `tf2ap.exe -yaml <chemin>` | Écrire le fichier joueur Archipelago, puis quitter |
| `tf2ap.exe -env` | Lister les variables d'environnement lues, puis quitter |
| `tf2ap.exe -version` | Afficher la version et les versions des outils |

### Variables d'environnement

Chaque réglage lit aussi une variable d'environnement, avec les noms donnés
dans [Les options de la partie](shape-of-the-run.md). Une variable l'emporte
sur les réglages enregistrés pour ce lancement :

```bat
set AP_ROOM=archipelago.gg:12345
set SRCDS_BOT_TEAM_SIZE=4
tf2ap.exe
```

### Où il range les choses

| Chemin | Contenu |
| --- | --- |
| `%USERPROFILE%\tf2-archipelago\` | Les fichiers de jeu, SourceMod et SteamCMD |
| `%USERPROFILE%\tf2-archipelago\tf2.yaml` | Le fichier joueur |
| `%USERPROFILE%\tf2-archipelago\bridge-state\` | Les checks et les déblocages de la partie |
| `%APPDATA%\tf2ap\config.json` | Vos réglages |
| `%LOCALAPPDATA%\Programs\Archipelago\` | L'application Archipelago, si elle est installée là |

**Install folder**, dans **Settings**, puis **Player options**, déplace les
trois premiers sur un autre disque.

Suite : [Créer la session](create-the-session.md).
