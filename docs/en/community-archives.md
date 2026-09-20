# Community archive snapshots

The launcher expects these **complete ZIP files, including maps and NAV files**. SHA-256 covers the ZIP bytes, not the extracted files. These are the snapshots used to review the mission catalog; changing an archive can change a mission without changing its name.

| Pack | Source | Bytes | SHA-256 |
| --- | --- | ---: | --- |
| Potato `archive-assets.zip` | `https://dlarchive.potato.tf/archive-assets.zip` | 2,594,253,886 | `e7e54f3167b97341d11cf1a1b30f437bf0651fec40e4e1d25232b883cf44bb69` |
| Moonlight `mlarchive-assets.zip` | `https://dlml.potato.tf/mlarchive-assets.zip` | 190,820,351 | `c6ba6c85c4466f012094388a59e15e4973c532cd1fd3bb2f404f1d7abc980149` |

To check a downloaded file yourself, run `sha256sum archive-assets.zip mlarchive-assets.zip` on Linux or `Get-FileHash archive-assets.zip -Algorithm SHA256` in PowerShell. Compare with the full hashes above. The launcher checks these hashes when it downloads or imports a pack and again before installation. A differing ZIP is held as `*.hash-mismatch`; the Missions page shows an **Ignore hash mismatch** confirmation if you decide to use those exact bytes despite possible missing missions or server instability. Any later file change requires a new confirmation.
