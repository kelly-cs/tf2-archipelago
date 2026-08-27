# Code signing policy

[SignPath.org](https://signpath.org) gives open source projects free code
signing. The certificate comes from the
[SignPath Foundation](https://signpath.org).

> [!WARNING]
> The application is open, not granted. SignPath signs nothing yet.
> `tf2ap.exe` ships unsigned, and Windows warns about it until that changes.

## What SignPath signs

`tf2ap.exe`, the Windows launcher, and nothing else. Authenticode has nothing to
say about an ELF binary, so nothing ever signs `tf2ap-linux-amd64`.

The release workflow signs the file before anything else reads it. The
checksums, the build attestation and the upload therefore all describe the
signed file, not the file that went into signing.

## Checks you can run today

Windows warns about the file, and a player is right to want to check what they
downloaded. Every release publishes three things:

- `SHA256SUMS`, covering every file the release carries.
- A VirusTotal report on both binaries.
- A build attestation, which Sigstore signs. It ties each binary's hash to this
  repository, this workflow and this commit:

```sh
gh attestation verify tf2ap.exe --repo m-this/tf2-archipelago
```

The attestation is not an Authenticode signature and Windows does not read it,
so it changes no scanner's mind. What it gives is a check anybody can run
against a file they already have.

`make launcher` builds the same file on your own machine, which is the check
that trusts nothing at all.
