# SLICKSHOES

On February 14th, 2020 the U.S. Department of Homeland Security (DHS) released a
Malware Analysis Report (MAR-10265965-2.v1) which provided information about a
trojan they referred to as SLICKSHOES. DHS attributed the trojan to a threat
group based in North Korea, often referred to as Hidden Cobra or the Lazarus
Group. SLICKSHOES provides backdoor functionality allowing the attackers remote
access to compromised systems using a custom C2 protocol.

## Network Setup

The dropper and backdoor binaries are Themida packed. Set up two virtual
machines and configure them to be on the same virtual network. Give the mock C2
server the C2 IP address from the malware.

## Links

* [https://www.us-cert.gov/ncas/analysis-reports/ar20-045b](https://www.us-cert.gov/ncas/analysis-reports/ar20-045b)

## IOCs

| Indicator                                                        | Type   | Context             |
|------------------------------------------------------------------|--------|---------------------|
| fdb87add07d3459c43cfa88744656f6c00effa6b7ec92cb7c8b911d233aeb4ac | SHA256 | SLICKSHOES Dropper  |
| 7250ccf4fad4d83d087a03d0dd67d1c00bf6cb8e7fa718140507a9d5ffa50b54 | SHA256 | SLICKSHOES Backdoor |
| 6fa550e835f796a2c95c2d8f2b4db88941cbc7a888e934478e88a274c70acca7 | SHA256 | SLICKSHOES Backdoor |
| 188.165.37.168                                                   | TCP/80 | SLICKSHOES C2       |
