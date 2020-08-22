# MATA

MATA (or Dacls) was first reported on in October of 2019 by 360 Netlab. At the
time 360 Netlab mentioned both Linux and Windows versions of the malware. In May
of 2020, Malwarebytes reported their detection of a macOS version of the Dacls
malware. In July of 2020 Kaspersky reported on their analysis of the framework
used in the Dacls malware, which they called MATA. The MATA name matches symbols
left in some of the malware samples. MATA binaries are capable of providing
backdoor access to remote attackers and makes use of a custom C2 protocol
running over TLS.

## Network Setup

The C2 IP address is stored as an Unicode string. Use a hex editor to find and
replace the C2 IP address with the IP address of your mock C2 server.

## Links

* [https://blog.netlab.360.com/dacls-the-dual-platform-rat-en/](https://blog.netlab.360.com/dacls-the-dual-platform-rat-en/)
* [https://blog.trendmicro.com/trendlabs-security-intelligence/new-macos-dacls-rat-backdoor-show-lazarus-multi-platform-attack-capability/](https://blog.trendmicro.com/trendlabs-security-intelligence/new-macos-dacls-rat-backdoor-show-lazarus-multi-platform-attack-capability/)
* [https://blog.malwarebytes.com/threat-analysis/2020/05/new-mac-variant-of-lazarus-dacls-rat-distributed-via-trojanized-2fa-app/](https://blog.malwarebytes.com/threat-analysis/2020/05/new-mac-variant-of-lazarus-dacls-rat-distributed-via-trojanized-2fa-app/)
* [https://securelist.com/mata-multi-platform-targeted-malware-framework/97746/](https://securelist.com/mata-multi-platform-targeted-malware-framework/97746/)

## IOCs

| Indicator                                                        | Type    | Context            |
|------------------------------------------------------------------|---------|--------------------|
| 846d8647d27a0d729df40b13a644f3bffdc95f6d0e600f2195c85628d59f1dc6 | SHA256  | MATA 64-bit Mach-O |
| ba5b781ebacac07c4b14f9430a23ca0442e294236bd8dd14d1f69c6661551db8 | SHA256  | MATA 64-bit ELF    |
| f34102ed2a2b8313651bdd1d4a08cf0601432f233d52b156745846bd5b90cbd3 | SHA256  | MATA 64-bit ELF    |
| 3095f9326c66c9a035cb12bf50e2115c3aa6f7860dab9a8b8f82a223f366283a | SHA256  | MATA 64-bit DLL    |
| d29bc522d23513cfbb5ff4542382e1b4f0df2fa6bced5fb479cd63b6f902c0eb | SHA256  | MATA 64-bit DLL    |
| 45ab66dbcb78158b2c2448207717646655d804bdc4f975c47fafbe21a0213fbc | SHA256  | MATA 64-bit DLL    |
| cdf74f48c9ea905682155441cf03f4207dbeb2a2f09c4605a5cf4a9a367286e8 | SHA256  | MATA 64-bit DLL    |
| 82d33a67c68f7c476a9ac1e960abc6a911f797446a2c24f0e13b92af1eb385b8 | SHA256  | MATA 32-bit EXE    |
| e3285f3c898230f8e75c33bb6a6cf37acfc292a0e6d0c607beea1c7a84686dfe | SHA256  | MATA 32-bit EXE    |
| f9686467a99cdb3928ccf40042d3e18451a9db97ef60f098656725a9fc3d9025 | SHA256  | MATA 32-bit DLL    |
| 40249bc29030349a85d18677483acb97aca028df8a55fda93728f253f72f2e0a | SHA256  | MATA 32-bit DLL    |
| cdac934dbd8831b962718fdbaf050ebaa8b89be6c98c8cd7479a9d91939c63c6 | SHA256  | MATA 32-bit DLL    |
| 104.232.71.7                                                     | TCP/443 | MATA C2            |
| 107.172.197.175                                                  | TCP/443 | MATA C2            |
| 172.93.184.62                                                    | TCP/443 | MATA C2            |
| 172.93.201.219                                                   | TCP/443 | MATA C2            |
| 185.62.58.207                                                    | TCP/443 | MATA C2            |
| 192.210.213.178                                                  | TCP/443 | MATA C2            |
| 198.180.198.6                                                    | TCP/443 | MATA C2            |
| 209.90.234.34                                                    | TCP/443 | MATA C2            |
| 23.227.196.116                                                   | TCP/443 | MATA C2            |
| 23.227.199.53                                                    | TCP/443 | MATA C2            |
| 23.227.199.69                                                    | TCP/443 | MATA C2            |
| 23.254.119.12                                                    | TCP/443 | MATA C2            |
| 23.81.246.179                                                    | TCP/443 | MATA C2            |
| 37.72.175.179                                                    | TCP/443 | MATA C2            |
| 64.188.19.117                                                    | TCP/443 | MATA C2            |
| 67.43.239.146                                                    | TCP/443 | MATA C2            |
| 74.121.190.121                                                   | TCP/443 | MATA C2            |
