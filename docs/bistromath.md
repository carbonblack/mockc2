# BISTROMATH

On February 14th, 2020 the U.S. Department of Homeland Security (DHS) released a
Malware Analysis Report (MAR-10265965-1.v1) which provided information about a
trojan they referred to as BISTROMATH. DHS attributed the trojan to a threat
group based in North Korea, often referred to as Hidden Cobra or the Lazarus
Group. Similar to the other Lazarus Group trojans, BISTROMATH provides backdoor
functionality allowing the attackers remote access to compromised systems using
a custom C2 protocol.

## Network Setup

The C2 IP address is stored as an ASCII string. Use a hex editor to find and
replace the C2 IP address with the IP address of your mock C2 server.

## Links

* [https://www.us-cert.gov/ncas/analysis-reports/ar20-045a](https://www.us-cert.gov/ncas/analysis-reports/ar20-045a)

## IOCs

| Indicator                                                        | Type                 | Context            |
|------------------------------------------------------------------|----------------------|--------------------|
| 618a67048d0a9217317c1d1790ad5f6b044eaa58a433bd46ec2fb9f9ff563dc6 | SHA256              | BISTROMATH Server   |
| 04d70bb249206a006f83db39bbe49ff6e520ea329e5fbb9c758d426b1c8dec30 | SHA256              | BISTROMATH Server   |
| 1ea6b3e99bbb67719c56ad07f5a12501855068a4a866f92db8dcdefaffa48a39 | SHA256              | BISTROMATH Dropper  |
| dfb3837fae611b985d294d2eab8dd3e75cb251791412d21b5d8ef93f14129d72 | SHA256              | BISTROMATH Dropper  |
| b6811b42023524e691b517d19d0321f890f91f35ebbdf1c12cbb92cda5b6de32 | SHA256              | BISTROMATH Dropper  |
| fb0962d9d7268cca7776794aadc7481669daaf24ce9a8551fbd8bebdabcdca7f | SHA256              | BISTROMATH Dropper  |
| 383fda44cc8896b73bce5ee557af1b1680f8ff402a96d6aaf828d04934d2b2d4 | SHA256              | BISTROMATH Dropper  |
| 32b0a1a4e192e9aa1f303afedf0aad4434f95740eabdc470214572ac59d959a8 | SHA256              | BISTROMATH Dropper  |
| 43193c4efa8689ff6de3fb18e30607bb941b43abb21e8cee0cfd664c6f4ad97c | SHA256              | BISTROMATH Backdoor |
| d6498226ca4c008542342e3fd4807031eb11bee8bf1d85ea8dcf5ded36df3679 | SHA256              | BISTROMATH Backdoor |
| 58eb4a96fa7ca302647a41d57b5910a794a0d3154befe5c545f02799734ca2a7 | SHA256              | BISTROMATH Backdoor |
| 133820ebac6e005737d5bb97a5db549490a9f210f4e95098bc9b0a7748f52d1f | SHA256              | BISTROMATH Backdoor |
| 738ba44188a93de6b5ca7e0bf0a77f66f677a0dda2b2e9ef4b91b1c8257da790 | SHA256              | BISTROMATH Backdoor |
| fdf9ce217edbda0e6169d6ba0d653b99b26686feb20a6c84ce33752ec6bb7e4f | SHA256              | BISTROMATH Backdoor |
| 380152e0fa536ad341231d6ab567ecf83e98f8947ff65a855ed5b7f26d108df5 | SHA256              | BISTROMATH Backdoor |
| 159.100.250.231                                                  | TCP/80<br/>TCP/8080 | BISTROMATH C2       |
| 45.76.87.134                                                     | TCP/8080            | BISTROMATH C2       |
