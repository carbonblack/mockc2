# mockc2

An interactive mock C2 server

## Building

The `Makefile` will generate, macOS, Linux and Windows binaries in the `build` 
folder. Simply run the following command to build the binaries:

```
make all
```

## Usage

### Main Menu

```
mockc2> help
Main Menu Help

  debug       Enable or disable debug output [on/off]
  exit        Exit and shut down mockc2
  help        Print the main menu help
  interact    Interact with connected agents
  listener    Start or shutdown a protocol listener
  list        List connected agents
  version     Print the mockc2 server version
```

### Agent Menu

```
agent[631e7862c77c73a4fba9883359c430240af10b9191aabaac3ee8accbd58bddcd]> help
Agent Menu Help

  exec        Execute a command on the agent
  exit        Exit and shut down mockc2
  help        Print the agent menu help
  download    Download a file from the agent
  main        Return to the main menu
  upload      Upload a file to the agent
```

## Network Setup

Getting a sample to communicate correctly with the mock C2 server can be a
challenge. This brief introduction covers the general approach that can be taken
with any malware. Refer to the malware family documentation below for specific
details on how to configure the malware to communicate with the mock C2 server.

### DNS C2

Malware that makes use of domain names for their C2 addresses are the most
straight forward to work with. Once you have determined what the FQDN is that
is being used you can add an entry to the detonation machines hosts file. On
Linux and macOS the file is `/etc/hosts`. On Windows the file is located at
`c:\windows\system32\drivers\etc\hosts`.

### IP C2

Malware that makes use of hard coded IP addresses can be a little more
complicated to set up. There are two approaches you can take. 

If the IP address is stored as ASCII or Unicode you can simply edit the IP
address using a hex editor. This is usually a fairly quick process as you can
use the hex editor to search for the IP address and then just manually replace
it with the address of the mock C2 server. Just take care to make sure that you
keep the total number of bytes the same. This might require some padding with
null bytes.

For malware that is packed or that stores the IP in an encrypted format this is
usually harder to set up. The best thing to do is to use two separate machines
that you can set up on the same network. This is most easily accomplished with
two separate virtual machines on the same virtual network. Set up one machine
to have the IP address of the C2 server from the malware. This machine will run
the mock C2 server. On the second machine simply give it an IP address on the
same network. When the malware is run it will connect to your virtual machine
rather than trying to route out over the internet.

## Malware Families

* [BISTROMATH](docs/bistromath.md)
* [CROSSWALK](docs/crosswalk.md)
* [HOTCROISSANT](docs/hotcroissant.md)
* [MATA](docs/mata.md)
* [ObliqueRAT](docs/obliquerat.md)
* [RedXOR](docs/redxor.md)
* [Rifdoor](docs/rifdoor.md)
* [SLICKSHOES](docs/slickshoes.md)
* [Yort](docs/yort.md)
