# Politique de signature de code

[SignPath.org](https://signpath.org) offre la signature de code gratuite aux
projets libres. Le certificat vient de la
[fondation SignPath](https://signpath.org).

> [!WARNING]
> Le projet a déposé la demande, SignPath ne l'a pas accordée. SignPath ne signe
> encore rien. Le projet publie donc `tf2ap.exe` sans signature, et Windows
> avertit à son sujet tant que cela ne change pas.

## Ce que SignPath signe

`tf2ap.exe`, le lanceur Windows, et rien d'autre. Authenticode n'a rien à dire
d'un binaire ELF, donc rien ne signe jamais `tf2ap-linux-amd64`.

Le workflow de publication signe le fichier avant que quoi que ce soit d'autre
le lise. Les sommes de contrôle, l'attestation de construction et le
téléversement décrivent donc tous le fichier signé. Aucun ne décrit le fichier
que le workflow a envoyé à la signature.

## Les vérifications possibles aujourd'hui

Windows avertit au sujet du fichier, et un joueur a raison de vouloir vérifier
ce qu'il a téléchargé. Chaque publication fournit trois choses :

- `SHA256SUMS`, qui couvre chaque fichier de la publication.
- Un rapport VirusTotal sur les deux binaires.
- Une attestation de construction, que Sigstore signe. Elle lie l'empreinte de
  chaque binaire à ce dépôt, à ce workflow et à ce commit :

```sh
gh attestation verify tf2ap.exe --repo m-this/tf2-archipelago
```

L'attestation n'est pas une signature Authenticode et Windows ne la lit pas.
Elle ne change donc l'avis d'aucun antivirus. Ce qu'elle apporte est une
vérification que n'importe qui peut faire sur un fichier déjà téléchargé.

`make launcher` construit le même fichier sur votre propre machine. C'est la
vérification qui ne fait confiance à personne.
