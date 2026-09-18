# Démarrer une nouvelle partie

Une seed ne change jamais une fois créée. Une autre partie demande une
nouvelle seed et une nouvelle room. Les fichiers de jeu ne bougent pas, donc
cela prend quelques minutes.

## Avec le lanceur

1. Changez les [options de la partie](../setup/shape-of-the-run.md) dans
   **Settings**, si vous voulez une partie différente.
2. Appuyez sur **Generate seed** sur la page **Player options**.
3. Envoyez le nouveau fichier sur
   [archipelago.gg/uploads](https://archipelago.gg/uploads) et créez une room.
4. Collez la nouvelle adresse de room dans **Settings**, puis
   **Archipelago room**.
5. Enregistrez, puis appuyez sur **Restart**.

## Avec Docker

1. Modifiez les lignes `MVM_` dans `.env`, si vous voulez une partie
   différente.
2. Lancez `make seed`. Il écrit un autre fichier dans `seed/`.
3. Envoyez ce fichier et créez une room.
4. Réglez `AP_PORT` sur le port de la nouvelle room.
5. Lancez `make restart`.

Gardez les anciens fichiers dans `seed/`. Chacun est une partie entière, et la
room d'une partie revient depuis son fichier.

### Si vous hébergez la session vous-même

Avec `COMPOSE_PROFILES=selfhost`, la seed vit dans le volume
`tf2-archipelago_apoutput`. Une nouvelle partie, c'est :

```sh
make down
docker volume rm tf2-archipelago_apoutput
make up
```

Modifiez `.env` entre la première et la troisième commande.

## Ce qui arrive à l'ancienne partie

Le bridge remarque que la room n'est pas celle dont il garde l'état. Alors :

1. il met son fichier d'état de côté, sous `bridge.<seed>.json` dans le même
   dossier ;
2. il repart sans check et sans déblocage ;
3. il dit au plugin que la partie a redémarré.

Rien n'est écrasé. Si vous pointez le serveur vers la mauvaise room par
erreur, l'ancienne partie est toujours sur le disque.

Rien d'autre ne perd une partie. Redémarrer le serveur, redémarrer la machine
et s'arrêter une semaine la gardent tous.

## Tout recommencer

- **Lanceur :** **Reset settings** dans **Settings** remet chaque réglage à sa
  valeur par défaut et garde les fichiers de jeu.
- **Docker :** `make clean` arrête la pile et supprime chaque volume, y
  compris les 14 Go de fichiers de jeu. Utilisez-le quand vous en avez fini
  avec le projet, pas entre deux parties.
