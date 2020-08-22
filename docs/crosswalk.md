# CROSSWALK

In August of 2019, FireEye reported on APT41, a Chinese state sponsored
espionage group. The group has been documented as targeting  healthcare,
high-tech, and telecommunications companies for traditional corporate espionage
purposes.  Additionally this group has also targeted companies in the video game
industry for financial gain. CROSSWALK is a modular backdoor application that
gathers system information and is capable of executing shell code in response to
C2 messages. 

## Network Setup

The C2 IP address is stored in an encrypted format. Set up two virtual machines
and configure them to be on the same virtual network. Give the mock C2 server
the C2 IP address from the malware.

## Links

* [https://www.carbonblack.com/blog/cb-threat-analysis-unit-technical-analysis-of-crosswalk/](https://www.carbonblack.com/blog/cb-threat-analysis-unit-technical-analysis-of-crosswalk/)
* [https://www.fireeye.com/blog/threat-research/2019/08/apt41-dual-espionage-and-cyber-crime-operation.html](https://www.fireeye.com/blog/threat-research/2019/08/apt41-dual-espionage-and-cyber-crime-operation.html)

## IOCs

| Indicator                                                        | Type                       | Context                     |
|------------------------------------------------------------------|----------------------------|-----------------------------|
| efe1c1bfe07981069d17102b8e5c313f769625fd86d95fedc32525518d4cb2de | SHA256                     | CROSSWALK dropper           |
| 300519fa1af5c36371ab438405eb641f184bd2f491bdf24f04e5ca9b86d1b39c | SHA256                     | CROSSWALK 32-bit executable |
| db866ef07dc1f2e1df1e6542323bc672dd245d88c0ee91ce0bd3da2c95aedf68 | SHA256                     | CROSSWALK 32-bit executable |
| 597fcb3b955e03b3d5a27ac450386c92515c9e8867f8ec3731fa05549cd8b0b0 | SHA256                     | CROSSWALK 64-bit executable |
| 45e1a153a4183778f27e5109d5767afb1fa5eec999a306d6b22f11a247c0389b | SHA256                     | CROSSWALK 64-bit executable |
| f6d0cd5b6aa6ccea3ba3cb63b26420f6579d4a07164944e1013e093c521c5687 | SHA256                     | CROSSWALK 64-bit executable |
| 9d0ac935b9e0d6c86fc2904477638af6e4b68d020c2956912e5109cc6219c08f | SHA256                     | CROSSWALK 64-bit DLL        |
| eb27a9f2ee9e0e8073402ee6aaeb6468b5c661b42951ebb1a77246e2261e1b66 | SHA256                     | CROSSWALK 64-bit DLL        |
| 1a7b33f00a4f3d9675ad79891e8eb2bf530242ae809e6630b52279e172148333 | SHA256                     | CROSSWALK 64-bit DLL        |
| c5e46e80c790be6997798241b852c8dbcd88e823a765a581dae228430114a8d4 | SHA256                     | CROSSWALK 64-bit DLL        |
| 160.16.85.174                                                    | TCP/443                    | CROSSWALK C2                |
| 45.32.226.32                                                     | TCP/443                    | CROSSWALK C2                |
| 49.51.138.80                                                     | TCP/443<br/>TCP/5938       | CROSSWALK C2                |
| readme[.]myddns[.]com                                            | DNS<br/>TCP/443            | CROSSWALK C2                |
| wpblog[.]dynamic-dns[.]net                                       | DNS<br/>TCP/443            | CROSSWALK C2                |
| remoteset[.]zyns[.]com                                           | DNS<br/>TCP/80<br/>TCP/443 | CROSSWALK C2                |
