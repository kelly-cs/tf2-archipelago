# MvM wave smoke probe

Subset: 91 prior Waves to inspect cases from waveprobe-20260921-main-full-final.

## Stability score: **15.4%** (14/91 tested wave/mode cases passed)

**Verified coverage:** 15.4% (14/91 eligible cases); **tested:** 91/91; **missions:** 45. A pass requires the game's wave-complete event and an observed enemy spawn.

The stability score uses tested cases; verified coverage uses the full eligible plan, so unrun waves cannot silently improve it. Reverse objectives are excluded.

This score measures the current automated probe at the configured speed and kill timing. A nonpass is a reproducible probe observation, not by itself proof that a player cannot finish the wave.

Started 2026-09-22T07:15:16Z · 12 isolated servers · requested speed 20× · 900s **real-time** limit after wave start · seed 1.

Every wave gets a 180s first pass; every nonpass is retried with the 900s real-time limit before the final score.

Source commit `58131b41b942dcce6b3ed9995fe584c80b5714de` · upstream main `c31e4d1030d2b576fbf6ee873e042340305cd30f`.

Retry runner commit `f8d3456402f3a554de2d488ab6fe874a3dad560d`. The first pass used the source commit above.

| Result | Bot Surge off | Bot Surge on | Total |
| --- | ---: | ---: | ---: |
| Passed | 7 | 7 | **14** |
| Wave Failed | 4 | 3 | **7** |
| Wave 0 | 0 | 0 | **0** |
| No Enemies Spawned | 14 | 11 | **25** |
| Wave Timed Out | 22 | 18 | **40** |
| Probe Error | 4 | 0 | **4** |
| Changelevel Failure | 1 | 0 | **1** |
| Load Blocked | 0 | 0 | **0** |
| Inconclusive | 0 | 0 | **0** |
| Not Run | 0 | 0 | **0** |
| Reverse Objective | 0 | 0 | **0** |

**Wave failed** is the game's loss event; **Wave 0** means the population manager did not initialize a wave; **No enemies spawned** means no BLU bot or tank was observed in a standard wave. **Wave timed out** means no completion after the real-time limit despite observed enemies. **Changelevel failure** means the responsive server remained on a different map after the load deadline; the command reply and map states are recorded below. Probe errors and other load blocks are separate.

| Mode | Stability | Passed | Tested | Eligible |
| --- | ---: | ---: | ---: | ---: |
| Bot Surge off | 13.5% | 7 | 52 | 52 |
| Bot Surge on | 17.9% | 7 | 39 | 39 |

## Waves to inspect

Each row is one wave in one mode. Progress uses the game's non-support enemy counter relative to its highest observed value; `unknown` means no baseline.

| Map | Mission | Mode | Wave | Result | Spawned | Auto killed | Alive | Progress | Real / game s |
| --- | --- | --- | ---: | --- | ---: | ---: | ---: | ---: | ---: |
| mvm_area_52_rc3 | mvm_area_52_rc3_adv_complex_chaos | Bot Surge off | 1 | Wave Timed Out | 4 | 0 | 0 | 0% | 902 / 18033 |
| mvm_bigrock | mvm_bigrock_adv_breaking_point | Bot Surge off | 8 | Wave Failed | 110 | 18 | 1 | 0% | 9 / 181 |
| mvm_bigrock | mvm_bigrock_adv_breaking_point | Bot Surge on | 8 | Wave Failed | 288 | 29 | 1 | 0% | 12 / 172 |
| mvm_bigrock | mvm_bigrock_int_bionic_breach | Bot Surge off | 7 | Wave Failed | 398 | 10 | 1 | 0% | 11 / 194 |
| mvm_bigrock | mvm_bigrock_int_bionic_breach | Bot Surge on | 7 | Wave Failed | 451 | 10 | 1 | 0% | 11 / 183 |
| mvm_coaltown | mvm_coaltown_exp_catastrophic_course | Bot Surge off | 1 | Wave Timed Out | 9705 | 702 | 1 | 0% | 901 / 18028 |
| mvm_coaltown | mvm_coaltown_exp_catastrophic_course | Bot Surge on | 2 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18041 |
| mvm_condemned_b3 | mvm_condemned_b3_exp_trespasser | Bot Surge on | 1 | Wave Timed Out | 2626 | 595 | 16 | 0% | 901 / 2176 |
| mvm_condemned_b3 | mvm_condemned_b3_exp_trespasser_remaster | Bot Surge off | 1 | Wave Timed Out | 1212 | 966 | 9 | 83% | 902 / 2273 |
| mvm_condemned_b3 | mvm_condemned_b3_exp_trespasser_remaster | Bot Surge on | 1 | Wave Timed Out | 1270 | 1096 | 7 | 83% | 902 / 3127 |
| mvm_creepside_b2 | mvm_creepside_b2_adv_catastrophic_conjuring | Bot Surge off | 1 | Wave Timed Out | 3600 | 0 | 0 | 0% | 902 / 18033 |
| mvm_creepside_b2 | mvm_creepside_b2_adv_dismal_devilry | Bot Surge off | 6 | Wave Timed Out | 161 | 15 | 0 | 0% | 902 / 17938 |
| mvm_creepside_b2 | mvm_creepside_b2_adv_dismal_devilry | Bot Surge on | 6 | Wave Timed Out | 190 | 1 | 0 | 0% | 901 / 17634 |
| mvm_creepside_b2 | mvm_creepside_b2_adv_dust_to_dust | Bot Surge on | 1 | Wave Timed Out | 277 | 277 | 0 | 0% | 902 / 18033 |
| mvm_creepside_b2 | mvm_creepside_b2_adv_dust_to_dust | Bot Surge off | 2 | Wave Timed Out | 372 | 372 | 0 | 0% | 902 / 18034 |
| mvm_creepside_b2 | mvm_creepside_b2_adv_dust_to_dust | Bot Surge off | 5 | Wave Timed Out | 4315 | 30 | 7 | 0% | 902 / 3384 |
| mvm_decoy | mvm_decoy_intermediate | Bot Surge on | 3 | Wave Timed Out | 854 | 852 | 2 | 0% | 902 / 18033 |
| mvm_decoy | mvm_decoy_intermediate | Bot Surge off | 4 | No Enemies Spawned | 0 | 0 | 0 | 0% | 901 / 18019 |
| mvm_frostwynd_rc1 | mvm_frostwynd_rc1_int_wicked_wizardry | Bot Surge off | 3 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18041 |
| mvm_frostwynd_rc1 | mvm_frostwynd_rc1_int_wicked_wizardry | Bot Surge off | 6 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18041 |
| mvm_ghost_town | mvm_ghost_town_adv_horrorsome_happenings | Bot Surge off | 7 | Wave Timed Out | 132 | 114 | 0 | 0% | 902 / 17973 |
| mvm_hideout_b3 | mvm_hideout_b3_adv_perilous_paradise | Bot Surge on | 2 | Wave Timed Out | 300 | 300 | 0 | 0% | 902 / 18036 |
| mvm_hideout_b3 | mvm_hideout_b3_adv_perilous_paradise | Bot Surge off | 4 | Wave Timed Out | 536 | 536 | 0 | 0% | 902 / 18033 |
| mvm_hideout_b3 | mvm_hideout_b3_int_intermediate | Bot Surge off | 3 | Probe Error | 44 | 22 | 0 | 0% | 692 / 12487 |
| mvm_kelly_rc1b | mvm_kelly_rc1b_adv_mobocracy | Bot Surge off | 1 | Wave Timed Out | 368 | 360 | 0 | 0% | 901 / 18025 |
| mvm_kelly_rc1b | mvm_kelly_rc1b_adv_mobocracy | Bot Surge on | 3 | Wave Timed Out | 3523 | 910 | 8 | 0% | 902 / 9320 |
| mvm_kelly_rc1b | mvm_kelly_rc1b_adv_ominous_outback | Bot Surge off | 6 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18033 |
| mvm_mannhattan | mvm_mannhattan_adv_scorched_skies | Bot Surge off | 1 | Wave Timed Out | 1038 | 438 | 0 | 0% | 902 / 18033 |
| mvm_mannworks | mvm_mannworks | Bot Surge on | 5 | Wave Timed Out | 16659 | 8468 | 13 | 0% | 902 / 12650 |
| mvm_mannworks | mvm_mannworks_advanced | Bot Surge off | 2 | Wave Timed Out | 440 | 440 | 0 | 0% | 902 / 18031 |
| mvm_mannworks | mvm_mannworks_advanced | Bot Surge on | 2 | Wave Timed Out | 440 | 440 | 0 | 0% | 902 / 18039 |
| mvm_mannworks | mvm_mannworks_intermediate2 | Bot Surge on | 1 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18040 |
| mvm_null_b9c | mvm_null_b9c_exp_0 | Bot Surge off | 3 | Wave Failed | 244 | 46 | 1 | 0% | 15 / 253 |
| mvm_null_b9c | mvm_null_b9c_exp_0 | Bot Surge on | 3 | Wave Failed | 264 | 41 | 1 | 0% | 24 / 223 |
| mvm_oilrig_rc5d | mvm_oilrig_rc5d_adv_rescindment | Bot Surge off | 1 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18039 |
| mvm_oilrig_rc5d | mvm_oilrig_rc5d_adv_rescindment | Bot Surge on | 1 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18038 |
| mvm_oilrig_rc5d | mvm_oilrig_rc5d_adv_rescindment | Bot Surge off | 2 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18036 |
| mvm_oilrig_rc5d | mvm_oilrig_rc5d_adv_rescindment | Bot Surge on | 2 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18036 |
| mvm_oilrig_rc5d | mvm_oilrig_rc5d_adv_rescindment | Bot Surge off | 3 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18034 |
| mvm_oilrig_rc5d | mvm_oilrig_rc5d_adv_rescindment | Bot Surge off | 5 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18039 |
| mvm_oilrig_rc5d | mvm_oilrig_rc5d_adv_rescindment | Bot Surge off | 6 | Probe Error | 0 | 0 | 0 | 0% | 450 / 7882 |
| mvm_oilrig_rc5d | mvm_oilrig_rc5d_int_orson_oligarchy | Bot Surge on | 4 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18031 |
| mvm_oxidize_rc3 | mvm_oxidize_rc3_int_snowy_slaughter | Bot Surge off | 2 | Wave Timed Out | 71406 | 145 | 0 | 0% | 901 / 17704 |
| mvm_radar_b10 | mvm_radar_b10_adv_rocky_ravage | Bot Surge on | 2 | Wave Timed Out | 459 | 452 | 1 | 0% | 902 / 18029 |
| mvm_radar_b10 | mvm_radar_b10_adv_rocky_ravage | Bot Surge off | 6 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18038 |
| mvm_radar_b10 | mvm_radar_b10_adv_rocky_ravage | Bot Surge on | 6 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18036 |
| mvm_radar_b10 | mvm_radar_b10_int_recruits_recognition | Bot Surge off | 2 | Wave Timed Out | 7125 | 1 | 1 | 0% | 902 / 1753 |
| mvm_redstone_ridge_rc5 | mvm_redstone_ridge_rc5_adv_armored_apparatus | Bot Surge off | 5 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18039 |
| mvm_redstone_ridge_rc5 | mvm_redstone_ridge_rc5_adv_armored_apparatus | Bot Surge on | 5 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18038 |
| mvm_rottenburg | mvm_rottenburg_adv_cybernetic_carnage | Bot Surge off | 4 | Wave Timed Out | 760 | 759 | 1 | 0% | 902 / 18034 |
| mvm_rottenburg | mvm_rottenburg_adv_marathon | Bot Surge off | 4 | Wave Timed Out | 911 | 864 | 0 | 0% | 901 / 18025 |
| mvm_rottenburg | mvm_rottenburg_adv_marathon | Bot Surge on | 4 | Wave Timed Out | 913 | 864 | 2 | 0% | 902 / 18028 |
| mvm_sharp_rc9 | mvm_sharp_rc9_adv_sudden_equinox | Bot Surge off | 4 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18038 |
| mvm_sharp_rc9 | mvm_sharp_rc9_adv_sudden_equinox | Bot Surge on | 4 | No Enemies Spawned | 0 | 0 | 0 | 0% | 901 / 18023 |
| mvm_sharp_rc9 | mvm_sharp_rc9_exp_gilead_gehenom | Bot Surge off | 3 | Wave Timed Out | 17 | 16 | 0 | 0% | 902 / 18039 |
| mvm_sludge_b6 | mvm_sludge_b6_int_logging_dreadwood | Bot Surge off | 3 | Wave Timed Out | 15334 | 5565 | 10 | 0% | 902 / 15696 |
| mvm_sludge_b6 | mvm_sludge_b6_int_logging_dreadwood | Bot Surge on | 3 | Wave Timed Out | 3722 | 303 | 8 | 0% | 902 / 1603 |
| mvm_sludge_b6 | mvm_sludge_b6_int_logging_dreadwood | Bot Surge on | 6 | Wave Timed Out | 360 | 360 | 0 | 0% | 902 / 18035 |
| mvm_snowpine_rc4_fix1 | mvm_snowpine_rc4_fix1_adv_permafrost_panic | Bot Surge off | 1 | Changelevel Failure | 0 | 0 | 0 | 0% | 0 / 0 |
| mvm_snowpine_rc4_fix1 | mvm_snowpine_rc4_fix1_adv_permafrost_panic | Bot Surge on | 1 | No Enemies Spawned | 0 | 0 | 0 | 0% | 901 / 18020 |
| mvm_snowpine_rc4_fix1 | mvm_snowpine_rc4_fix1_adv_permafrost_panic | Bot Surge off | 3 | Probe Error | 211 | 186 | 0 | 0% | 541 / 9274 |
| mvm_snowpine_rc4_fix1 | mvm_snowpine_rc4_fix1_adv_permafrost_panic | Bot Surge on | 3 | Wave Timed Out | 386 | 361 | 0 | 0% | 902 / 18035 |
| mvm_snowpine_rc4_fix1 | mvm_snowpine_rc4_fix1_adv_permafrost_panic | Bot Surge off | 5 | Wave Timed Out | 9667 | 718 | 2 | 0% | 901 / 18030 |
| mvm_snowpine_rc4_fix1 | mvm_snowpine_rc4_fix1_int_alpine_assault | Bot Surge off | 1 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18035 |
| mvm_snowpine_rc4_fix1 | mvm_snowpine_rc4_fix1_int_alpine_assault | Bot Surge on | 1 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18038 |
| mvm_snowpine_rc4_fix1 | mvm_snowpine_rc4_fix1_int_alpine_assault | Bot Surge off | 3 | No Enemies Spawned | 0 | 0 | 0 | 0% | 901 / 18025 |
| mvm_snowpine_rc4_fix1 | mvm_snowpine_rc4_fix1_int_alpine_assault | Bot Surge on | 3 | No Enemies Spawned | 0 | 0 | 0 | 0% | 902 / 18036 |
| mvm_snowpine_rc4_fix1 | mvm_snowpine_rc4_fix1_int_alpine_assault | Bot Surge off | 5 | No Enemies Spawned | 0 | 0 | 0 | 0% | 901 / 18022 |
| mvm_spybase_rc8 | mvm_spybase_rc8_exp_waters_of_a_robot_regime | Bot Surge off | 6 | Probe Error | 3 | 1 | 0 | 0% | 641 / 11696 |
| mvm_spybase_rc8 | mvm_spybase_rc8_exp_waters_of_a_robot_regime | Bot Surge on | 6 | Wave Timed Out | 19 | 1 | 0 | 0% | 901 / 18027 |
| mvm_terrorlict_final1c5 | mvm_terrorlict_final1c5_adv_accursed_aggrievocation | Bot Surge off | 2 | Wave Timed Out | 1990 | 1982 | 2 | 0% | 901 / 18023 |
| mvm_terrorlict_final1c5 | mvm_terrorlict_final1c5_adv_accursed_aggrievocation | Bot Surge off | 4 | Wave Timed Out | 469 | 469 | 0 | 0% | 902 / 18036 |
| mvm_terrorlict_final1c5 | mvm_terrorlict_final1c5_adv_accursed_aggrievocation | Bot Surge on | 4 | Wave Timed Out | 469 | 469 | 0 | 0% | 902 / 18035 |
| mvm_terrorlict_final1c5 | mvm_terrorlict_final1c5_adv_accursed_aggrievocation | Bot Surge off | 6 | Wave Timed Out | 1050 | 405 | 1 | 0% | 901 / 18027 |
| mvm_terrorlict_final1c5 | mvm_terrorlict_final1c5_adv_accursed_aggrievocation | Bot Surge on | 6 | Wave Timed Out | 6950 | 3250 | 22 | 0% | 902 / 8044 |
| mvm_terrorlict_final1c5 | mvm_terrorlict_final1c5_exp_echoes_of_a_warzone | Bot Surge off | 1 | Wave Failed | 77 | 36 | 1 | 0% | 11 / 217 |
| mvm_terrorlict_final1c5 | mvm_terrorlict_final1c5_exp_echoes_of_a_warzone | Bot Surge on | 1 | Wave Timed Out | 7050 | 2896 | 8 | 0% | 902 / 14310 |

### Timelines

<details><summary>mvm_area_52_rc3_adv_complex_chaos · Bot Surge off · wave 1 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 19 | 2 | 0 | 0 | 93 | 0% |
| 80 | 1628 | 4 | 0 | 0 | 93 | 0% |
| 161 | 3235 | 4 | 0 | 0 | 93 | 0% |
| 251 | 5045 | 4 | 0 | 0 | 93 | 0% |
| 332 | 6653 | 4 | 0 | 0 | 93 | 0% |
| 412 | 8262 | 4 | 0 | 0 | 93 | 0% |
| 493 | 9871 | 4 | 0 | 0 | 93 | 0% |
| 573 | 11479 | 4 | 0 | 0 | 93 | 0% |
| 653 | 13087 | 4 | 0 | 0 | 93 | 0% |
| 744 | 14896 | 4 | 0 | 0 | 93 | 0% |
| 824 | 16505 | 4 | 0 | 0 | 93 | 0% |
| 901 | 18033 | 4 | 0 | 0 | 93 | 0% |

Reason: wave exceeded wall-clock time limit: wave 1 still active after 15m0s (18032.9 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=1 expected=1 game=18057.2 engine=927.0 timescale=20.0 remaining=93 spawns=4 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_bigrock_adv_breaking_point · Bot Surge off · wave 8 · Wave Failed</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 9 | 0 | 6 | 2 | 0% |
| 8 | 181 | 110 | 18 | 1 | 2 | 0% |

Reason: game or probe failed wave 8: wave_failed

Failure snapshot:

```text
WAVEPROBE_DEBUG state=4 wave=8 expected=8 game=206.3 engine=33.8 timescale=20.0 remaining=2 spawns=110 kills=18 attempts=18 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_bigrock_adv_breaking_point · Bot Surge on · wave 8 · Wave Failed</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 19 | 18 | 0 | 17 | 2 | 0% |
| 10 | 154 | 281 | 28 | 12 | 2 | 0% |
| 11 | 172 | 288 | 29 | 1 | 2 | 0% |

Reason: game or probe failed wave 8: wave_failed

Failure snapshot:

```text
WAVEPROBE_DEBUG state=4 wave=8 expected=8 game=198.7 engine=946.2 timescale=20.0 remaining=2 spawns=288 kills=29 attempts=29 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_bigrock_int_bionic_breach · Bot Surge off · wave 7 · Wave Failed</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 35 | 0 | 8 | 2 | 0% |
| 10 | 194 | 398 | 10 | 1 | 2 | 0% |

Reason: game or probe failed wave 7: wave_failed

Failure snapshot:

```text
WAVEPROBE_DEBUG state=4 wave=7 expected=7 game=219.7 engine=2759.3 timescale=20.0 remaining=2 spawns=398 kills=10 attempts=10 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_bigrock_int_bionic_breach · Bot Surge on · wave 7 · Wave Failed</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 16 | 44 | 0 | 6 | 2 | 0% |
| 10 | 183 | 451 | 10 | 1 | 2 | 0% |

Reason: game or probe failed wave 7: wave_failed

Failure snapshot:

```text
WAVEPROBE_DEBUG state=4 wave=7 expected=7 game=209.2 engine=1905.6 timescale=20.0 remaining=2 spawns=451 kills=10 attempts=10 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_coaltown_exp_catastrophic_course · Bot Surge off · wave 1 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 18 | 10 | 0 | 0 | 57 | 0% |
| 80 | 1628 | 877 | 62 | 1 | 57 | 0% |
| 161 | 3237 | 1743 | 125 | 0 | 57 | 0% |
| 251 | 5047 | 2719 | 196 | 1 | 57 | 0% |
| 332 | 6655 | 3583 | 258 | 1 | 57 | 0% |
| 412 | 8262 | 4449 | 320 | 3 | 57 | 0% |
| 493 | 9870 | 5314 | 383 | 1 | 57 | 0% |
| 573 | 11477 | 6179 | 446 | 1 | 57 | 0% |
| 653 | 13085 | 7045 | 509 | 2 | 57 | 0% |
| 744 | 14894 | 8017 | 579 | 0 | 57 | 0% |
| 824 | 16501 | 8883 | 642 | 1 | 57 | 0% |
| 900 | 18028 | 9705 | 702 | 1 | 57 | 0% |

Reason: wave exceeded wall-clock time limit: wave 1 still active after 15m0s (18027.6 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=1 expected=1 game=18052.9 engine=927.9 timescale=20.0 remaining=57 spawns=9705 kills=702 attempts=702 defender=32
WAVEPROBE_BOT client=30 userid=4 class=2 hp=125 timer=0 deadline=2.8 origin=-825,759,464 name=Sniper
WAVEPROBE_DEBUG_END alive=1 listed=1
```

</details>

<details><summary>mvm_coaltown_exp_catastrophic_course · Bot Surge on · wave 2 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 0 | 0 | 0 | 100 | 0% |
| 80 | 1629 | 0 | 0 | 0 | 100 | 0% |
| 161 | 3238 | 0 | 0 | 0 | 100 | 0% |
| 251 | 5048 | 0 | 0 | 0 | 100 | 0% |
| 332 | 6658 | 0 | 0 | 0 | 100 | 0% |
| 412 | 8267 | 0 | 0 | 0 | 100 | 0% |
| 493 | 9876 | 0 | 0 | 0 | 100 | 0% |
| 573 | 11484 | 0 | 0 | 0 | 100 | 0% |
| 654 | 13093 | 0 | 0 | 0 | 100 | 0% |
| 744 | 14903 | 0 | 0 | 0 | 100 | 0% |
| 825 | 16512 | 0 | 0 | 0 | 100 | 0% |
| 901 | 18041 | 0 | 0 | 0 | 100 | 0% |

Reason: wave exceeded wall-clock time limit: wave 2 still active after 15m0s (18040.6 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=2 expected=2 game=18066.9 engine=1854.5 timescale=20.0 remaining=100 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_condemned_b3_exp_trespasser · Bot Surge on · wave 1 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 9 | 19 | 0 | 5 | 1196 | 0% |
| 81 | 560 | 683 | 132 | 15 | 1196 | 0% |
| 163 | 988 | 1185 | 235 | 18 | 1196 | 0% |
| 235 | 1253 | 1475 | 326 | 14 | 1196 | 0% |
| 319 | 1492 | 1762 | 372 | 15 | 1196 | 0% |
| 403 | 1663 | 1997 | 418 | 8 | 1196 | 0% |
| 488 | 1812 | 2177 | 483 | 6 | 1196 | 0% |
| 573 | 1977 | 2370 | 510 | 5 | 1196 | 0% |
| 658 | 2046 | 2449 | 546 | 17 | 1196 | 0% |
| 732 | 2111 | 2521 | 580 | 11 | 1196 | 0% |
| 816 | 2142 | 2568 | 587 | 12 | 1196 | 0% |
| 900 | 2176 | 2626 | 595 | 16 | 1196 | 0% |

Reason: wave exceeded wall-clock time limit: wave 1 still active after 15m0s (2176.0 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=1 expected=1 game=2201.3 engine=926.9 timescale=20.0 remaining=1196 spawns=2626 kills=595 attempts=595 defender=32
WAVEPROBE_BOT client=6 userid=28 class=1 hp=125 timer=0 deadline=14.3 origin=-906,356,-92 name=Zombie Scout
WAVEPROBE_BOT client=7 userid=27 class=1 hp=125 timer=0 deadline=20.0 origin=-893,602,-113 name=Zombie Scout
WAVEPROBE_BOT client=9 userid=25 class=7 hp=175 timer=0 deadline=15.6 origin=-972,1984,248 name=Burning Zombie
WAVEPROBE_BOT client=10 userid=24 class=7 hp=175 timer=0 deadline=10.9 origin=-1141,1514,244 name=Burning Zombie
WAVEPROBE_BOT client=11 userid=23 class=7 hp=175 timer=0 deadline=23.4 origin=-1746,2249,280 name=Burning Zombie
WAVEPROBE_BOT client=12 userid=22 class=2 hp=100 timer=0 deadline=14.5 origin=1319,632,21 name=Possessed
WAVEPROBE_BOT client=13 userid=21 class=7 hp=175 timer=0 deadline=11.8 origin=-968,1274,192 name=Burning Zombie
WAVEPROBE_BOT client=14 userid=20 class=1 hp=125 timer=0 deadline=18.7 origin=-1423,488,30 name=Zombie Scout
WAVEPROBE_BOT client=16 userid=18 class=7 hp=175 timer=0 deadline=15.6 origin=-1210,1934,255 name=Burning Zombie
WAVEPROBE_BOT client=17 userid=17 class=7 hp=175 timer=0 deadline=11.4 origin=-1190,1807,238 name=Burning Zombie
WAVEPROBE_BOT client=18 userid=16 class=2 hp=100 timer=0 deadline=17.3 origin=-1598,364,-15 name=Possessed
WAVEPROBE_BOT client=20 userid=14 class=2 hp=100 timer=0 deadline=22.5 origin=-1577,320,-15 name=Possessed
WAVEPROBE_DEBUG_END alive=15 listed=12
```

</details>

<details><summary>mvm_condemned_b3_exp_trespasser_remaster · Bot Surge off · wave 1 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 6 | 26 | 0 | 27 | 179 | 84% |
| 81 | 741 | 450 | 431 | 10 | 190 | 83% |
| 163 | 1158 | 657 | 571 | 8 | 190 | 83% |
| 245 | 1427 | 791 | 658 | 9 | 190 | 83% |
| 329 | 1631 | 893 | 734 | 10 | 190 | 83% |
| 413 | 1798 | 976 | 801 | 11 | 190 | 83% |
| 488 | 1906 | 1030 | 834 | 9 | 190 | 83% |
| 573 | 1996 | 1075 | 865 | 10 | 190 | 83% |
| 659 | 2070 | 1111 | 896 | 10 | 190 | 83% |
| 741 | 2142 | 1147 | 918 | 11 | 190 | 83% |
| 827 | 2211 | 1181 | 945 | 9 | 190 | 83% |
| 901 | 2273 | 1212 | 966 | 9 | 190 | 83% |

Reason: wave exceeded wall-clock time limit: wave 1 still active after 15m0s (2273.1 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=1 expected=1 game=2298.4 engine=1872.5 timescale=20.0 remaining=190 spawns=1212 kills=966 attempts=966 defender=32
WAVEPROBE_BOT client=9 userid=67 class=7 hp=100 timer=0 deadline=3.1 origin=14,752,-51 name=Headless Zombie
WAVEPROBE_BOT client=11 userid=65 class=6 hp=200 timer=0 deadline=16.1 origin=1310,697,17 name=Headless Zombie
WAVEPROBE_BOT client=13 userid=63 class=1 hp=100 timer=0 deadline=15.8 origin=-1227,-253,61 name=Headless Zombie
WAVEPROBE_BOT client=15 userid=61 class=5 hp=100 timer=0 deadline=10.6 origin=121,1526,192 name=Headless Zombie
WAVEPROBE_BOT client=20 userid=56 class=3 hp=100 timer=0 deadline=3.4 origin=57,844,-5 name=Headless Zombie
WAVEPROBE_BOT client=22 userid=54 class=3 hp=100 timer=0 deadline=11.0 origin=-734,-439,-97 name=Headless Zombie
WAVEPROBE_BOT client=25 userid=51 class=10 hp=100 timer=0 deadline=8.2 origin=-788,1906,197 name=Headless Zombie
WAVEPROBE_BOT client=27 userid=49 class=3 hp=40 timer=0 deadline=18.5 origin=453,1054,-63 name=Headless Zombie
WAVEPROBE_DEBUG_END alive=8 listed=8
```

</details>

<details><summary>mvm_condemned_b3_exp_trespasser_remaster · Bot Surge on · wave 1 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 6 | 21 | 0 | 22 | 179 | 84% |
| 81 | 883 | 548 | 532 | 7 | 190 | 83% |
| 162 | 1592 | 772 | 711 | 7 | 190 | 83% |
| 233 | 2002 | 903 | 812 | 7 | 190 | 83% |
| 316 | 2273 | 992 | 881 | 7 | 190 | 83% |
| 400 | 2454 | 1050 | 926 | 7 | 190 | 83% |
| 485 | 2590 | 1094 | 960 | 7 | 190 | 83% |
| 570 | 2713 | 1134 | 991 | 7 | 190 | 83% |
| 656 | 2832 | 1173 | 1021 | 7 | 190 | 83% |
| 732 | 2929 | 1205 | 1047 | 7 | 190 | 83% |
| 819 | 3034 | 1239 | 1075 | 7 | 190 | 83% |
| 901 | 3127 | 1270 | 1096 | 7 | 190 | 83% |

Reason: wave exceeded wall-clock time limit: wave 1 still active after 15m0s (3127.0 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=1 expected=1 game=3153.3 engine=3647.1 timescale=20.0 remaining=190 spawns=1270 kills=1096 attempts=1096 defender=32
WAVEPROBE_BOT client=17 userid=46 class=5 hp=100 timer=0 deadline=2.7 origin=-112,522,192 name=Headless Zombie
WAVEPROBE_BOT client=18 userid=45 class=1 hp=125 timer=0 deadline=6.3 origin=-1767,152,-15 name=Zombie Scout
WAVEPROBE_BOT client=20 userid=43 class=9 hp=100 timer=0 deadline=11.4 origin=562,990,196 name=Headless Zombie
WAVEPROBE_BOT client=26 userid=37 class=7 hp=100 timer=0 deadline=1.7 origin=-361,1071,192 name=Headless Zombie
WAVEPROBE_BOT client=28 userid=35 class=1 hp=125 timer=0 deadline=15.8 origin=-1517,300,-15 name=Zombie Scout
WAVEPROBE_BOT client=29 userid=34 class=1 hp=125 timer=0 deadline=11.8 origin=-1668,269,-15 name=Zombie Scout
WAVEPROBE_DEBUG_END alive=6 listed=6
```

</details>

<details><summary>mvm_creepside_b2_adv_catastrophic_conjuring · Bot Surge off · wave 1 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 4 | 0 | 0 | 149 | 0% |
| 80 | 1628 | 325 | 0 | 0 | 149 | 0% |
| 161 | 3237 | 646 | 0 | 0 | 149 | 0% |
| 251 | 5046 | 1008 | 0 | 1 | 149 | 0% |
| 332 | 6653 | 1328 | 0 | 0 | 149 | 0% |
| 412 | 8262 | 1649 | 0 | 0 | 149 | 0% |
| 493 | 9870 | 1970 | 0 | 0 | 149 | 0% |
| 573 | 11478 | 2291 | 0 | 0 | 149 | 0% |
| 653 | 13087 | 2612 | 0 | 0 | 149 | 0% |
| 744 | 14897 | 2974 | 0 | 0 | 149 | 0% |
| 824 | 16506 | 3295 | 0 | 0 | 149 | 0% |
| 901 | 18033 | 3600 | 0 | 0 | 149 | 0% |

Reason: wave exceeded wall-clock time limit: wave 1 still active after 15m0s (18033.2 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=1 expected=1 game=18058.3 engine=928.8 timescale=20.0 remaining=149 spawns=3600 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_creepside_b2_adv_dismal_devilry · Bot Surge off · wave 6 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 15 | 0 | 4 | 99 | 0% |
| 80 | 1576 | 161 | 15 | 0 | 99 | 0% |
| 161 | 3184 | 161 | 15 | 0 | 99 | 0% |
| 251 | 4984 | 161 | 15 | 0 | 99 | 0% |
| 332 | 6592 | 161 | 15 | 0 | 99 | 0% |
| 412 | 8184 | 161 | 15 | 0 | 99 | 0% |
| 492 | 9784 | 161 | 15 | 0 | 99 | 0% |
| 573 | 11392 | 161 | 15 | 0 | 99 | 0% |
| 653 | 13000 | 161 | 15 | 0 | 99 | 0% |
| 744 | 14800 | 161 | 15 | 0 | 99 | 0% |
| 824 | 16409 | 161 | 15 | 0 | 99 | 0% |
| 901 | 17938 | 161 | 15 | 0 | 99 | 0% |

Reason: wave exceeded wall-clock time limit: wave 6 still active after 15m0s (17937.8 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=6 expected=6 game=17964.0 engine=1914.4 timescale=20.0 remaining=99 spawns=161 kills=15 attempts=38 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_creepside_b2_adv_dismal_devilry · Bot Surge on · wave 6 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 8 | 28 | 0 | 7 | 99 | 0% |
| 81 | 1449 | 190 | 1 | 0 | 99 | 0% |
| 161 | 3013 | 190 | 1 | 0 | 99 | 0% |
| 252 | 4823 | 190 | 1 | 0 | 99 | 0% |
| 332 | 6429 | 190 | 1 | 0 | 99 | 0% |
| 412 | 8017 | 190 | 1 | 0 | 99 | 0% |
| 493 | 9574 | 190 | 1 | 0 | 99 | 0% |
| 573 | 11138 | 190 | 1 | 0 | 99 | 0% |
| 654 | 12740 | 190 | 1 | 0 | 99 | 0% |
| 744 | 14519 | 190 | 1 | 0 | 99 | 0% |
| 825 | 16127 | 190 | 1 | 0 | 99 | 0% |
| 900 | 17634 | 190 | 1 | 0 | 99 | 0% |

Reason: wave exceeded wall-clock time limit: wave 6 still active after 15m0s (17634.0 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=6 expected=6 game=17660.4 engine=3654.5 timescale=20.0 remaining=99 spawns=190 kills=1 attempts=22 defender=21
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_creepside_b2_adv_dust_to_dust · Bot Surge on · wave 1 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 0 | 0 | 0 | 127 | 0% |
| 80 | 1629 | 25 | 25 | 0 | 127 | 0% |
| 161 | 3236 | 50 | 49 | 1 | 127 | 0% |
| 251 | 5046 | 78 | 77 | 1 | 127 | 0% |
| 332 | 6654 | 102 | 102 | 0 | 127 | 0% |
| 412 | 8262 | 127 | 127 | 0 | 127 | 0% |
| 493 | 9871 | 152 | 151 | 1 | 127 | 0% |
| 573 | 11478 | 176 | 176 | 0 | 127 | 0% |
| 653 | 13087 | 201 | 201 | 0 | 127 | 0% |
| 744 | 14897 | 229 | 228 | 1 | 127 | 0% |
| 824 | 16505 | 254 | 253 | 1 | 127 | 0% |
| 901 | 18033 | 277 | 277 | 0 | 127 | 0% |

Reason: wave exceeded wall-clock time limit: wave 1 still active after 15m0s (18032.8 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=1 expected=1 game=18058.0 engine=2814.3 timescale=20.0 remaining=127 spawns=277 kills=277 attempts=277 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_creepside_b2_adv_dust_to_dust · Bot Surge off · wave 2 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 21 | 0 | 0 | 0 | 98 | 0% |
| 80 | 1629 | 34 | 34 | 0 | 98 | 0% |
| 161 | 3238 | 66 | 66 | 0 | 98 | 0% |
| 251 | 5046 | 104 | 104 | 0 | 98 | 0% |
| 332 | 6654 | 138 | 136 | 2 | 98 | 0% |
| 412 | 8264 | 170 | 170 | 0 | 98 | 0% |
| 493 | 9872 | 204 | 204 | 0 | 98 | 0% |
| 573 | 11481 | 236 | 236 | 0 | 98 | 0% |
| 653 | 13089 | 270 | 270 | 0 | 98 | 0% |
| 744 | 14899 | 308 | 306 | 2 | 98 | 0% |
| 824 | 16507 | 340 | 340 | 0 | 98 | 0% |
| 901 | 18034 | 372 | 372 | 0 | 98 | 0% |

Reason: wave exceeded wall-clock time limit: wave 2 still active after 15m0s (18033.8 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=2 expected=2 game=18060.1 engine=4585.3 timescale=20.0 remaining=98 spawns=372 kills=372 attempts=372 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_creepside_b2_adv_dust_to_dust · Bot Surge off · wave 5 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 2 | 0 | 1 | 74 | 0% |
| 81 | 722 | 689 | 15 | 6 | 74 | 0% |
| 164 | 1142 | 1266 | 17 | 8 | 74 | 0% |
| 246 | 1426 | 1651 | 18 | 7 | 74 | 0% |
| 329 | 1709 | 2032 | 19 | 10 | 74 | 0% |
| 411 | 2067 | 2517 | 24 | 7 | 74 | 0% |
| 483 | 2292 | 2825 | 26 | 10 | 74 | 0% |
| 566 | 2538 | 3160 | 26 | 14 | 74 | 0% |
| 650 | 2762 | 3463 | 29 | 12 | 74 | 0% |
| 733 | 2980 | 3760 | 29 | 6 | 74 | 0% |
| 817 | 3178 | 4037 | 30 | 13 | 74 | 0% |
| 901 | 3384 | 4315 | 30 | 7 | 74 | 0% |

Reason: wave exceeded wall-clock time limit: wave 5 still active after 15m0s (3384.3 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=5 expected=5 game=21449.4 engine=3716.5 timescale=20.0 remaining=74 spawns=4317 kills=30 attempts=30 defender=32
WAVEPROBE_BOT client=8 userid=71 class=4 hp=143 timer=0 deadline=8.2 origin=992,7302,498 name=Demoman
WAVEPROBE_BOT client=10 userid=69 class=4 hp=175 timer=0 deadline=18.8 origin=187,8012,672 name=Demoman
WAVEPROBE_BOT client=15 userid=64 class=4 hp=162 timer=0 deadline=13.0 origin=1508,7495,471 name=Demoman
WAVEPROBE_BOT client=18 userid=61 class=5 hp=150 timer=0 deadline=12.2 origin=153,7689,672 name=Big Heal Medic
WAVEPROBE_BOT client=23 userid=56 class=4 hp=175 timer=0 deadline=24.0 origin=187,8012,691 name=Demoman
WAVEPROBE_BOT client=24 userid=55 class=6 hp=375 timer=0 deadline=13.9 origin=411,7599,672 name=Shotgun Heavy
WAVEPROBE_BOT client=26 userid=53 class=5 hp=150 timer=0 deadline=12.6 origin=746,7390,489 name=Big Heal Medic
WAVEPROBE_BOT client=31 userid=51 class=6 hp=377 timer=0 deadline=17.3 origin=974,6987,543 name=Shotgun Heavy
WAVEPROBE_DEBUG_END alive=8 listed=8
```

</details>

<details><summary>mvm_decoy_intermediate · Bot Surge on · wave 3 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 0 | 0 | 0 | 108 | 0% |
| 80 | 1629 | 74 | 74 | 0 | 108 | 0% |
| 161 | 3237 | 150 | 150 | 0 | 108 | 0% |
| 251 | 5047 | 236 | 236 | 0 | 108 | 0% |
| 332 | 6654 | 314 | 312 | 2 | 108 | 0% |
| 412 | 8263 | 390 | 388 | 2 | 108 | 0% |
| 493 | 9871 | 466 | 465 | 1 | 108 | 0% |
| 573 | 11480 | 542 | 542 | 0 | 108 | 0% |
| 653 | 13088 | 620 | 618 | 2 | 108 | 0% |
| 744 | 14897 | 704 | 704 | 0 | 108 | 0% |
| 824 | 16504 | 782 | 780 | 2 | 108 | 0% |
| 901 | 18033 | 854 | 852 | 2 | 108 | 0% |

Reason: wave exceeded wall-clock time limit: wave 3 still active after 15m0s (18032.9 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=3 expected=3 game=18057.2 engine=919.2 timescale=20.0 remaining=108 spawns=854 kills=852 attempts=852 defender=32
WAVEPROBE_BOT client=30 userid=4 class=8 hp=125 timer=0 deadline=3.8 origin=750,2350,768 name=Spy
WAVEPROBE_BOT client=31 userid=3 class=8 hp=125 timer=0 deadline=2.1 origin=-850,2350,768 name=Spy
WAVEPROBE_DEBUG_END alive=2 listed=2
```

</details>

<details><summary>mvm_decoy_intermediate · Bot Surge off · wave 4 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 19 | 0 | 0 | 0 | 84 | 0% |
| 80 | 1627 | 0 | 0 | 0 | 84 | 0% |
| 161 | 3236 | 0 | 0 | 0 | 84 | 0% |
| 251 | 5046 | 0 | 0 | 0 | 84 | 0% |
| 332 | 6655 | 0 | 0 | 0 | 84 | 0% |
| 412 | 8264 | 0 | 0 | 0 | 84 | 0% |
| 493 | 9874 | 0 | 0 | 0 | 84 | 0% |
| 573 | 11483 | 0 | 0 | 0 | 84 | 0% |
| 654 | 13092 | 0 | 0 | 0 | 84 | 0% |
| 744 | 14903 | 0 | 0 | 0 | 84 | 0% |
| 825 | 16510 | 0 | 0 | 0 | 84 | 0% |
| 900 | 18019 | 0 | 0 | 0 | 84 | 0% |

Reason: wave exceeded wall-clock time limit: wave 4 still active after 15m0s (18019.2 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=4 expected=4 game=18043.5 engine=978.1 timescale=20.0 remaining=84 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_frostwynd_rc1_int_wicked_wizardry · Bot Surge off · wave 3 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 21 | 0 | 0 | 0 | 63 | 0% |
| 80 | 1631 | 0 | 0 | 0 | 63 | 0% |
| 161 | 3239 | 0 | 0 | 0 | 63 | 0% |
| 251 | 5049 | 0 | 0 | 0 | 63 | 0% |
| 332 | 6658 | 0 | 0 | 0 | 63 | 0% |
| 412 | 8268 | 0 | 0 | 0 | 63 | 0% |
| 493 | 9876 | 0 | 0 | 0 | 63 | 0% |
| 573 | 11486 | 0 | 0 | 0 | 63 | 0% |
| 654 | 13095 | 0 | 0 | 0 | 63 | 0% |
| 744 | 14904 | 0 | 0 | 0 | 63 | 0% |
| 825 | 16513 | 0 | 0 | 0 | 63 | 0% |
| 901 | 18041 | 0 | 0 | 0 | 63 | 0% |

Reason: wave exceeded wall-clock time limit: wave 3 still active after 15m0s (18040.8 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=3 expected=3 game=18066.0 engine=930.6 timescale=20.0 remaining=63 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_frostwynd_rc1_int_wicked_wizardry · Bot Surge off · wave 6 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 21 | 0 | 0 | 0 | 75 | 0% |
| 80 | 1631 | 0 | 0 | 0 | 75 | 0% |
| 161 | 3239 | 0 | 0 | 0 | 75 | 0% |
| 251 | 5048 | 0 | 0 | 0 | 75 | 0% |
| 332 | 6657 | 0 | 0 | 0 | 75 | 0% |
| 412 | 8265 | 0 | 0 | 0 | 75 | 0% |
| 493 | 9874 | 0 | 0 | 0 | 75 | 0% |
| 573 | 11483 | 0 | 0 | 0 | 75 | 0% |
| 654 | 13093 | 0 | 0 | 0 | 75 | 0% |
| 744 | 14903 | 0 | 0 | 0 | 75 | 0% |
| 825 | 16512 | 0 | 0 | 0 | 75 | 0% |
| 901 | 18041 | 0 | 0 | 0 | 75 | 0% |

Reason: wave exceeded wall-clock time limit: wave 6 still active after 15m0s (18040.6 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=6 expected=6 game=18065.8 engine=930.1 timescale=20.0 remaining=75 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_ghost_town_adv_horrorsome_happenings · Bot Surge off · wave 7 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 21 | 4 | 0 | 0 | 132 | 0% |
| 80 | 1570 | 132 | 114 | 0 | 132 | 0% |
| 161 | 3178 | 132 | 114 | 0 | 132 | 0% |
| 251 | 4988 | 132 | 114 | 0 | 132 | 0% |
| 332 | 6597 | 132 | 114 | 0 | 132 | 0% |
| 412 | 8204 | 132 | 114 | 0 | 132 | 0% |
| 493 | 9813 | 132 | 114 | 0 | 132 | 0% |
| 573 | 11421 | 132 | 114 | 0 | 132 | 0% |
| 653 | 13030 | 132 | 114 | 0 | 132 | 0% |
| 744 | 14839 | 132 | 114 | 0 | 132 | 0% |
| 824 | 16446 | 132 | 114 | 0 | 132 | 0% |
| 901 | 17973 | 132 | 114 | 0 | 132 | 0% |

Reason: wave exceeded wall-clock time limit: wave 7 still active after 15m0s (17973.2 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=7 expected=7 game=17998.5 engine=927.9 timescale=20.0 remaining=132 spawns=132 kills=114 attempts=114 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_hideout_b3_adv_perilous_paradise · Bot Surge on · wave 2 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 22 | 0 | 0 | 0 | 95 | 0% |
| 80 | 1629 | 27 | 27 | 0 | 95 | 0% |
| 161 | 3237 | 54 | 54 | 0 | 95 | 0% |
| 251 | 5046 | 84 | 84 | 0 | 95 | 0% |
| 332 | 6655 | 111 | 111 | 0 | 95 | 0% |
| 412 | 8264 | 138 | 137 | 1 | 95 | 0% |
| 493 | 9872 | 164 | 164 | 0 | 95 | 0% |
| 573 | 11481 | 191 | 191 | 0 | 95 | 0% |
| 653 | 13089 | 218 | 218 | 0 | 95 | 0% |
| 744 | 14899 | 248 | 248 | 0 | 95 | 0% |
| 824 | 16508 | 275 | 274 | 1 | 95 | 0% |
| 901 | 18036 | 300 | 300 | 0 | 95 | 0% |

Reason: wave exceeded wall-clock time limit: wave 2 still active after 15m0s (18036.0 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=2 expected=2 game=18061.3 engine=927.8 timescale=20.0 remaining=95 spawns=300 kills=300 attempts=300 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_hideout_b3_adv_perilous_paradise · Bot Surge off · wave 4 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 19 | 0 | 0 | 0 | 62 | 0% |
| 80 | 1627 | 48 | 48 | 0 | 62 | 0% |
| 161 | 3234 | 96 | 96 | 0 | 62 | 0% |
| 251 | 5044 | 150 | 150 | 0 | 62 | 0% |
| 332 | 6652 | 198 | 198 | 0 | 62 | 0% |
| 412 | 8260 | 246 | 246 | 0 | 62 | 0% |
| 492 | 9869 | 294 | 294 | 0 | 62 | 0% |
| 573 | 11478 | 342 | 341 | 1 | 62 | 0% |
| 653 | 13087 | 390 | 388 | 2 | 62 | 0% |
| 744 | 14897 | 444 | 443 | 1 | 62 | 0% |
| 824 | 16504 | 492 | 490 | 2 | 62 | 0% |
| 901 | 18033 | 536 | 536 | 0 | 62 | 0% |

Reason: wave exceeded wall-clock time limit: wave 4 still active after 15m0s (18032.6 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=4 expected=4 game=18058.0 engine=2738.3 timescale=20.0 remaining=62 spawns=536 kills=536 attempts=536 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_hideout_b3_int_intermediate · Bot Surge off · wave 3 · Probe Error</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 12 | 6 | 0 | 4 | 82 | 0% |
| 61 | 887 | 44 | 22 | 0 | 82 | 0% |
| 121 | 2093 | 44 | 22 | 0 | 82 | 0% |
| 171 | 3098 | 44 | 22 | 0 | 82 | 0% |
| 232 | 4305 | 44 | 22 | 0 | 82 | 0% |
| 292 | 5512 | 44 | 22 | 0 | 82 | 0% |
| 352 | 6718 | 44 | 22 | 0 | 82 | 0% |
| 412 | 7923 | 44 | 22 | 0 | 82 | 0% |
| 473 | 9129 | 44 | 22 | 0 | 82 | 0% |
| 523 | 10134 | 44 | 22 | 0 | 82 | 0% |
| 583 | 11341 | 44 | 22 | 0 | 82 | 0% |
| 690 | 1 | 0 | 0 | 0 | 0 | unknown |

Reason: probe state reset during wave 3: state=idle pop=mvm_decoy_adv_grutesque_getaway gamewave=0

Failure snapshot:

```text
WAVEPROBE_DEBUG state=0 wave=0 expected=0 game=1.0 engine=16.0 timescale=1.0 remaining=0 spawns=0 kills=0 attempts=0 defender=0
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_kelly_rc1b_adv_mobocracy · Bot Surge off · wave 1 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 9 | 0 | 1 | 114 | 0% |
| 80 | 1628 | 41 | 33 | 0 | 114 | 0% |
| 161 | 3235 | 73 | 65 | 0 | 114 | 0% |
| 251 | 5043 | 109 | 101 | 0 | 114 | 0% |
| 332 | 6651 | 141 | 133 | 0 | 114 | 0% |
| 412 | 8259 | 173 | 165 | 0 | 114 | 0% |
| 492 | 9866 | 205 | 197 | 0 | 114 | 0% |
| 573 | 11474 | 238 | 229 | 1 | 114 | 0% |
| 653 | 13081 | 270 | 261 | 1 | 114 | 0% |
| 744 | 14890 | 306 | 297 | 1 | 114 | 0% |
| 824 | 16497 | 338 | 329 | 1 | 114 | 0% |
| 900 | 18025 | 368 | 360 | 0 | 114 | 0% |

Reason: wave exceeded wall-clock time limit: wave 1 still active after 15m0s (18024.6 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=1 expected=1 game=18050.0 engine=927.3 timescale=20.0 remaining=114 spawns=368 kills=360 attempts=360 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_kelly_rc1b_adv_mobocracy · Bot Surge on · wave 3 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 9 | 0 | 1 | 132 | 0% |
| 81 | 845 | 318 | 93 | 8 | 132 | 0% |
| 162 | 1741 | 663 | 179 | 9 | 132 | 0% |
| 243 | 3041 | 1151 | 298 | 5 | 132 | 0% |
| 324 | 4346 | 1638 | 418 | 0 | 132 | 0% |
| 404 | 5672 | 2133 | 536 | 5 | 132 | 0% |
| 495 | 7111 | 2688 | 687 | 5 | 132 | 0% |
| 576 | 7988 | 3022 | 779 | 5 | 132 | 0% |
| 658 | 8575 | 3240 | 832 | 9 | 132 | 0% |
| 740 | 8957 | 3386 | 874 | 5 | 132 | 0% |
| 825 | 9174 | 3466 | 895 | 4 | 132 | 0% |
| 901 | 9320 | 3523 | 910 | 8 | 132 | 0% |

Reason: wave exceeded wall-clock time limit: wave 3 still active after 15m0s (9320.0 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=3 expected=3 game=9345.4 engine=2746.4 timescale=20.0 remaining=132 spawns=3523 kills=910 attempts=910 defender=32
WAVEPROBE_BOT client=13 userid=26 class=2 hp=125 timer=0 deadline=14.0 origin=422,2143,-77 name=Canadian Sniper
WAVEPROBE_BOT client=14 userid=25 class=2 hp=35 timer=0 deadline=5.6 origin=333,1948,-90 name=Canadian Sniper
WAVEPROBE_BOT client=16 userid=23 class=2 hp=125 timer=0 deadline=9.2 origin=458,1858,-75 name=Canadian Sniper
WAVEPROBE_BOT client=18 userid=21 class=7 hp=175 timer=0 deadline=19.5 origin=-1224,3405,144 name=Pyro
WAVEPROBE_BOT client=19 userid=20 class=5 hp=150 timer=0 deadline=22.9 origin=-1202,3324,144 name=Kritzkrieg Medic
WAVEPROBE_BOT client=20 userid=19 class=6 hp=6500 timer=0 deadline=14.5 origin=-1216,3209,144 name=Giant Steel Gauntlet
WAVEPROBE_BOT client=21 userid=18 class=6 hp=6500 timer=0 deadline=17.5 origin=-1015,2861,196 name=Giant Steel Gauntlet
WAVEPROBE_BOT client=22 userid=17 class=5 hp=150 timer=0 deadline=15.8 origin=-1100,2946,181 name=Kritzkrieg Medic
WAVEPROBE_DEBUG_END alive=8 listed=8
```

</details>

<details><summary>mvm_kelly_rc1b_adv_ominous_outback · Bot Surge off · wave 6 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 19 | 0 | 0 | 0 | 78 | 0% |
| 80 | 1628 | 0 | 0 | 0 | 78 | 0% |
| 161 | 3237 | 0 | 0 | 0 | 78 | 0% |
| 251 | 5047 | 0 | 0 | 0 | 78 | 0% |
| 332 | 6655 | 0 | 0 | 0 | 78 | 0% |
| 412 | 8263 | 0 | 0 | 0 | 78 | 0% |
| 493 | 9872 | 0 | 0 | 0 | 78 | 0% |
| 573 | 11479 | 0 | 0 | 0 | 78 | 0% |
| 653 | 13087 | 0 | 0 | 0 | 78 | 0% |
| 744 | 14897 | 0 | 0 | 0 | 78 | 0% |
| 824 | 16506 | 0 | 0 | 0 | 78 | 0% |
| 901 | 18033 | 0 | 0 | 0 | 78 | 0% |

Reason: wave exceeded wall-clock time limit: wave 6 still active after 15m0s (18033.4 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=6 expected=6 game=18058.7 engine=925.6 timescale=20.0 remaining=78 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_mannhattan_adv_scorched_skies · Bot Surge off · wave 1 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 0 | 0 | 0 | 83 | 0% |
| 80 | 1628 | 94 | 38 | 2 | 83 | 0% |
| 161 | 3236 | 185 | 78 | 0 | 83 | 0% |
| 251 | 5045 | 290 | 122 | 0 | 83 | 0% |
| 332 | 6653 | 383 | 160 | 2 | 83 | 0% |
| 412 | 8261 | 475 | 200 | 0 | 83 | 0% |
| 493 | 9870 | 568 | 240 | 0 | 83 | 0% |
| 573 | 11478 | 662 | 278 | 2 | 83 | 0% |
| 653 | 13086 | 753 | 318 | 0 | 83 | 0% |
| 744 | 14896 | 858 | 362 | 0 | 83 | 0% |
| 824 | 16504 | 951 | 400 | 2 | 83 | 0% |
| 901 | 18033 | 1038 | 438 | 0 | 83 | 0% |

Reason: wave exceeded wall-clock time limit: wave 1 still active after 15m0s (18033.0 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=1 expected=1 game=18058.1 engine=925.6 timescale=20.0 remaining=83 spawns=1038 kills=438 attempts=438 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_mannworks · Bot Surge on · wave 5 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 12 | 28 | 0 | 14 | 53 | 0% |
| 81 | 1165 | 1535 | 769 | 14 | 53 | 0% |
| 161 | 2320 | 3059 | 1543 | 16 | 53 | 0% |
| 252 | 3615 | 4755 | 2408 | 15 | 53 | 0% |
| 333 | 4769 | 6271 | 3181 | 14 | 53 | 0% |
| 413 | 5925 | 7799 | 3958 | 13 | 53 | 0% |
| 494 | 7052 | 9279 | 4715 | 12 | 53 | 0% |
| 575 | 8196 | 10785 | 5482 | 11 | 53 | 0% |
| 656 | 9304 | 12249 | 6224 | 13 | 53 | 0% |
| 746 | 10553 | 13891 | 7064 | 11 | 53 | 0% |
| 827 | 11650 | 15347 | 7799 | 16 | 53 | 0% |
| 901 | 12650 | 16659 | 8468 | 13 | 53 | 0% |

Reason: wave exceeded wall-clock time limit: wave 5 still active after 15m0s (12650.4 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=5 expected=5 game=12675.8 engine=926.6 timescale=20.0 remaining=53 spawns=16659 kills=8468 attempts=8468 defender=32
WAVEPROBE_BOT client=11 userid=23 class=3 hp=200 timer=0 deadline=3.1 origin=2118,2611,557 name=Soldier
WAVEPROBE_BOT client=12 userid=22 class=3 hp=200 timer=0 deadline=4.2 origin=2046,2670,585 name=Soldier
WAVEPROBE_BOT client=13 userid=21 class=3 hp=200 timer=0 deadline=7.9 origin=1968,2683,634 name=Soldier
WAVEPROBE_BOT client=14 userid=20 class=3 hp=200 timer=0 deadline=6.7 origin=2224,2775,544 name=Soldier
WAVEPROBE_BOT client=15 userid=19 class=3 hp=200 timer=0 deadline=1.5 origin=1880,2600,675 name=Soldier
WAVEPROBE_BOT client=16 userid=18 class=3 hp=200 timer=0 deadline=5.4 origin=2042,2758,568 name=Soldier
WAVEPROBE_BOT client=19 userid=15 class=3 hp=200 timer=0 deadline=10.6 origin=2052,2569,614 name=Soldier
WAVEPROBE_BOT client=23 userid=11 class=3 hp=200 timer=0 deadline=17.2 origin=2182,2651,544 name=Soldier
WAVEPROBE_BOT client=25 userid=9 class=3 hp=200 timer=0 deadline=20.7 origin=2161,2792,544 name=Soldier
WAVEPROBE_BOT client=26 userid=8 class=3 hp=200 timer=0 deadline=9.0 origin=1968,2504,659 name=Soldier
WAVEPROBE_BOT client=27 userid=7 class=3 hp=200 timer=0 deadline=12.4 origin=2080,2800,544 name=Soldier
WAVEPROBE_BOT client=28 userid=6 class=3 hp=200 timer=0 deadline=16.0 origin=2261,2700,544 name=Soldier
WAVEPROBE_DEBUG_END alive=13 listed=12
```

</details>

<details><summary>mvm_mannworks_advanced · Bot Surge off · wave 2 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 0 | 0 | 0 | 72 | 0% |
| 80 | 1627 | 40 | 40 | 0 | 72 | 0% |
| 161 | 3235 | 80 | 79 | 1 | 72 | 0% |
| 251 | 5045 | 124 | 123 | 1 | 72 | 0% |
| 332 | 6654 | 162 | 162 | 0 | 72 | 0% |
| 412 | 8262 | 202 | 202 | 0 | 72 | 0% |
| 493 | 9870 | 242 | 240 | 2 | 72 | 0% |
| 573 | 11478 | 280 | 280 | 0 | 72 | 0% |
| 653 | 13087 | 320 | 319 | 1 | 72 | 0% |
| 744 | 14895 | 364 | 364 | 0 | 72 | 0% |
| 824 | 16503 | 402 | 402 | 0 | 72 | 0% |
| 901 | 18031 | 440 | 440 | 0 | 72 | 0% |

Reason: wave exceeded wall-clock time limit: wave 2 still active after 15m0s (18030.8 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=2 expected=2 game=18057.2 engine=2746.5 timescale=20.0 remaining=72 spawns=440 kills=440 attempts=440 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_mannworks_advanced · Bot Surge on · wave 2 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 22 | 2 | 0 | 2 | 72 | 0% |
| 80 | 1630 | 40 | 40 | 0 | 72 | 0% |
| 161 | 3239 | 80 | 80 | 0 | 72 | 0% |
| 251 | 5048 | 124 | 124 | 0 | 72 | 0% |
| 332 | 6656 | 162 | 162 | 0 | 72 | 0% |
| 412 | 8264 | 202 | 202 | 0 | 72 | 0% |
| 493 | 9874 | 242 | 240 | 2 | 72 | 0% |
| 573 | 11483 | 280 | 280 | 0 | 72 | 0% |
| 653 | 13092 | 320 | 320 | 0 | 72 | 0% |
| 744 | 14902 | 364 | 364 | 0 | 72 | 0% |
| 824 | 16510 | 402 | 402 | 0 | 72 | 0% |
| 901 | 18039 | 440 | 440 | 0 | 72 | 0% |

Reason: wave exceeded wall-clock time limit: wave 2 still active after 15m0s (18038.7 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=2 expected=2 game=18063.8 engine=3660.4 timescale=20.0 remaining=72 spawns=440 kills=440 attempts=440 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_mannworks_intermediate2 · Bot Surge on · wave 1 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 0 | 0 | 0 | 80 | 0% |
| 80 | 1628 | 0 | 0 | 0 | 80 | 0% |
| 161 | 3237 | 0 | 0 | 0 | 80 | 0% |
| 251 | 5047 | 0 | 0 | 0 | 80 | 0% |
| 332 | 6657 | 0 | 0 | 0 | 80 | 0% |
| 412 | 8266 | 0 | 0 | 0 | 80 | 0% |
| 493 | 9875 | 0 | 0 | 0 | 80 | 0% |
| 573 | 11484 | 0 | 0 | 0 | 80 | 0% |
| 654 | 13092 | 0 | 0 | 0 | 80 | 0% |
| 744 | 14902 | 0 | 0 | 0 | 80 | 0% |
| 825 | 16511 | 0 | 0 | 0 | 80 | 0% |
| 901 | 18040 | 0 | 0 | 0 | 80 | 0% |

Reason: wave exceeded wall-clock time limit: wave 1 still active after 15m0s (18039.6 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=1 expected=1 game=18065.0 engine=2748.2 timescale=20.0 remaining=80 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_null_b9c_exp_0 · Bot Surge off · wave 3 · Wave Failed</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 17 | 0 | 18 | 234 | 0% |
| 10 | 190 | 198 | 37 | 18 | 234 | 0% |
| 14 | 253 | 244 | 46 | 1 | 234 | 0% |

Reason: game or probe failed wave 3: wave_failed

Failure snapshot:

```text
WAVEPROBE_DEBUG state=4 wave=3 expected=3 game=278.1 engine=55.5 timescale=20.0 remaining=234 spawns=244 kills=46 attempts=46 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_null_b9c_exp_0 · Bot Surge on · wave 3 · Wave Failed</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 12 | 22 | 0 | 23 | 234 | 0% |
| 10 | 107 | 125 | 29 | 17 | 234 | 0% |
| 20 | 192 | 244 | 37 | 9 | 234 | 0% |
| 23 | 223 | 264 | 41 | 1 | 234 | 0% |

Reason: game or probe failed wave 3: wave_failed

Failure snapshot:

```text
WAVEPROBE_DEBUG state=4 wave=3 expected=3 game=248.0 engine=1868.6 timescale=20.0 remaining=234 spawns=264 kills=41 attempts=41 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_oilrig_rc5d_adv_rescindment · Bot Surge off · wave 1 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 19 | 0 | 0 | 0 | 118 | 0% |
| 80 | 1628 | 0 | 0 | 0 | 118 | 0% |
| 161 | 3236 | 0 | 0 | 0 | 118 | 0% |
| 251 | 5047 | 0 | 0 | 0 | 118 | 0% |
| 332 | 6656 | 0 | 0 | 0 | 118 | 0% |
| 412 | 8264 | 0 | 0 | 0 | 118 | 0% |
| 493 | 9872 | 0 | 0 | 0 | 118 | 0% |
| 573 | 11481 | 0 | 0 | 0 | 118 | 0% |
| 654 | 13089 | 0 | 0 | 0 | 118 | 0% |
| 744 | 14901 | 0 | 0 | 0 | 118 | 0% |
| 825 | 16510 | 0 | 0 | 0 | 118 | 0% |
| 901 | 18039 | 0 | 0 | 0 | 118 | 0% |

Reason: wave exceeded wall-clock time limit: wave 1 still active after 15m0s (18038.7 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=1 expected=1 game=18064.0 engine=964.0 timescale=20.0 remaining=118 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_oilrig_rc5d_adv_rescindment · Bot Surge on · wave 1 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 19 | 0 | 0 | 0 | 118 | 0% |
| 80 | 1627 | 0 | 0 | 0 | 118 | 0% |
| 161 | 3235 | 0 | 0 | 0 | 118 | 0% |
| 251 | 5045 | 0 | 0 | 0 | 118 | 0% |
| 332 | 6653 | 0 | 0 | 0 | 118 | 0% |
| 412 | 8262 | 0 | 0 | 0 | 118 | 0% |
| 493 | 9871 | 0 | 0 | 0 | 118 | 0% |
| 573 | 11479 | 0 | 0 | 0 | 118 | 0% |
| 653 | 13088 | 0 | 0 | 0 | 118 | 0% |
| 744 | 14899 | 0 | 0 | 0 | 118 | 0% |
| 825 | 16509 | 0 | 0 | 0 | 118 | 0% |
| 901 | 18038 | 0 | 0 | 0 | 118 | 0% |

Reason: wave exceeded wall-clock time limit: wave 1 still active after 15m0s (18037.5 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=1 expected=1 game=18063.6 engine=2747.1 timescale=20.0 remaining=118 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_oilrig_rc5d_adv_rescindment · Bot Surge off · wave 2 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 19 | 0 | 0 | 0 | 82 | 0% |
| 80 | 1627 | 0 | 0 | 0 | 82 | 0% |
| 161 | 3236 | 0 | 0 | 0 | 82 | 0% |
| 251 | 5045 | 0 | 0 | 0 | 82 | 0% |
| 332 | 6654 | 0 | 0 | 0 | 82 | 0% |
| 412 | 8262 | 0 | 0 | 0 | 82 | 0% |
| 493 | 9870 | 0 | 0 | 0 | 82 | 0% |
| 573 | 11479 | 0 | 0 | 0 | 82 | 0% |
| 653 | 13088 | 0 | 0 | 0 | 82 | 0% |
| 744 | 14899 | 0 | 0 | 0 | 82 | 0% |
| 824 | 16508 | 0 | 0 | 0 | 82 | 0% |
| 901 | 18036 | 0 | 0 | 0 | 82 | 0% |

Reason: wave exceeded wall-clock time limit: wave 2 still active after 15m0s (18036.5 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=2 expected=2 game=18062.8 engine=3668.8 timescale=20.0 remaining=82 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_oilrig_rc5d_adv_rescindment · Bot Surge on · wave 2 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 19 | 0 | 0 | 0 | 82 | 0% |
| 80 | 1628 | 0 | 0 | 0 | 82 | 0% |
| 161 | 3236 | 0 | 0 | 0 | 82 | 0% |
| 251 | 5044 | 0 | 0 | 0 | 82 | 0% |
| 332 | 6653 | 0 | 0 | 0 | 82 | 0% |
| 412 | 8262 | 0 | 0 | 0 | 82 | 0% |
| 493 | 9871 | 0 | 0 | 0 | 82 | 0% |
| 573 | 11480 | 0 | 0 | 0 | 82 | 0% |
| 653 | 13089 | 0 | 0 | 0 | 82 | 0% |
| 744 | 14899 | 0 | 0 | 0 | 82 | 0% |
| 824 | 16508 | 0 | 0 | 0 | 82 | 0% |
| 901 | 18036 | 0 | 0 | 0 | 82 | 0% |

Reason: wave exceeded wall-clock time limit: wave 2 still active after 15m0s (18035.6 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=2 expected=2 game=18060.9 engine=925.8 timescale=20.0 remaining=82 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_oilrig_rc5d_adv_rescindment · Bot Surge off · wave 3 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 0 | 0 | 0 | 110 | 0% |
| 80 | 1628 | 0 | 0 | 0 | 110 | 0% |
| 161 | 3237 | 0 | 0 | 0 | 110 | 0% |
| 251 | 5047 | 0 | 0 | 0 | 110 | 0% |
| 332 | 6656 | 0 | 0 | 0 | 110 | 0% |
| 412 | 8263 | 0 | 0 | 0 | 110 | 0% |
| 493 | 9872 | 0 | 0 | 0 | 110 | 0% |
| 573 | 11479 | 0 | 0 | 0 | 110 | 0% |
| 653 | 13087 | 0 | 0 | 0 | 110 | 0% |
| 744 | 14898 | 0 | 0 | 0 | 110 | 0% |
| 824 | 16506 | 0 | 0 | 0 | 110 | 0% |
| 901 | 18034 | 0 | 0 | 0 | 110 | 0% |

Reason: wave exceeded wall-clock time limit: wave 3 still active after 15m0s (18034.1 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=3 expected=3 game=18060.5 engine=4596.7 timescale=20.0 remaining=110 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_oilrig_rc5d_adv_rescindment · Bot Surge off · wave 5 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 19 | 0 | 0 | 0 | 118 | 0% |
| 80 | 1628 | 0 | 0 | 0 | 118 | 0% |
| 161 | 3237 | 0 | 0 | 0 | 118 | 0% |
| 251 | 5047 | 0 | 0 | 0 | 118 | 0% |
| 332 | 6656 | 0 | 0 | 0 | 118 | 0% |
| 412 | 8265 | 0 | 0 | 0 | 118 | 0% |
| 493 | 9873 | 0 | 0 | 0 | 118 | 0% |
| 573 | 11482 | 0 | 0 | 0 | 118 | 0% |
| 654 | 13090 | 0 | 0 | 0 | 118 | 0% |
| 744 | 14902 | 0 | 0 | 0 | 118 | 0% |
| 825 | 16511 | 0 | 0 | 0 | 118 | 0% |
| 901 | 18039 | 0 | 0 | 0 | 118 | 0% |

Reason: wave exceeded wall-clock time limit: wave 5 still active after 15m0s (18038.6 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=5 expected=5 game=18065.0 engine=1922.7 timescale=20.0 remaining=118 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_oilrig_rc5d_adv_rescindment · Bot Surge off · wave 6 · Probe Error</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 21 | 0 | 0 | 0 | 119 | 0% |
| 40 | 825 | 0 | 0 | 0 | 119 | 0% |
| 70 | 1428 | 0 | 0 | 0 | 119 | 0% |
| 111 | 2232 | 0 | 0 | 0 | 119 | 0% |
| 151 | 3037 | 0 | 0 | 0 | 119 | 0% |
| 181 | 3639 | 0 | 0 | 0 | 119 | 0% |
| 221 | 4443 | 0 | 0 | 0 | 119 | 0% |
| 251 | 5047 | 0 | 0 | 0 | 119 | 0% |
| 292 | 5851 | 0 | 0 | 0 | 119 | 0% |
| 332 | 6655 | 0 | 0 | 0 | 119 | 0% |
| 362 | 7259 | 0 | 0 | 0 | 119 | 0% |
| 449 | 1 | 0 | 0 | 0 | 0 | unknown |

Reason: probe state reset during wave 6: state=idle pop=mvm_decoy_adv_grutesque_getaway gamewave=0

Failure snapshot:

```text
WAVEPROBE_DEBUG state=0 wave=0 expected=0 game=1.0 engine=21.0 timescale=1.0 remaining=0 spawns=0 kills=0 attempts=0 defender=0
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_oilrig_rc5d_int_orson_oligarchy · Bot Surge on · wave 4 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 0 | 0 | 0 | 60 | 0% |
| 80 | 1627 | 0 | 0 | 0 | 60 | 0% |
| 161 | 3236 | 0 | 0 | 0 | 60 | 0% |
| 251 | 5044 | 0 | 0 | 0 | 60 | 0% |
| 332 | 6652 | 0 | 0 | 0 | 60 | 0% |
| 412 | 8260 | 0 | 0 | 0 | 60 | 0% |
| 492 | 9867 | 0 | 0 | 0 | 60 | 0% |
| 573 | 11476 | 0 | 0 | 0 | 60 | 0% |
| 653 | 13085 | 0 | 0 | 0 | 60 | 0% |
| 744 | 14894 | 0 | 0 | 0 | 60 | 0% |
| 824 | 16502 | 0 | 0 | 0 | 60 | 0% |
| 901 | 18031 | 0 | 0 | 0 | 60 | 0% |

Reason: wave exceeded wall-clock time limit: wave 4 still active after 15m0s (18030.6 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=4 expected=4 game=18057.0 engine=5969.3 timescale=20.0 remaining=60 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_oxidize_rc3_int_snowy_slaughter · Bot Surge off · wave 2 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 19 | 70 | 0 | 1 | 42 | 0% |
| 80 | 1626 | 6532 | 19 | 0 | 42 | 0% |
| 161 | 3236 | 13040 | 27 | 3 | 42 | 0% |
| 251 | 5045 | 20345 | 41 | 3 | 42 | 0% |
| 332 | 6648 | 26805 | 55 | 2 | 42 | 0% |
| 412 | 8253 | 33305 | 61 | 2 | 42 | 0% |
| 493 | 9809 | 39573 | 77 | 0 | 42 | 0% |
| 573 | 11339 | 45745 | 89 | 0 | 42 | 0% |
| 654 | 12947 | 52244 | 99 | 1 | 42 | 0% |
| 744 | 14756 | 59538 | 114 | 2 | 42 | 0% |
| 825 | 16338 | 65912 | 129 | 0 | 42 | 0% |
| 900 | 17704 | 71406 | 145 | 0 | 42 | 0% |

Reason: wave exceeded wall-clock time limit: wave 2 still active after 15m0s (17704.4 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=2 expected=2 game=17730.7 engine=1828.1 timescale=20.0 remaining=42 spawns=71407 kills=145 attempts=145 defender=32
WAVEPROBE_BOT client=15 userid=22 class=1 hp=125 timer=0 deadline=19.6 origin=-1987,4475,0 name=Scout
WAVEPROBE_DEBUG_END alive=1 listed=1
```

</details>

<details><summary>mvm_radar_b10_adv_rocky_ravage · Bot Surge on · wave 2 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 18 | 2 | 1 | 0 | 128 | 0% |
| 80 | 1627 | 49 | 43 | 0 | 128 | 0% |
| 161 | 3235 | 90 | 83 | 1 | 128 | 0% |
| 251 | 5043 | 135 | 128 | 1 | 128 | 0% |
| 332 | 6652 | 175 | 168 | 1 | 128 | 0% |
| 412 | 8260 | 215 | 209 | 0 | 128 | 0% |
| 493 | 9869 | 255 | 249 | 0 | 128 | 0% |
| 573 | 11477 | 295 | 289 | 0 | 128 | 0% |
| 653 | 13085 | 335 | 329 | 0 | 128 | 0% |
| 744 | 14894 | 381 | 374 | 1 | 128 | 0% |
| 824 | 16502 | 421 | 414 | 1 | 128 | 0% |
| 901 | 18029 | 459 | 452 | 1 | 128 | 0% |

Reason: wave exceeded wall-clock time limit: wave 2 still active after 15m0s (18028.9 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=2 expected=2 game=18055.2 engine=1837.8 timescale=20.0 remaining=128 spawns=459 kills=452 attempts=452 defender=32
WAVEPROBE_BOT client=31 userid=10 class=2 hp=125 timer=0 deadline=4.8 origin=2156,1163,-218 name=Sniper
WAVEPROBE_DEBUG_END alive=1 listed=1
```

</details>

<details><summary>mvm_radar_b10_adv_rocky_ravage · Bot Surge off · wave 6 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 21 | 0 | 0 | 0 | 84 | 0% |
| 80 | 1630 | 0 | 0 | 0 | 84 | 0% |
| 161 | 3238 | 0 | 0 | 0 | 84 | 0% |
| 251 | 5048 | 0 | 0 | 0 | 84 | 0% |
| 332 | 6657 | 0 | 0 | 0 | 84 | 0% |
| 412 | 8265 | 0 | 0 | 0 | 84 | 0% |
| 493 | 9873 | 0 | 0 | 0 | 84 | 0% |
| 573 | 11483 | 0 | 0 | 0 | 84 | 0% |
| 654 | 13092 | 0 | 0 | 0 | 84 | 0% |
| 744 | 14901 | 0 | 0 | 0 | 84 | 0% |
| 824 | 16510 | 0 | 0 | 0 | 84 | 0% |
| 901 | 18038 | 0 | 0 | 0 | 84 | 0% |

Reason: wave exceeded wall-clock time limit: wave 6 still active after 15m0s (18038.5 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=6 expected=6 game=36100.5 engine=2741.4 timescale=20.0 remaining=84 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_radar_b10_adv_rocky_ravage · Bot Surge on · wave 6 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 22 | 0 | 0 | 0 | 84 | 0% |
| 80 | 1630 | 0 | 0 | 0 | 84 | 0% |
| 161 | 3238 | 0 | 0 | 0 | 84 | 0% |
| 251 | 5047 | 0 | 0 | 0 | 84 | 0% |
| 332 | 6657 | 0 | 0 | 0 | 84 | 0% |
| 412 | 8264 | 0 | 0 | 0 | 84 | 0% |
| 493 | 9873 | 0 | 0 | 0 | 84 | 0% |
| 573 | 11480 | 0 | 0 | 0 | 84 | 0% |
| 653 | 13089 | 0 | 0 | 0 | 84 | 0% |
| 744 | 14898 | 0 | 0 | 0 | 84 | 0% |
| 824 | 16508 | 0 | 0 | 0 | 84 | 0% |
| 901 | 18036 | 0 | 0 | 0 | 84 | 0% |

Reason: wave exceeded wall-clock time limit: wave 6 still active after 15m0s (18035.9 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=6 expected=6 game=18061.3 engine=3675.0 timescale=20.0 remaining=84 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_radar_b10_int_recruits_recognition · Bot Surge off · wave 2 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 2 | 8 | 0 | 1 | 93 | 0% |
| 85 | 229 | 930 | 1 | 0 | 93 | 0% |
| 159 | 405 | 1648 | 1 | 1 | 93 | 0% |
| 244 | 582 | 2364 | 1 | 0 | 93 | 0% |
| 329 | 741 | 3012 | 1 | 1 | 93 | 0% |
| 415 | 907 | 3684 | 1 | 1 | 93 | 0% |
| 490 | 1049 | 4263 | 1 | 0 | 93 | 0% |
| 575 | 1203 | 4890 | 1 | 0 | 93 | 0% |
| 660 | 1349 | 5482 | 1 | 1 | 93 | 0% |
| 746 | 1495 | 6076 | 1 | 1 | 93 | 0% |
| 822 | 1622 | 6590 | 1 | 1 | 93 | 0% |
| 901 | 1753 | 7125 | 1 | 1 | 93 | 0% |

Reason: wave exceeded wall-clock time limit: wave 2 still active after 15m0s (1753.4 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=2 expected=2 game=1778.8 engine=4558.3 timescale=20.0 remaining=93 spawns=7126 kills=1 attempts=1 defender=32
WAVEPROBE_BOT client=13 userid=73 class=3 hp=200 timer=0 deadline=20.7 origin=3596,2136,304 name=Extended Conch Soldier
WAVEPROBE_DEBUG_END alive=1 listed=1
```

</details>

<details><summary>mvm_redstone_ridge_rc5_adv_armored_apparatus · Bot Surge off · wave 5 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 0 | 0 | 0 | 151 | 0% |
| 80 | 1628 | 0 | 0 | 0 | 151 | 0% |
| 161 | 3237 | 0 | 0 | 0 | 151 | 0% |
| 251 | 5047 | 0 | 0 | 0 | 151 | 0% |
| 332 | 6657 | 0 | 0 | 0 | 151 | 0% |
| 412 | 8267 | 0 | 0 | 0 | 151 | 0% |
| 493 | 9876 | 0 | 0 | 0 | 151 | 0% |
| 573 | 11484 | 0 | 0 | 0 | 151 | 0% |
| 654 | 13092 | 0 | 0 | 0 | 151 | 0% |
| 744 | 14902 | 0 | 0 | 0 | 151 | 0% |
| 825 | 16511 | 0 | 0 | 0 | 151 | 0% |
| 901 | 18039 | 0 | 0 | 0 | 151 | 0% |

Reason: wave exceeded wall-clock time limit: wave 5 still active after 15m0s (18038.8 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=5 expected=5 game=18065.2 engine=1836.8 timescale=20.0 remaining=151 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_redstone_ridge_rc5_adv_armored_apparatus · Bot Surge on · wave 5 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 0 | 0 | 0 | 151 | 0% |
| 80 | 1629 | 0 | 0 | 0 | 151 | 0% |
| 161 | 3237 | 0 | 0 | 0 | 151 | 0% |
| 251 | 5047 | 0 | 0 | 0 | 151 | 0% |
| 332 | 6656 | 0 | 0 | 0 | 151 | 0% |
| 412 | 8266 | 0 | 0 | 0 | 151 | 0% |
| 493 | 9873 | 0 | 0 | 0 | 151 | 0% |
| 573 | 11482 | 0 | 0 | 0 | 151 | 0% |
| 653 | 13090 | 0 | 0 | 0 | 151 | 0% |
| 744 | 14900 | 0 | 0 | 0 | 151 | 0% |
| 824 | 16509 | 0 | 0 | 0 | 151 | 0% |
| 901 | 18038 | 0 | 0 | 0 | 151 | 0% |

Reason: wave exceeded wall-clock time limit: wave 5 still active after 15m0s (18037.9 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=5 expected=5 game=18063.2 engine=2750.6 timescale=20.0 remaining=151 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_rottenburg_adv_cybernetic_carnage · Bot Surge off · wave 4 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 1 | 0 | 1 | 73 | 0% |
| 80 | 1628 | 69 | 68 | 1 | 73 | 0% |
| 161 | 3237 | 136 | 136 | 0 | 73 | 0% |
| 251 | 5047 | 213 | 212 | 1 | 73 | 0% |
| 332 | 6656 | 281 | 280 | 1 | 73 | 0% |
| 412 | 8264 | 348 | 348 | 0 | 73 | 0% |
| 493 | 9874 | 417 | 416 | 1 | 73 | 0% |
| 573 | 11481 | 484 | 483 | 1 | 73 | 0% |
| 653 | 13089 | 552 | 551 | 1 | 73 | 0% |
| 744 | 14898 | 628 | 627 | 1 | 73 | 0% |
| 824 | 16505 | 695 | 695 | 0 | 73 | 0% |
| 901 | 18034 | 760 | 759 | 1 | 73 | 0% |

Reason: wave exceeded wall-clock time limit: wave 4 still active after 15m0s (18033.6 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=4 expected=4 game=18059.9 engine=1835.6 timescale=20.0 remaining=73 spawns=760 kills=759 attempts=759 defender=32
WAVEPROBE_BOT client=31 userid=5 class=8 hp=125 timer=0 deadline=11.5 origin=-1950,1712,-169 name=Spy
WAVEPROBE_DEBUG_END alive=1 listed=1
```

</details>

<details><summary>mvm_rottenburg_adv_marathon · Bot Surge off · wave 4 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 18 | 36 | 0 | 2 | 54 | 0% |
| 80 | 1624 | 133 | 84 | 2 | 54 | 0% |
| 161 | 3231 | 209 | 160 | 2 | 54 | 0% |
| 251 | 5039 | 295 | 246 | 2 | 54 | 0% |
| 332 | 6648 | 371 | 324 | 0 | 54 | 0% |
| 412 | 8256 | 447 | 400 | 0 | 54 | 0% |
| 492 | 9864 | 525 | 476 | 2 | 54 | 0% |
| 573 | 11472 | 601 | 552 | 2 | 54 | 0% |
| 653 | 13081 | 677 | 628 | 2 | 54 | 0% |
| 744 | 14890 | 763 | 714 | 2 | 54 | 0% |
| 824 | 16497 | 839 | 792 | 0 | 54 | 0% |
| 900 | 18025 | 911 | 864 | 0 | 54 | 0% |

Reason: wave exceeded wall-clock time limit: wave 4 still active after 15m0s (18024.8 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=4 expected=4 game=18050.2 engine=2765.0 timescale=20.0 remaining=54 spawns=911 kills=864 attempts=864 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_rottenburg_adv_marathon · Bot Surge on · wave 4 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 17 | 56 | 0 | 10 | 54 | 0% |
| 80 | 1627 | 133 | 84 | 2 | 54 | 0% |
| 161 | 3235 | 209 | 162 | 0 | 54 | 0% |
| 251 | 5044 | 295 | 247 | 1 | 54 | 0% |
| 332 | 6652 | 371 | 324 | 0 | 54 | 0% |
| 412 | 8259 | 447 | 400 | 0 | 54 | 0% |
| 492 | 9868 | 525 | 476 | 2 | 54 | 0% |
| 573 | 11475 | 601 | 552 | 2 | 54 | 0% |
| 653 | 13084 | 677 | 629 | 1 | 54 | 0% |
| 744 | 14893 | 763 | 715 | 1 | 54 | 0% |
| 824 | 16501 | 839 | 792 | 0 | 54 | 0% |
| 900 | 18028 | 913 | 864 | 2 | 54 | 0% |

Reason: wave exceeded wall-clock time limit: wave 4 still active after 15m0s (18028.1 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=4 expected=4 game=18054.4 engine=3687.2 timescale=20.0 remaining=54 spawns=913 kills=864 attempts=864 defender=32
WAVEPROBE_BOT client=10 userid=53 class=9 hp=500 timer=0 deadline=13.6 origin=401,-2571,-126 name=Rocketeer Engineer
WAVEPROBE_BOT client=17 userid=46 class=9 hp=500 timer=0 deadline=18.9 origin=442,-2563,-126 name=Rocketeer Engineer
WAVEPROBE_DEBUG_END alive=2 listed=2
```

</details>

<details><summary>mvm_sharp_rc9_adv_sudden_equinox · Bot Surge off · wave 4 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 0 | 0 | 0 | 79 | 0% |
| 80 | 1629 | 0 | 0 | 0 | 79 | 0% |
| 161 | 3239 | 0 | 0 | 0 | 79 | 0% |
| 251 | 5048 | 0 | 0 | 0 | 79 | 0% |
| 332 | 6657 | 0 | 0 | 0 | 79 | 0% |
| 412 | 8265 | 0 | 0 | 0 | 79 | 0% |
| 493 | 9874 | 0 | 0 | 0 | 79 | 0% |
| 573 | 11482 | 0 | 0 | 0 | 79 | 0% |
| 654 | 13092 | 0 | 0 | 0 | 79 | 0% |
| 744 | 14903 | 0 | 0 | 0 | 79 | 0% |
| 825 | 16512 | 0 | 0 | 0 | 79 | 0% |
| 901 | 18038 | 0 | 0 | 0 | 79 | 0% |

Reason: wave exceeded wall-clock time limit: wave 4 still active after 15m0s (18038.4 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=4 expected=4 game=18064.8 engine=1834.7 timescale=20.0 remaining=79 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_sharp_rc9_adv_sudden_equinox · Bot Surge on · wave 4 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 21 | 0 | 0 | 0 | 79 | 0% |
| 80 | 1630 | 0 | 0 | 0 | 79 | 0% |
| 161 | 3238 | 0 | 0 | 0 | 79 | 0% |
| 251 | 5049 | 0 | 0 | 0 | 79 | 0% |
| 332 | 6658 | 0 | 0 | 0 | 79 | 0% |
| 412 | 8267 | 0 | 0 | 0 | 79 | 0% |
| 493 | 9876 | 0 | 0 | 0 | 79 | 0% |
| 573 | 11485 | 0 | 0 | 0 | 79 | 0% |
| 654 | 13094 | 0 | 0 | 0 | 79 | 0% |
| 744 | 14904 | 0 | 0 | 0 | 79 | 0% |
| 825 | 16514 | 0 | 0 | 0 | 79 | 0% |
| 900 | 18023 | 0 | 0 | 0 | 79 | 0% |

Reason: wave exceeded wall-clock time limit: wave 4 still active after 15m0s (18022.8 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=4 expected=4 game=18049.2 engine=2778.0 timescale=20.0 remaining=79 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_sharp_rc9_exp_gilead_gehenom · Bot Surge off · wave 3 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 2 | 0 | 1 | 89 | 0% |
| 80 | 1629 | 17 | 16 | 0 | 89 | 0% |
| 161 | 3238 | 17 | 16 | 0 | 89 | 0% |
| 251 | 5047 | 17 | 16 | 0 | 89 | 0% |
| 332 | 6655 | 17 | 16 | 0 | 89 | 0% |
| 412 | 8265 | 17 | 16 | 0 | 89 | 0% |
| 493 | 9875 | 17 | 16 | 0 | 89 | 0% |
| 573 | 11484 | 17 | 16 | 0 | 89 | 0% |
| 654 | 13093 | 17 | 16 | 0 | 89 | 0% |
| 744 | 14902 | 17 | 16 | 0 | 89 | 0% |
| 825 | 16510 | 17 | 16 | 0 | 89 | 0% |
| 901 | 18039 | 17 | 16 | 0 | 89 | 0% |

Reason: wave exceeded wall-clock time limit: wave 3 still active after 15m0s (18038.8 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=3 expected=3 game=18065.2 engine=3694.1 timescale=20.0 remaining=89 spawns=17 kills=16 attempts=16 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_sludge_b6_int_logging_dreadwood · Bot Surge off · wave 3 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 21 | 16 | 0 | 0 | 130 | 0% |
| 81 | 1580 | 1528 | 520 | 11 | 130 | 0% |
| 161 | 3071 | 2985 | 1053 | 8 | 130 | 0% |
| 252 | 4807 | 4683 | 1674 | 10 | 130 | 0% |
| 332 | 6362 | 6199 | 2228 | 7 | 130 | 0% |
| 413 | 7897 | 7705 | 2776 | 10 | 130 | 0% |
| 493 | 9136 | 8921 | 3222 | 10 | 130 | 0% |
| 574 | 10398 | 10152 | 3670 | 13 | 130 | 0% |
| 655 | 11818 | 11539 | 4176 | 9 | 130 | 0% |
| 745 | 13551 | 13234 | 4796 | 9 | 130 | 0% |
| 826 | 14865 | 14521 | 5269 | 8 | 130 | 0% |
| 901 | 15696 | 15334 | 5565 | 10 | 130 | 0% |

Reason: wave exceeded wall-clock time limit: wave 3 still active after 15m0s (15695.7 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=3 expected=3 game=15722.0 engine=1836.0 timescale=20.0 remaining=130 spawns=15334 kills=5565 attempts=5565 defender=32
WAVEPROBE_BOT client=11 userid=46 class=7 hp=175 timer=0 deadline=16.2 origin=-1136,-300,992 name=Pyro
WAVEPROBE_BOT client=12 userid=45 class=7 hp=175 timer=0 deadline=23.8 origin=-1144,-310,992 name=Pyro
WAVEPROBE_BOT client=17 userid=40 class=7 hp=175 timer=0 deadline=22.1 origin=-1143,-309,992 name=Pyro
WAVEPROBE_BOT client=18 userid=39 class=7 hp=175 timer=0 deadline=21.4 origin=-1135,-309,992 name=Pyro
WAVEPROBE_BOT client=19 userid=38 class=7 hp=175 timer=0 deadline=18.6 origin=-1141,-309,992 name=Pyro
WAVEPROBE_BOT client=21 userid=36 class=7 hp=175 timer=0 deadline=16.7 origin=-1137,-298,992 name=Pyro
WAVEPROBE_BOT client=22 userid=35 class=8 hp=125 timer=0 deadline=14.5 origin=937,-3450,534 name=Spy
WAVEPROBE_BOT client=25 userid=32 class=7 hp=175 timer=0 deadline=15.0 origin=-1137,-317,992 name=Pyro
WAVEPROBE_BOT client=28 userid=29 class=7 hp=175 timer=0 deadline=20.2 origin=-1139,-308,992 name=Pyro
WAVEPROBE_BOT client=30 userid=27 class=8 hp=125 timer=0 deadline=19.3 origin=-712,-5012,832 name=Spy
WAVEPROBE_DEBUG_END alive=10 listed=10
```

</details>

<details><summary>mvm_sludge_b6_int_logging_dreadwood · Bot Surge on · wave 3 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 4 | 14 | 0 | 8 | 130 | 0% |
| 84 | 216 | 450 | 32 | 12 | 130 | 0% |
| 159 | 384 | 843 | 66 | 8 | 130 | 0% |
| 245 | 548 | 1209 | 100 | 11 | 130 | 0% |
| 330 | 723 | 1618 | 135 | 8 | 130 | 0% |
| 417 | 888 | 2018 | 170 | 0 | 130 | 0% |
| 492 | 1024 | 2372 | 183 | 10 | 130 | 0% |
| 578 | 1162 | 2710 | 217 | 7 | 130 | 0% |
| 663 | 1295 | 3021 | 239 | 8 | 130 | 0% |
| 749 | 1398 | 3234 | 261 | 12 | 130 | 0% |
| 826 | 1507 | 3504 | 278 | 10 | 130 | 0% |
| 901 | 1603 | 3722 | 303 | 8 | 130 | 0% |

Reason: wave exceeded wall-clock time limit: wave 3 still active after 15m0s (1603.3 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=3 expected=3 game=1628.6 engine=2785.2 timescale=20.0 remaining=130 spawns=3722 kills=303 attempts=303 defender=32
WAVEPROBE_BOT client=10 userid=93 class=7 hp=175 timer=0 deadline=16.7 origin=-1072,-349,992 name=Pyro
WAVEPROBE_BOT client=13 userid=90 class=7 hp=175 timer=0 deadline=16.0 origin=-987,-236,992 name=Pyro
WAVEPROBE_BOT client=15 userid=88 class=7 hp=175 timer=0 deadline=19.6 origin=-900,-391,992 name=Pyro
WAVEPROBE_BOT client=16 userid=87 class=7 hp=175 timer=0 deadline=13.1 origin=-943,-456,992 name=Pyro
WAVEPROBE_BOT client=18 userid=85 class=7 hp=175 timer=0 deadline=21.9 origin=-1056,-258,992 name=Pyro
WAVEPROBE_BOT client=19 userid=84 class=7 hp=175 timer=0 deadline=14.3 origin=-889,-322,992 name=Pyro
WAVEPROBE_BOT client=21 userid=82 class=7 hp=175 timer=0 deadline=17.9 origin=-908,-237,992 name=Pyro
WAVEPROBE_BOT client=24 userid=79 class=7 hp=175 timer=0 deadline=18.4 origin=-1035,-425,992 name=Pyro
WAVEPROBE_DEBUG_END alive=8 listed=8
```

</details>

<details><summary>mvm_sludge_b6_int_logging_dreadwood · Bot Surge on · wave 6 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 21 | 0 | 0 | 0 | 61 | 0% |
| 80 | 1630 | 33 | 32 | 1 | 61 | 0% |
| 161 | 3238 | 65 | 64 | 1 | 61 | 0% |
| 251 | 5048 | 101 | 101 | 0 | 61 | 0% |
| 332 | 6655 | 133 | 133 | 0 | 61 | 0% |
| 412 | 8264 | 165 | 165 | 0 | 61 | 0% |
| 493 | 9873 | 197 | 197 | 0 | 61 | 0% |
| 573 | 11480 | 229 | 229 | 0 | 61 | 0% |
| 653 | 13089 | 262 | 261 | 1 | 61 | 0% |
| 744 | 14898 | 298 | 297 | 1 | 61 | 0% |
| 824 | 16506 | 330 | 329 | 1 | 61 | 0% |
| 901 | 18035 | 360 | 360 | 0 | 61 | 0% |

Reason: wave exceeded wall-clock time limit: wave 6 still active after 15m0s (18035.1 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=6 expected=6 game=18061.5 engine=4563.6 timescale=20.0 remaining=61 spawns=360 kills=360 attempts=360 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_snowpine_rc4_fix1_adv_permafrost_panic · Bot Surge off · wave 1 · Changelevel Failure</summary>

No wave timeline: the mission did not load or the probe did not start.

Reason: changelevel did not reach mvm_snowpine_rc4_fix1: before mvm_creepside_b2/mvm_creepside_b2_adv_catastrophic_conjuring, after mvm_decoy/mvm_decoy_adv_grutesque_getaway (idle); reply ""; command error ""; timed out after 1m30s (last status {State:idle Reason:none Map:mvm_decoy Pop:mvm_decoy_adv_grutesque_getaway Max:0 GameWave:0 Expected:0 Observed:0 Bots:0 Tanks:0 BotSpawns:0 TankSpawns:0 Attempts:0 Alive:0 Remaining:0 Initial:0 DefClass:0 DefTeam:0 PlayerTeam:2 EnemyTeam:3 Elapsed:1 Progress:-1})

Changelevel evidence: requested `mvm_snowpine_rc4_fix1`, before `mvm_creepside_b2`/`mvm_creepside_b2_adv_catastrophic_conjuring`, after `mvm_decoy`/`mvm_decoy_adv_grutesque_getaway` (state `idle`). Reply: ``; command error: ``.

</details>

<details><summary>mvm_snowpine_rc4_fix1_adv_permafrost_panic · Bot Surge on · wave 1 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 0 | 0 | 0 | 79 | 0% |
| 80 | 1628 | 0 | 0 | 0 | 79 | 0% |
| 161 | 3237 | 0 | 0 | 0 | 79 | 0% |
| 251 | 5046 | 0 | 0 | 0 | 79 | 0% |
| 332 | 6656 | 0 | 0 | 0 | 79 | 0% |
| 412 | 8266 | 0 | 0 | 0 | 79 | 0% |
| 493 | 9874 | 0 | 0 | 0 | 79 | 0% |
| 573 | 11483 | 0 | 0 | 0 | 79 | 0% |
| 654 | 13093 | 0 | 0 | 0 | 79 | 0% |
| 744 | 14903 | 0 | 0 | 0 | 79 | 0% |
| 825 | 16512 | 0 | 0 | 0 | 79 | 0% |
| 900 | 18020 | 0 | 0 | 0 | 79 | 0% |

Reason: wave exceeded wall-clock time limit: wave 1 still active after 15m0s (18020.5 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=1 expected=1 game=18045.7 engine=2823.4 timescale=20.0 remaining=79 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_snowpine_rc4_fix1_adv_permafrost_panic · Bot Surge off · wave 3 · Probe Error</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 21 | 0 | 0 | 0 | 99 | 0% |
| 41 | 800 | 42 | 17 | 0 | 99 | 0% |
| 91 | 1784 | 62 | 36 | 1 | 99 | 0% |
| 132 | 2572 | 78 | 52 | 1 | 99 | 0% |
| 172 | 3362 | 93 | 68 | 0 | 99 | 0% |
| 223 | 4337 | 113 | 87 | 1 | 99 | 0% |
| 263 | 5107 | 128 | 103 | 0 | 99 | 0% |
| 314 | 6075 | 147 | 122 | 0 | 99 | 0% |
| 355 | 6836 | 163 | 137 | 1 | 99 | 0% |
| 396 | 7602 | 178 | 152 | 1 | 99 | 0% |
| 447 | 8542 | 197 | 171 | 1 | 99 | 0% |
| 540 | 1 | 0 | 0 | 0 | 0 | unknown |

Reason: probe state reset during wave 3: state=idle pop=mvm_decoy_adv_grutesque_getaway gamewave=0

Failure snapshot:

```text
WAVEPROBE_DEBUG state=0 wave=0 expected=0 game=1.0 engine=16.9 timescale=1.0 remaining=0 spawns=0 kills=0 attempts=0 defender=0
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_snowpine_rc4_fix1_adv_permafrost_panic · Bot Surge on · wave 3 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 21 | 0 | 0 | 0 | 99 | 0% |
| 80 | 1629 | 59 | 33 | 1 | 99 | 0% |
| 161 | 3237 | 91 | 65 | 1 | 99 | 0% |
| 251 | 5048 | 127 | 101 | 1 | 99 | 0% |
| 332 | 6657 | 159 | 134 | 0 | 99 | 0% |
| 412 | 8264 | 191 | 166 | 0 | 99 | 0% |
| 493 | 9873 | 223 | 198 | 0 | 99 | 0% |
| 573 | 11481 | 255 | 230 | 0 | 99 | 0% |
| 653 | 13088 | 287 | 262 | 0 | 99 | 0% |
| 744 | 14898 | 324 | 298 | 1 | 99 | 0% |
| 824 | 16508 | 356 | 330 | 1 | 99 | 0% |
| 901 | 18035 | 386 | 361 | 0 | 99 | 0% |

Reason: wave exceeded wall-clock time limit: wave 3 still active after 15m0s (18034.9 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=3 expected=3 game=18061.3 engine=4570.8 timescale=20.0 remaining=99 spawns=386 kills=361 attempts=361 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_snowpine_rc4_fix1_adv_permafrost_panic · Bot Surge off · wave 5 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 21 | 0 | 0 | 0 | 92 | 0% |
| 80 | 1629 | 853 | 64 | 2 | 92 | 0% |
| 161 | 3236 | 1717 | 130 | 0 | 92 | 0% |
| 251 | 5046 | 2689 | 200 | 2 | 92 | 0% |
| 332 | 6654 | 3551 | 264 | 0 | 92 | 0% |
| 412 | 8261 | 4415 | 328 | 1 | 92 | 0% |
| 492 | 9870 | 5279 | 392 | 0 | 92 | 0% |
| 573 | 11478 | 6145 | 456 | 2 | 92 | 0% |
| 653 | 13085 | 7009 | 520 | 2 | 92 | 0% |
| 744 | 14895 | 7981 | 594 | 0 | 92 | 0% |
| 824 | 16502 | 8845 | 658 | 0 | 92 | 0% |
| 900 | 18030 | 9667 | 718 | 2 | 92 | 0% |

Reason: wave exceeded wall-clock time limit: wave 5 still active after 15m0s (18029.7 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=5 expected=5 game=18056.0 engine=4603.6 timescale=20.0 remaining=92 spawns=9667 kills=718 attempts=718 defender=32
WAVEPROBE_BOT client=29 userid=77 class=2 hp=125 timer=0 deadline=3.6 origin=247,-1800,197 name=Sniper
WAVEPROBE_BOT client=30 userid=76 class=2 hp=125 timer=0 deadline=8.9 origin=289,-2034,215 name=Sniper
WAVEPROBE_DEBUG_END alive=2 listed=2
```

</details>

<details><summary>mvm_snowpine_rc4_fix1_int_alpine_assault · Bot Surge off · wave 1 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 0 | 0 | 0 | 81 | 0% |
| 80 | 1629 | 0 | 0 | 0 | 81 | 0% |
| 161 | 3236 | 0 | 0 | 0 | 81 | 0% |
| 251 | 5046 | 0 | 0 | 0 | 81 | 0% |
| 332 | 6654 | 0 | 0 | 0 | 81 | 0% |
| 412 | 8263 | 0 | 0 | 0 | 81 | 0% |
| 493 | 9872 | 0 | 0 | 0 | 81 | 0% |
| 573 | 11480 | 0 | 0 | 0 | 81 | 0% |
| 653 | 13089 | 0 | 0 | 0 | 81 | 0% |
| 744 | 14898 | 0 | 0 | 0 | 81 | 0% |
| 824 | 16506 | 0 | 0 | 0 | 81 | 0% |
| 901 | 18035 | 0 | 0 | 0 | 81 | 0% |

Reason: wave exceeded wall-clock time limit: wave 1 still active after 15m0s (18035.3 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=1 expected=1 game=18060.7 engine=3922.6 timescale=20.0 remaining=81 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_snowpine_rc4_fix1_int_alpine_assault · Bot Surge on · wave 1 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 0 | 0 | 0 | 81 | 0% |
| 80 | 1629 | 0 | 0 | 0 | 81 | 0% |
| 161 | 3238 | 0 | 0 | 0 | 81 | 0% |
| 251 | 5047 | 0 | 0 | 0 | 81 | 0% |
| 332 | 6656 | 0 | 0 | 0 | 81 | 0% |
| 412 | 8265 | 0 | 0 | 0 | 81 | 0% |
| 493 | 9874 | 0 | 0 | 0 | 81 | 0% |
| 573 | 11484 | 0 | 0 | 0 | 81 | 0% |
| 654 | 13092 | 0 | 0 | 0 | 81 | 0% |
| 744 | 14902 | 0 | 0 | 0 | 81 | 0% |
| 824 | 16510 | 0 | 0 | 0 | 81 | 0% |
| 901 | 18038 | 0 | 0 | 0 | 81 | 0% |

Reason: wave exceeded wall-clock time limit: wave 1 still active after 15m0s (18038.5 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=1 expected=1 game=18063.8 engine=5517.6 timescale=20.0 remaining=81 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_snowpine_rc4_fix1_int_alpine_assault · Bot Surge off · wave 3 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 21 | 0 | 0 | 0 | 75 | 0% |
| 80 | 1630 | 0 | 0 | 0 | 75 | 0% |
| 161 | 3238 | 0 | 0 | 0 | 75 | 0% |
| 251 | 5050 | 0 | 0 | 0 | 75 | 0% |
| 332 | 6659 | 0 | 0 | 0 | 75 | 0% |
| 412 | 8268 | 0 | 0 | 0 | 75 | 0% |
| 493 | 9876 | 0 | 0 | 0 | 75 | 0% |
| 573 | 11485 | 0 | 0 | 0 | 75 | 0% |
| 654 | 13095 | 0 | 0 | 0 | 75 | 0% |
| 744 | 14906 | 0 | 0 | 0 | 75 | 0% |
| 825 | 16514 | 0 | 0 | 0 | 75 | 0% |
| 900 | 18025 | 0 | 0 | 0 | 75 | 0% |

Reason: wave exceeded wall-clock time limit: wave 3 still active after 15m0s (18024.6 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=3 expected=3 game=18050.9 engine=6385.6 timescale=20.0 remaining=75 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_snowpine_rc4_fix1_int_alpine_assault · Bot Surge on · wave 3 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 0 | 0 | 0 | 75 | 0% |
| 80 | 1628 | 0 | 0 | 0 | 75 | 0% |
| 161 | 3236 | 0 | 0 | 0 | 75 | 0% |
| 251 | 5046 | 0 | 0 | 0 | 75 | 0% |
| 332 | 6655 | 0 | 0 | 0 | 75 | 0% |
| 412 | 8265 | 0 | 0 | 0 | 75 | 0% |
| 493 | 9874 | 0 | 0 | 0 | 75 | 0% |
| 573 | 11482 | 0 | 0 | 0 | 75 | 0% |
| 654 | 13092 | 0 | 0 | 0 | 75 | 0% |
| 744 | 14901 | 0 | 0 | 0 | 75 | 0% |
| 824 | 16509 | 0 | 0 | 0 | 75 | 0% |
| 901 | 18036 | 0 | 0 | 0 | 75 | 0% |

Reason: wave exceeded wall-clock time limit: wave 3 still active after 15m0s (18036.5 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=3 expected=3 game=18062.8 engine=2996.1 timescale=20.0 remaining=75 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_snowpine_rc4_fix1_int_alpine_assault · Bot Surge off · wave 5 · No Enemies Spawned</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 21 | 0 | 0 | 0 | 24 | 0% |
| 80 | 1630 | 0 | 0 | 0 | 24 | 0% |
| 161 | 3238 | 0 | 0 | 0 | 24 | 0% |
| 251 | 5049 | 0 | 0 | 0 | 24 | 0% |
| 332 | 6658 | 0 | 0 | 0 | 24 | 0% |
| 412 | 8267 | 0 | 0 | 0 | 24 | 0% |
| 493 | 9876 | 0 | 0 | 0 | 24 | 0% |
| 573 | 11486 | 0 | 0 | 0 | 24 | 0% |
| 654 | 13095 | 0 | 0 | 0 | 24 | 0% |
| 744 | 14905 | 0 | 0 | 0 | 24 | 0% |
| 825 | 16513 | 0 | 0 | 0 | 24 | 0% |
| 900 | 18022 | 0 | 0 | 0 | 24 | 0% |

Reason: wave exceeded wall-clock time limit: wave 5 still active after 15m0s (18021.6 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=5 expected=5 game=18050.7 engine=7304.2 timescale=20.0 remaining=24 spawns=0 kills=0 attempts=0 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_spybase_rc8_exp_waters_of_a_robot_regime · Bot Surge off · wave 6 · Probe Error</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 15 | 3 | 0 | 3 | 1 | 0% |
| 50 | 1022 | 3 | 1 | 0 | 1 | 0% |
| 111 | 2229 | 3 | 1 | 0 | 1 | 0% |
| 161 | 3234 | 3 | 1 | 0 | 1 | 0% |
| 211 | 4238 | 3 | 1 | 0 | 1 | 0% |
| 271 | 5445 | 3 | 1 | 0 | 1 | 0% |
| 322 | 6450 | 3 | 1 | 0 | 1 | 0% |
| 382 | 7656 | 3 | 1 | 0 | 1 | 0% |
| 432 | 8661 | 3 | 1 | 0 | 1 | 0% |
| 483 | 9666 | 3 | 1 | 0 | 1 | 0% |
| 543 | 10872 | 3 | 1 | 0 | 1 | 0% |
| 640 | 1 | 0 | 0 | 0 | 0 | unknown |

Reason: probe state reset during wave 6: state=idle pop=mvm_decoy_adv_grutesque_getaway gamewave=0

Failure snapshot:

```text
WAVEPROBE_DEBUG state=0 wave=0 expected=0 game=1.0 engine=19.5 timescale=1.0 remaining=0 spawns=0 kills=0 attempts=0 defender=0
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_spybase_rc8_exp_waters_of_a_robot_regime · Bot Surge on · wave 6 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 19 | 19 | 1 | 0 | 1 | 0% |
| 80 | 1628 | 19 | 1 | 0 | 1 | 0% |
| 161 | 3236 | 19 | 1 | 0 | 1 | 0% |
| 251 | 5045 | 19 | 1 | 0 | 1 | 0% |
| 332 | 6653 | 19 | 1 | 0 | 1 | 0% |
| 412 | 8260 | 19 | 1 | 0 | 1 | 0% |
| 492 | 9868 | 19 | 1 | 0 | 1 | 0% |
| 573 | 11476 | 19 | 1 | 0 | 1 | 0% |
| 653 | 13084 | 19 | 1 | 0 | 1 | 0% |
| 744 | 14892 | 19 | 1 | 0 | 1 | 0% |
| 824 | 16500 | 19 | 1 | 0 | 1 | 0% |
| 900 | 18027 | 19 | 1 | 0 | 1 | 0% |

Reason: wave exceeded wall-clock time limit: wave 6 still active after 15m0s (18026.9 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=6 expected=6 game=18053.1 engine=1888.0 timescale=20.0 remaining=1 spawns=19 kills=1 attempts=1 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_terrorlict_final1c5_adv_accursed_aggrievocation · Bot Surge off · wave 2 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 21 | 2 | 0 | 1 | 66 | 0% |
| 80 | 1628 | 186 | 178 | 2 | 66 | 0% |
| 161 | 3235 | 362 | 354 | 2 | 66 | 0% |
| 251 | 5044 | 562 | 553 | 3 | 66 | 0% |
| 332 | 6647 | 738 | 730 | 2 | 66 | 0% |
| 412 | 8255 | 916 | 908 | 2 | 66 | 0% |
| 492 | 9863 | 1092 | 1083 | 3 | 66 | 0% |
| 573 | 11471 | 1268 | 1260 | 2 | 66 | 0% |
| 653 | 13079 | 1444 | 1436 | 2 | 66 | 0% |
| 744 | 14888 | 1644 | 1636 | 2 | 66 | 0% |
| 824 | 16496 | 1820 | 1812 | 2 | 66 | 0% |
| 900 | 18023 | 1990 | 1982 | 2 | 66 | 0% |

Reason: wave exceeded wall-clock time limit: wave 2 still active after 15m0s (18022.8 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=2 expected=2 game=18049.2 engine=1839.2 timescale=20.0 remaining=66 spawns=1990 kills=1982 attempts=1982 defender=32
WAVEPROBE_BOT client=30 userid=5 class=7 hp=175 timer=0 deadline=21.3 origin=-1364,1529,-3 name=Monocultus Pyro
WAVEPROBE_BOT client=31 userid=4 class=7 hp=175 timer=0 deadline=13.0 origin=-1249,1639,-3 name=Monocultus Pyro
WAVEPROBE_DEBUG_END alive=2 listed=2
```

</details>

<details><summary>mvm_terrorlict_final1c5_adv_accursed_aggrievocation · Bot Surge off · wave 4 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 3 | 0 | 3 | 89 | 0% |
| 80 | 1629 | 45 | 43 | 2 | 89 | 0% |
| 161 | 3238 | 85 | 85 | 0 | 89 | 0% |
| 251 | 5047 | 133 | 133 | 0 | 89 | 0% |
| 332 | 6656 | 175 | 175 | 0 | 89 | 0% |
| 412 | 8264 | 217 | 215 | 2 | 89 | 0% |
| 493 | 9873 | 259 | 257 | 2 | 89 | 0% |
| 573 | 11481 | 299 | 299 | 0 | 89 | 0% |
| 653 | 13090 | 341 | 341 | 0 | 89 | 0% |
| 744 | 14900 | 389 | 387 | 2 | 89 | 0% |
| 824 | 16509 | 431 | 429 | 2 | 89 | 0% |
| 901 | 18036 | 469 | 469 | 0 | 89 | 0% |

Reason: wave exceeded wall-clock time limit: wave 4 still active after 15m0s (18036.1 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=4 expected=4 game=18062.2 engine=1839.0 timescale=20.0 remaining=89 spawns=469 kills=469 attempts=469 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_terrorlict_final1c5_adv_accursed_aggrievocation · Bot Surge on · wave 4 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 21 | 3 | 0 | 3 | 89 | 0% |
| 80 | 1630 | 45 | 43 | 2 | 89 | 0% |
| 161 | 3238 | 85 | 85 | 0 | 89 | 0% |
| 251 | 5048 | 133 | 133 | 0 | 89 | 0% |
| 332 | 6657 | 175 | 175 | 0 | 89 | 0% |
| 412 | 8265 | 217 | 215 | 2 | 89 | 0% |
| 493 | 9874 | 259 | 257 | 2 | 89 | 0% |
| 573 | 11481 | 299 | 299 | 0 | 89 | 0% |
| 653 | 13089 | 341 | 341 | 0 | 89 | 0% |
| 744 | 14898 | 389 | 387 | 2 | 89 | 0% |
| 824 | 16506 | 431 | 429 | 2 | 89 | 0% |
| 901 | 18035 | 469 | 469 | 0 | 89 | 0% |

Reason: wave exceeded wall-clock time limit: wave 4 still active after 15m0s (18035.1 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=4 expected=4 game=18061.5 engine=3732.7 timescale=20.0 remaining=89 spawns=469 kills=469 attempts=469 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_terrorlict_final1c5_adv_accursed_aggrievocation · Bot Surge off · wave 6 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 21 | 7 | 0 | 2 | 53 | 0% |
| 80 | 1629 | 140 | 46 | 1 | 53 | 0% |
| 161 | 3237 | 229 | 79 | 1 | 53 | 0% |
| 251 | 5047 | 329 | 121 | 0 | 53 | 0% |
| 332 | 6654 | 418 | 155 | 1 | 53 | 0% |
| 412 | 8262 | 508 | 190 | 1 | 53 | 0% |
| 492 | 9870 | 597 | 222 | 0 | 53 | 0% |
| 573 | 11478 | 686 | 258 | 0 | 53 | 0% |
| 653 | 13084 | 775 | 294 | 0 | 53 | 0% |
| 744 | 14892 | 876 | 331 | 1 | 53 | 0% |
| 824 | 16500 | 965 | 372 | 1 | 53 | 0% |
| 900 | 18027 | 1050 | 405 | 1 | 53 | 0% |

Reason: wave exceeded wall-clock time limit: wave 6 still active after 15m0s (18026.9 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=6 expected=6 game=18053.3 engine=4577.7 timescale=20.0 remaining=53 spawns=1050 kills=405 attempts=405 defender=32
WAVEPROBE_BOT client=26 userid=42 class=1 hp=125 timer=0 deadline=12.1 origin=-1316,1597,-3 name=Voyage Scout
WAVEPROBE_DEBUG_END alive=1 listed=1
```

</details>

<details><summary>mvm_terrorlict_final1c5_adv_accursed_aggrievocation · Bot Surge on · wave 6 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 20 | 8 | 0 | 4 | 53 | 0% |
| 81 | 1060 | 922 | 440 | 22 | 53 | 0% |
| 162 | 1827 | 1582 | 748 | 22 | 53 | 0% |
| 244 | 2391 | 2066 | 982 | 16 | 53 | 0% |
| 325 | 2939 | 2550 | 1200 | 22 | 53 | 0% |
| 406 | 3845 | 3320 | 1565 | 18 | 53 | 0% |
| 497 | 4732 | 4090 | 1911 | 22 | 53 | 0% |
| 578 | 5436 | 4706 | 2195 | 22 | 53 | 0% |
| 659 | 6084 | 5256 | 2456 | 22 | 53 | 0% |
| 741 | 6746 | 5828 | 2727 | 22 | 53 | 0% |
| 822 | 7334 | 6334 | 2964 | 21 | 53 | 0% |
| 901 | 8044 | 6950 | 3250 | 22 | 53 | 0% |

Reason: wave exceeded wall-clock time limit: wave 6 still active after 15m0s (8043.7 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=6 expected=6 game=26111.8 engine=4635.2 timescale=20.0 remaining=53 spawns=6950 kills=3250 attempts=3250 defender=32
WAVEPROBE_BOT client=10 userid=79 class=1 hp=125 timer=0 deadline=17.2 origin=-1361,1465,-3 name=Voyage Scout
WAVEPROBE_BOT client=11 userid=78 class=1 hp=125 timer=0 deadline=12.2 origin=-1430,-1768,-401 name=Voyage Scout
WAVEPROBE_BOT client=12 userid=77 class=1 hp=125 timer=0 deadline=9.7 origin=-1417,-1617,-407 name=Voyage Scout
WAVEPROBE_BOT client=13 userid=76 class=1 hp=125 timer=0 deadline=8.5 origin=-1305,-1498,-423 name=Voyage Scout
WAVEPROBE_BOT client=14 userid=75 class=1 hp=125 timer=0 deadline=16.0 origin=-1503,1577,32 name=Voyage Scout
WAVEPROBE_BOT client=15 userid=74 class=1 hp=125 timer=0 deadline=6.9 origin=-1212,1511,-3 name=Voyage Scout
WAVEPROBE_BOT client=16 userid=73 class=1 hp=125 timer=0 deadline=6.2 origin=-1270,-1571,-432 name=Voyage Scout
WAVEPROBE_BOT client=17 userid=72 class=1 hp=125 timer=0 deadline=12.4 origin=-1624,1713,-85 name=Voyage Scout
WAVEPROBE_BOT client=18 userid=71 class=1 hp=125 timer=0 deadline=11.9 origin=-1296,-1710,-428 name=Voyage Scout
WAVEPROBE_BOT client=19 userid=70 class=1 hp=125 timer=0 deadline=13.3 origin=-1212,1386,15 name=Voyage Scout
WAVEPROBE_BOT client=20 userid=69 class=1 hp=125 timer=0 deadline=10.6 origin=-1567,-1658,-408 name=Voyage Scout
WAVEPROBE_BOT client=21 userid=68 class=1 hp=125 timer=0 deadline=14.4 origin=-1212,1764,-3 name=Voyage Scout
WAVEPROBE_DEBUG_END alive=22 listed=12
```

</details>

<details><summary>mvm_terrorlict_final1c5_exp_echoes_of_a_warzone · Bot Surge off · wave 1 · Wave Failed</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 21 | 10 | 0 | 11 | 764 | 0% |
| 10 | 217 | 77 | 36 | 1 | 764 | 0% |

Reason: game or probe failed wave 1: wave_failed

Failure snapshot:

```text
WAVEPROBE_DEBUG state=4 wave=1 expected=1 game=242.0 engine=1164.9 timescale=20.0 remaining=764 spawns=77 kills=36 attempts=36 defender=32
WAVEPROBE_DEBUG_END alive=0 listed=0
```

</details>

<details><summary>mvm_terrorlict_final1c5_exp_echoes_of_a_warzone · Bot Surge on · wave 1 · Wave Timed Out</summary>

| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 14 | 15 | 0 | 16 | 764 | 0% |
| 81 | 1101 | 629 | 222 | 21 | 764 | 0% |
| 162 | 2281 | 1211 | 478 | 6 | 764 | 0% |
| 253 | 3661 | 1874 | 729 | 16 | 764 | 0% |
| 334 | 4922 | 2473 | 1013 | 5 | 764 | 0% |
| 415 | 5984 | 3007 | 1253 | 3 | 764 | 0% |
| 495 | 7354 | 3688 | 1559 | 20 | 764 | 0% |
| 576 | 8695 | 4332 | 1828 | 10 | 764 | 0% |
| 657 | 10030 | 4965 | 2092 | 2 | 764 | 0% |
| 748 | 11370 | 5629 | 2318 | 4 | 764 | 0% |
| 828 | 12921 | 6395 | 2609 | 2 | 764 | 0% |
| 901 | 14310 | 7050 | 2896 | 8 | 764 | 0% |

Reason: wave exceeded wall-clock time limit: wave 1 still active after 15m0s (14310.3 game seconds)

Failure snapshot:

```text
WAVEPROBE_DEBUG state=2 wave=1 expected=1 game=14335.6 engine=5504.5 timescale=20.0 remaining=764 spawns=7050 kills=2896 attempts=2896 defender=32
WAVEPROBE_BOT client=11 userid=37 class=2 hp=125 timer=0 deadline=9.0 origin=-1290,-110,-324 name=Sniper
WAVEPROBE_BOT client=13 userid=35 class=8 hp=125 timer=0 deadline=5.2 origin=3415,192,-16 name=Spy
WAVEPROBE_BOT client=19 userid=29 class=8 hp=125 timer=0 deadline=6.9 origin=3456,-417,-16 name=Spy
WAVEPROBE_BOT client=20 userid=28 class=2 hp=125 timer=0 deadline=3.8 origin=-1139,72,-308 name=Sniper
WAVEPROBE_BOT client=21 userid=27 class=2 hp=125 timer=0 deadline=7.3 origin=-1592,-237,-313 name=Sniper
WAVEPROBE_BOT client=31 userid=17 class=8 hp=125 timer=0 deadline=0.5 origin=3459,186,-11 name=Spy
WAVEPROBE_DEBUG_END alive=6 listed=6
```

</details>

## Files

- [Machine-readable summary](SUMMARY.json)
- [screen-0.jsonl](screen-0.jsonl)
- [screen-1.jsonl](screen-1.jsonl)
- [screen-10.jsonl](screen-10.jsonl)
- [screen-11.jsonl](screen-11.jsonl)
- [screen-2.jsonl](screen-2.jsonl)
- [screen-3.jsonl](screen-3.jsonl)
- [screen-4.jsonl](screen-4.jsonl)
- [screen-5.jsonl](screen-5.jsonl)
- [screen-6.jsonl](screen-6.jsonl)
- [screen-7.jsonl](screen-7.jsonl)
- [screen-8.jsonl](screen-8.jsonl)
- [screen-9.jsonl](screen-9.jsonl)
- [retest-0.jsonl](retest-0.jsonl)
- [retest-1.jsonl](retest-1.jsonl)
- [retest-10.jsonl](retest-10.jsonl)
- [retest-11.jsonl](retest-11.jsonl)
- [retest-2.jsonl](retest-2.jsonl)
- [retest-3.jsonl](retest-3.jsonl)
- [retest-4.jsonl](retest-4.jsonl)
- [retest-5.jsonl](retest-5.jsonl)
- [retest-6.jsonl](retest-6.jsonl)
- [retest-7.jsonl](retest-7.jsonl)
- [retest-8.jsonl](retest-8.jsonl)
- [retest-9.jsonl](retest-9.jsonl)

A pass is one completion witness at this seed and timing; it does not prove every authored spawn appeared or every timing will complete.
