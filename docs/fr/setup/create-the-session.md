# Créer la session

Une session Archipelago est une **seed** hébergée dans une **room**. Le
serveur joue une room qui existe déjà. Cette page en crée une.

1. Choisissez les [options de la partie](shape-of-the-run.md).
2. Générez la seed avec l'application Archipelago.
3. Envoyez la seed sur `archipelago.gg` et créez une room.
4. Donnez l'adresse de la room au serveur.

Mann vs Machine ne fait pas partie des jeux livrés avec Archipelago. Le site
ne peut donc pas générer la seed pour vous. Votre machine la génère, et le site
l'héberge.

## 1. Choisir les options de la partie

Les options de la partie décident de sa longueur, de sa difficulté et de ce
qui la termine. La seed les garde. Pour en changer une plus tard, il faut une
nouvelle seed et une nouvelle room.

- **Lanceur :** ouvrez **Settings**, puis **Player options**, **Rewards**,
  **Balancing** et **Missions**.
- **Docker :** modifiez les lignes `MVM_` dans `.env`.
- **Application Archipelago à la main :** modifiez le fichier YAML. Voir
  [Avec l'application Archipelago](#avec-lapplication-archipelago).

[Les options de la partie](shape-of-the-run.md) décrit chaque option.

## 2. Générer la seed

### Avec le lanceur

1. Installez l'[application Archipelago](https://github.com/ArchipelagoMW/Archipelago/releases).
   Prenez la version que le lanceur fixe. `tf2ap.exe -version` l'affiche.
2. Dans le lanceur, ouvrez **Settings**, puis **Player options**.
3. Appuyez sur **Generate seed**. Le lanceur écrit `tf2.yaml`, lance le
   générateur et ouvre le dossier qui contient le résultat. Le résultat est un
   `.zip` du type `AP_53174869021847362095.zip`.

Si **Generate seed** dit qu'il ne trouve pas l'application Archipelago,
réglez **Archipelago app** sur la même page avec le dossier de l'application.

**Check Run Selection**, sur la page **Missions**, vérifie la réserve avant
de générer. Il vous dit si les missions choisies contiennent assez de checks
pour les objets de la partie.

### Avec l'application Archipelago

Utilisez cette méthode pour jouer avec des gens dans d'autres jeux, ou pour
modifier le YAML à la main.

1. Téléchargez `tf2_mvm.apworld` depuis la
   [version](https://github.com/m-this/tf2-archipelago/releases/latest).
2. Double-cliquez dessus, ou copiez-le dans le dossier `custom_worlds/` de
   l'application.
3. Obtenez un fichier joueur. Soit lancez `tf2ap.exe -yaml tf2.yaml` avec le
   lanceur, soit appuyez sur **Generate Template Options** dans le Launcher de
   l'application et prenez `Team Fortress 2 Mann vs Machine.yaml`.
4. Modifiez le fichier. [Les options de la partie](shape-of-the-run.md) liste
   les options.
5. Mettez-le dans le dossier `Players/` de l'application, à côté des fichiers
   des autres joueurs.
6. Lancez **Generate**. Le résultat est dans le dossier `output/` de
   l'application.

Le `name` du fichier est le nom du slot. Le **Slot name** du lanceur, dans
**Settings**, puis **Archipelago room**, doit être le même. La valeur par
défaut est `tf2`.

### Avec Docker

```sh
make seed
```

La commande écrit la seed dans `seed/` et affiche son nom. Gardez les fichiers
dans `seed/`. Une room perdue revient depuis son fichier.

Si l'envoi échoue à cause de la version, lisez la version d'Archipelago en bas
du site. Réglez `ARCHIPELAGO_VERSION` dans `deploy/env/versions.env` sur cette
version et relancez `make seed`.

## 3. Envoyer la seed et créer une room

1. Ouvrez [archipelago.gg/uploads](https://archipelago.gg/uploads).
2. Envoyez le `.zip`.
3. Cliquez sur **Create New Room**.

Le site ne demande pas de compte. La page de la room affiche :

- l'adresse de la room, du type `archipelago.gg:12345`,
- un lien vers le tracker, où vos joueurs suivent la partie depuis un
  navigateur.

Chaque nouvelle room reçoit un nouveau port. Toute personne qui a l'adresse
peut rejoindre la room. Mettez donc un mot de passe sur la page de la room si
l'adresse sort du cercle de vos amis.

## 4. Donner l'adresse de la room au serveur

- **Lanceur :** collez l'adresse dans **Settings**, puis **Archipelago room**.
  Mettez-y aussi le mot de passe de la room, si vous en avez mis un.
  Enregistrez, puis appuyez sur **Restart**.
- **Docker :** écrivez les deux moitiés dans `.env`, puis `make restart` :

```sh
AP_HOST=archipelago.gg
AP_PORT=12345
AP_TLS=true
AP_PASSWORD=
```

La ligne d'état du lanceur, ou le journal du bridge, dit ensuite `connected to
archipelago`. La page de la room dit `tf2 (Team #1) playing Team Fortress 2
Mann vs Machine has joined`.

## Héberger la session vous-même

La pile Docker peut héberger la room sur votre machine. Mettez ces lignes
dans `.env` :

```sh
COMPOSE_PROFILES=selfhost
AP_HOST=archipelago
AP_PORT=38281
AP_TLS=false
```

`make up` démarre alors un troisième conteneur qui génère la seed à son
premier démarrage et l'héberge. Vous n'envoyez rien.

Ce que cela coûte :

- Vos joueurs n'ont ni page de room ni tracker.
- Un joueur dans un autre jeu a besoin d'un deuxième port public.
  `deploy/compose.yml` dit lequel.

Suite : [Les options de la partie](shape-of-the-run.md), ou
[Inviter vos amis](invite-your-friends.md) si la session est prête.
