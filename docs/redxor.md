# RedXOR

RedXOR is a backdoor targetting Linux systems. It masquerades as the [polkitd](https://linux.die.net/man/8/polkitd) daemon. Based on victims as well as Tactics, Techniques, and Procedures (TTPs), it is believed to be attributed to a high profile Chinese threat actor. It makes use of a network communication protocol that looks like normal HTTP traffic but has XOR encoded payloads in the content body.

## Network Setup

The `0423258b94e8a9af58ad63ea493818618de2d8c60cf75ec7980edcaa34dcc919` sample makes use of a DNS name of `update.cloudjscdn.com` for it's C2 communication. Simply modify the `hosts` file on the machine where the malware is running and point it to the IP address of your MockC2 server.

## Links

* [https://www.intezer.com/blog/malware-analysis/new-linux-backdoor-redxor-likely-operated-by-chinese-nation-state-actor/](https://www.intezer.com/blog/malware-analysis/new-linux-backdoor-redxor-likely-operated-by-chinese-nation-state-actor/)

## IOCs

| Indicator                                                        | Type     | Context           |
|------------------------------------------------------------------|----------|-------------------|
| 0423258b94e8a9af58ad63ea493818618de2d8c60cf75ec7980edcaa34dcc919 | SHA256   | RedXOR 64-bit ELF |
| 0a76c55fa88d4c134012a5136c09fb938b4be88a382f88bf2804043253b0559f | SHA256   | RedXOR 64-bit ELF |
| 0423258b94e8a9af58ad63ea493818618de2d8c60cf75ec7980edcaa34dcc919 | SHA256   | RedXOR 64-bit ELF |
| update.cloudjscdn.com                                            | TCP/8080 | RedXOR C2         |
| 158.247.208.230                                                  | TCP/8080 | RedXOR C2         |
| www.centosupdateonline.com                                       | TCP/8080 | RedXOR C2         |
