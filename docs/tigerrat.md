# TigerRAT

On September 5th, 2021, The Korea Internet & Security Agency (KISA) released a
report on a new threat they dubbed TigerRAT. The newly found malware shared 
similarities with malware that were previously reported on by Kaspersky and 
Malwarebytes. These have been attributed to Andariel a sub-group of Lazarus.

## Network Setup

TigerRAT makes use of DES encryption to store the C2 information. Set up two
virtual machines and configure them to be on the same virtual network. Give the
mock C2 server the C2 IP address from the malware.

## Links

* [https://twitter.com/heavyrain_89/status/1434696945268756481](https://twitter.com/heavyrain_89/status/1434696945268756481)
* [https://www.boho.or.kr/filedownload.do?attach_file_seq=3277&attach_file_id=EpF3277.pdf](https://www.boho.or.kr/filedownload.do?attach_file_seq=3277&attach_file_id=EpF3277.pdf )
* [https://securelist.com/andariel-evolves-to-target-south-korea-with-ransomware/102811/](https://securelist.com/andariel-evolves-to-target-south-korea-with-ransomware/102811/)
* [https://blog.malwarebytes.com/threat-intelligence/2021/04/lazarus-apt-conceals-malicious-code-within-bmp-file-to-drop-its-rat/](https://blog.malwarebytes.com/threat-intelligence/2021/04/lazarus-apt-conceals-malicious-code-within-bmp-file-to-drop-its-rat/)

## IOCs

|Indicator                                                         |Type     |Context          |
|------------------------------------------------------------------|---------|-----------------|
| 1f8dcfaebbcd7e71c2872e0ba2fc6db81d651cf654a21d33c78eae6662e62392 | SHA256  | TigerRAT Loader |
| f32f6b229913d68daad937cc72a57aa45291a9d623109ed48938815aa7b6005c | SHA256  | TigerRAT        |
| 29c6044d65af0073424ccc01abcb8411cbdc52720cac957a3012773c4380bab3 | SHA256  | TigerRAT        |
| fed94f461145681dc9347b382497a72542424c64b6ae6fcf945f4becd2d46c32 | SHA256  | TigerRAT        |
| 6dcfb2f52521672743f4888e992229896b98ab0e6bd979311ebdb4dcccc2b2e6 | SHA256  | TigerRAT        |
| 52.202.193.124                                                   | TCP/443 | TigerRAT C2     |
| 185.208.158.204                                                  | TCP/443 | TigerRAT C2     |
| 185.208.158.208                                                  | TCP/443 | TigerRAT C2     |
