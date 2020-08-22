# Rifdoor

According to AhnLab, Rifdoor dates back to a 2015 attack on exhibitors in the
Seoul International Aerospace & Defense Exhibition (ADEX). It was sent to
exhibitors in an email with an Excel or Word document containing macros,
pretending to be from the organizer of the event. This trojan continued to be
seen in attacks well into 2016. Rifdoor is a basic remote access trojan. 

## Network Setup

The C2 IP address is encrypted with a single byte key of `0xF`. Take the
original C2 IP address as well as your mock C2 IP address and encrypt the
strings with the single byte key of `0xF`. Then use a hex editor to find and
replace the encrypted C2 IP address with the encrypted IP address of your mock
C2 server.

## Links

* [https://www.carbonblack.com/blog/vmware-carbon-black-tau-threat-analysis-the-evolution-of-lazarus/](https://www.carbonblack.com/blog/vmware-carbon-black-tau-threat-analysis-the-evolution-of-lazarus/)
* [https://global.ahnlab.com/global/upload/download/techreport/%5BAhnLab%5DAndariel_a_Subgroup_of_Lazarus%20(3).pdf](https://global.ahnlab.com/global/upload/download/techreport/%5BAhnLab%5DAndariel_a_Subgroup_of_Lazarus%20(3).pdf)

## IOCs

| Indicator                                                        | Type               | Context                   |
|------------------------------------------------------------------|--------------------|---------------------------|
| a9915977c810fb2d61be8ff9d177de4d10bd3b24bdcbb3bb8ab73bcfdc501995 | SHA256             | Rifdoor 32-bit executable |
| 57d1df9f6c079e67e883a25cfbb124d33812b5fcdb6288977c4b8ebc1c3350de | SHA256             | Rifdoor 32-bit executable |
| 0a0c09f81a3fac2af99fab077e8c81a6674adc190a1077b04e2956f1968aeff3 | SHA256             | Rifdoor 32-bit executable |
| c9455e218220e81670ddd3c534011a68863ca9e09ab8215cc72da543ca910b81 | SHA256             | Rifdoor 32-bit executable |
| 192.99.223.115                                                   | TCP/80<br/>TCP/443 | Rifdoor C2                |
| 165.194.123.67                                                   | TCP/8008           | Rifdoor C2                |
| 111.68.7.74                                                      | TCP/443            | Rifdoor C2                |
