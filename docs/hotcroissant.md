# HOTCROISSANT

On February 14th, 2020 the U.S. Department of Homeland Security (DHS) released a
Malware Analysis Report (MAR-10271944-1.v1) which provided information about a
trojan they referred to as HOTCROISSANT. DHS attributed the trojan to a threat
group based in North Korea, often referred to as Hidden Cobra. This group, also
known as the Lazarus Group, continues to be very active. During 2019, they’ve
targeted organizations in South Korea, Russia, and the United States with
motives that range from espionage and sabotage to attacks purely for financial
gain. HOTCROISSANT provides backdoor functionality allowing the attackers remote
access to compromised systems using a custom C2 protocol.

## Network Setup

The C2 IP address is encrypted with a single byte key of `0xF`. Take the
original C2 IP address as well as your mock C2 IP address and encrypt the
strings with the single byte key of `0xF`. Then use a hex editor to find and
replace the encrypted C2 IP address with the encrypted IP address of your mock
C2 server.

## Links

* [https://www.carbonblack.com/blog/vmware-carbon-black-tau-threat-analysis-the-evolution-of-lazarus/](https://www.carbonblack.com/blog/vmware-carbon-black-tau-threat-analysis-the-evolution-of-lazarus/)
* [https://www.us-cert.gov/ncas/analysis-reports/ar20-045d](https://www.us-cert.gov/ncas/analysis-reports/ar20-045d)

## IOCs

| Indicator                                                        | Type     | Context                        |
|------------------------------------------------------------------|----------|--------------------------------|
| 7ec13c5258e4b3455f2e8af1c55ac74de6195b837235b58bc32f95dd6f25370c | SHA256   | HOTCROISSANT 32-bit executable |
| 0ea57d676fe7bb7f75387becffffbd7e6037151e581389d5b864270b296bb765 | SHA256   | HOTCROISSANT 32-bit executable |
| b689815a0c97414e0bba0f6cf72029691c8254041e105ed69f6f921d49e88a4d | SHA256   | HOTCROISSANT 32-bit executable |
| 8ee7da59f68c691c9eca1ac70ff03155ed07808c7a66dee49886b51a59e00085 | SHA256   | HOTCROISSANT 32-bit executable |
| 315c06bd8c75f99722fd014b4fb4bd8934049cde09afead9b46bddf4cdd63171 | SHA256   | HOTCROISSANT 32-bit executable |
| 61fb7ee577c3738e5594b31c227fd1a49715d8c353272b0b23bdf9de32007df9 | SHA256   | HOTCROISSANT 32-bit executable |
| 8c2bfc3c41e7112e922696301224549671b5c1ce3c5a138e8a4cb104ec8b16d6 | SHA256   | HOTCROISSANT 32-bit executable |
| b0952756f2841d0c86c606e30372445650eb414ec3f1cd1b590bfd836c72ab3e | SHA256   | HOTCROISSANT 32-bit executable |
| b4a7651a7d667074817720b49c190518a625b7216cbedf5a0034a862dfb0c882 | SHA256   | HOTCROISSANT 32-bit executable |
| a289fffa8b700d75b07bed5a81e36476307f72caa912b8164030ce41eaf3fd6b | SHA256   | HOTCROISSANT 32-bit executable |
| 1a0fcfc7cc9c03d30c927a136eb7fe9eb03150ef4550405016c71b13727a2ed1 | SHA256   | HOTCROISSANT 32-bit executable |
| 1275274692f1990939706e7b3217e8426639b7b0ee2b4244492f6d5fe42d97f4 | SHA256   | HOTCROISSANT 32-bit executable |
| b490d47d284afc330703f28400546283133da64bca5348cfe371a5775014fc27 | SHA256   | HOTCROISSANT 32-bit executable |
| 580d168f5f92ab4cab77f19527bf85dda1ad3a491a958869c930a4d42a1c91f6 | SHA256   | HOTCROISSANT 32-bit executable |
| cd980ee808d90db514a4893019bdf59c7f6fdab80592439da72b8d9c72598c16 | SHA256   | HOTCROISSANT 32-bit executable |
| 19d400b5df46d4e3fb7aaa61865f97afa1ec8cbe351bc2db8bbbdbff336494b3 | SHA256   | HOTCROISSANT 32-bit executable |
| 172.93.110.85                                                    | TCP/80   | HOTCROISSANT C2                |
| 176.31.15.195                                                    | TCP/8445 | HOTCROISSANT C2                |
| 94.177.123.138                                                   | TCP/8088 | HOTCROISSANT C2                |
| 51.254.60.208                                                    | TCP/443  | HOTCROISSANT C2                |
| 192.99.223.115                                                   | TCP/8080 | HOTCROISSANT C2                |
| 61.78.63.123                                                     | TCP/443  | HOTCROISSANT C2                |
| 86.106.131.150                                                   | TCP/443  | HOTCROISSANT C2                |
| 192.161.182.250                                                  | TCP/80   | HOTCROISSANT C2                |
