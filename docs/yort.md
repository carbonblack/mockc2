# Yort

Yort refers to macOS malware created by the Lazarus Group. The threat was 
originally reported on in March of 2019 by Kaspersky. Newer samples were found
in November of 2019 distributed as a fake photo viewing application. Yort is a
backdoor application that communicates using the HTTP protocol and a TLS
connection.

## Network Setup

Add the C2 FQDN to the detonation machines host file and point the domain to
your mock C2 IP address.

## Links

* [https://www.carbonblack.com/blog/threat-analysis-unit-tau-threat-intelligence-notification-osx-yort/](https://www.carbonblack.com/blog/threat-analysis-unit-tau-threat-intelligence-notification-osx-yort/)
* [https://securelist.com/cryptocurrency-businesses-still-being-targeted-by-lazarus/90019/](https://securelist.com/cryptocurrency-businesses-still-being-targeted-by-lazarus/90019/)

## IOCs

| Indicator                                                        | Type   | Context                    |
|------------------------------------------------------------------|--------|----------------------------|
| 735365ef9aa6cca946cfef9a4b85f68e7f9f03011da0cf5f5ab517a381e40d02 | SHA256 | Yort Dropper 64-bit Mach-O |
| 6f7a5f1d52d3bfc6f175bf2bbb665e4bd99b0453e2d2e27712fe9b71c55962dc | SHA256 | Yort 64-bit Mach-O         |
| 65cc7663fa5c5665ad5d9c6bec2b6257612f9f0c0ce7e4399e6dc8b464ea88c0 | SHA256 | Yort 64-bit Mach-O         |
| e63640c53204a59ba59f2c310964149ca3616d79adc40a6c3abd5bf669511756 | SHA256 | Yort 64-bit Mach-O         |
| 3c2f7b8a167433c95aa919da9216f0624032ac9ed9dec71c3c56cacfd5cd1837 | SHA256 | Yort 64-bit Mach-O         |
| c3fa787af394de14d94538e49f0a02f40b403da396d1aba765a60b0bc2dcfdac | SHA256 | Yort 64-bit Mach-O         |
| f9ffb15a6bf559773b0df7d8a89d9440819ab285f17a7b0a98626c14164d170f | SHA256 | Yort 64-bit Mach-O         |
| towingoperations[.]com                                           | DNS    | Yort C2                    |
| baseballcharlemagnelegardeur[.]com                               | DNS    | Yort C2                    |
| tangowithcolette[.]com                                           | DNS    | Yort C2                    |
| crabbedly[.]club                                                 | DNS    | Yort C2                    |
| craypot[.]live                                                   | DNS    | Yort C2                    |
| indagator[.]club                                                 | DNS    | Yort C2                    |
| fudcitydelivers[.]com                                            | DNS    | Yort C2                    |
| sctemarkets[.]com                                                | DNS    | Yort C2                    |
