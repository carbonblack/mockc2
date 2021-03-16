# ObliqueRat

ObliqueRAT is a malicious remote access trojan (RAT) that is distributed using malicious Microsoft Office documents. ObliqueRAT is linked to the Transparent Tribe APT group targeting organizations in South Asia. It's a simple RAT that makes use of a mostly text based protocol and supports a handful of commands.

## Network Setup

In the older samples, the C2 IP address is stored as an ASCII string. Use a hex editor to find and replace the C2 IP address with the IP address of your mock C2 server.

## Links

* [https://blog.talosintelligence.com/2020/02/obliquerat-hits-victims-via-maldocs.html](https://blog.talosintelligence.com/2020/02/obliquerat-hits-victims-via-maldocs.html)
* [https://blog.talosintelligence.com/2021/02/obliquerat-new-campaign.html](https://blog.talosintelligence.com/2021/02/obliquerat-new-campaign.html)

## IOCs

| Indicator                                                        | Type     | Context                      |
|------------------------------------------------------------------|----------|------------------------------|
| 37c7500ed49671fe78bd88afa583bfb59f33d3ee135a577908d633b4e9aa4035 | SHA256   | ObliqueRAT 32-bit executable |
| 9da1a55b88bda3810ccd482051dc7e0088e8539ef8da5ddd29c583f593244e1c | SHA256   | ObliqueRAT 32-bit executable |
| 0ade4e834f34ed7693ebbe0354c668a6cb9821de581beaf1f3faae08150bd60d | SHA256   | ObliqueRAT 32-bit executable |
| 185.117.73.222                                                   | TCP/3344 | ObliqueRAT C2                |
| 185.183.98.182                                                   | TCP/4701 | ObliqueRAT C2                |
