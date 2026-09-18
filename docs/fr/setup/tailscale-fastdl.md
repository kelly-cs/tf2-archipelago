# Téléchargement rapide des cartes avec Tailscale

FastDL laisse un joueur qui rejoint télécharger une carte communautaire en
HTTPS. Le transfert du serveur de jeu lui-même est lent et relance parfois le
téléchargement sans fin. Tailscale Funnel publie ces fichiers sans serveur web
à gérer et sans port à rediriger sur la box.

- Seul le serveur fait tourner Tailscale. Les joueurs utilisent une adresse
  HTTPS publique et n'installent rien.
- Cela ne change que le téléchargement des cartes. **Who can reach it** reste
  ce qu'il était.
- Funnel demande MagicDNS, des certificats HTTPS et la permission Funnel sur
  votre tailnet. La [page Funnel](https://tailscale.com/kb/1223/funnel) de
  Tailscale les explique.
- Funnel est public et a des limites de débit. Testez la plus grosse carte
  avant un événement.

## Windows

1. [Installez Tailscale](https://tailscale.com/download/windows) et
   connectez-vous depuis son icône de la barre des tâches.
2. Dans le lanceur, ouvrez **Settings**, puis **Networking**.
3. Appuyez sur **Set up / check Tailscale Funnel**. Si une page de navigateur
   vous demande d'approuver Funnel, approuvez, puis appuyez de nouveau sur le
   bouton.
4. Cochez **Tailscale FastDL**. Enregistrez, puis appuyez sur **Start**.
5. Cherchez `public Tailscale Funnel FastDL ready` dans le journal.

La vérification ne se fait qu'une fois. Le lanceur retient le réglage et
recrée la route à chaque démarrage. **Stop** retire la route et rien d'autre.

## Linux

1. Installez Tailscale et connectez-vous avec ses
   [instructions Linux](https://tailscale.com/download/linux).
2. Lancez la vérification :

```sh
./tf2ap-linux-amd64 -setup-funnel
```

Si elle affiche une URL d'approbation, ouvrez-la dans n'importe quel
navigateur, approuvez Funnel, et relancez la commande. Cela marche en SSH.

Si Tailscale répond `Access denied: serve config denied`, laissez votre
utilisateur gérer Funnel, puis relancez la vérification sans `sudo` :

```sh
sudo tailscale set --operator=$USER
```

Activez ensuite l'option :

- Avec un bureau : **Settings**, puis **Networking**, cochez
  **Tailscale FastDL**, enregistrez.
- Sans : `./tf2ap-linux-amd64 -configure` et répondez oui à la question sur
  Funnel.

Un service peut le forcer pour un lancement :

```sh
TAILSCALE_FASTDL=1 FASTDL_PORT=27080 ./tf2ap-linux-amd64 -console
```

À chaque démarrage, le lanceur vérifie que Tailscale est connecté et applique
cette route pour la durée de vie du serveur :

```text
https://nom-du-serveur.exemple-tailnet.ts.net/tf  ->  http://127.0.0.1:27080/tf
```

Si la connexion ou l'approbation Funnel a expiré, le lanceur s'arrête avant
le démarrage du serveur de jeu et affiche quoi faire.

### Nettoyage des routes

Le lanceur possède deux routes et ne touche à rien d'autre dans Tailscale :

- `/tf2ap-funnel-setup`, créée par `-setup-funnel` et retirée dès que la
  vérification réussit.
- `/tf`, créée à **Start** et retirée à **Stop**, **Restart**, **Quit**,
  Ctrl+C et à la sortie normale.

Un arrêt forcé ou une coupure de courant peut laisser `/tf` en place.
Retirez-la avec :

```sh
tailscale funnel --https=443 --set-path=/tf off
```

## Docker

La pile inclut le conteneur officiel `tailscale/tailscale` à côté du serveur
FastDL Caddy. Vous n'installez pas Tailscale sur la machine.

1. Réglez ceci dans `.env`, ou cochez **Tailscale FastDL** sur la page
   d'administration :

```ini
TAILSCALE_FASTDL=1
TAILSCALE_HOSTNAME=tf2-fastdl
FASTDL_BIND=127.0.0.1
```

2. Appliquez et ouvrez la page d'administration :

```sh
docker compose up -d --force-recreate
```

3. Sur **Settings**, puis **Networking**, appuyez sur
   **Set up / check Funnel**. Le premier appui donne un lien de connexion
   Tailscale. Connectez-vous, revenez, appuyez de nouveau. Si le tailnet n'a
   jamais utilisé Funnel, le deuxième appui donne un lien d'approbation.
   Approuvez, appuyez encore une fois.
4. Attendez le message de prêt. Le journal du serveur de jeu affiche alors :

```text
[AP] using Tailscale Funnel FastDL at https://tf2-fastdl.example.ts.net/tf
```

Le serveur de jeu attend la route avant de démarrer. Si Tailscale ne peut pas
se connecter, le serveur de jeu attend, et la page d'administration reste
disponible. Inspectez le conteneur avec :

```sh
docker compose logs tailscale-fastdl
```

Le volume `tailscale_fastdl_state` garde l'identité de l'appareil.
`make clean` le supprime, et le démarrage suivant demande une nouvelle
connexion.

## Ce que font les amis

Rien. Ils rejoignent avec la ligne de connexion normale. Le jeu lit l'adresse
de téléchargement depuis le serveur et va chercher chaque fichier manquant à
cette adresse.

## Ce qui devient public

Le serveur ne publie que `maps`, `materials`, `models`, `sound`, `particles`
et `resource`. Les fichiers de configuration, les plugins et les mots de
passe ne le sont pas. L'adresse ne liste aucun fichier.
