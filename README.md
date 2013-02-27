# squirrel
Squirrel is a platform that emulates 802.11 networks over ethernet. It helps testing wireless network applications and user-space routing daemons. Unlike network simulators like ns-3 in which Applications or routing daemons are written in simulation script, those running with squirrel are **real programs** that can actually also run on a Laptop, cell phone, or any other real mobile devices that run Linux[1].


Squirrel works above *Data Link Layer (Layer 2)* and (slightly) below *Network Layer (Layer 3)* in OSI. It creates TUN interface that programs see as a network interface just like `wlan0`. Squirrel handles the TUN interface and bridges multiple nodes into a virtual 802.11 wireless network, and applies packet loss, traffic shaping, etc. to IP packets sent through it.

Squirrel has a simple plugin mechanism that enables developing models. There are two types of models, `MobilityManager` and `September`[2] (see [models/common](http://godoc.org/github.com/songgao/squirrel/models/common)). `MobilityManager` assigns and updates virtual position for each mobile node; `September` decides for each packet whether it should be delivered or not.

[1]: It's possible for other unix but currently it only supports Linux.  
[2]: The name *September* is from a science fiction TV series [*Fringe*](http://en.wikipedia.org/wiki/Fringe_(TV_series)), in which *September* is the name of an *Observer* who helped save humanity.

## Install
If you don't have Go (golang) installed, get it from package manager of your distribution
```bash
$ sudo apt-get install golang  # Ubuntu
```
```bash
$ sudo pacman -S go  # Arch Linux
```
or follow [Go Getting Started guide](http://golang.org/doc/install)

Installing squirrel should be quite straightforward:
```bash
$ go get -u github.com/songgao/squirrel # You might need sudo here as well
```
If the command finishes OK, check the installation by executing `squirrel` command:
```bash
$ squirrel
```
You are ready to go if it prints something like
```
  -c=false: Run as client.
  -f="": Configuration file.
  -m=false: Run as master.
```

## Usage
In squirrel, master node can be on the same host as one of the client nodes, but different client nodes cannot run on a same host. You need at least two Linux hosts to try out squirrel. This will enable a configuration with one master and two clients. Either real hosts or virtual machines will work, but make sure they are inter-connected with ethernet at least capable of 100 Mbps.

Create configuration files with following content or get the examples from squirrel source (`client.conf.json.example` and `master.conf.json.example`):

`master.conf.json`
```json
{
    "ListenAddress":                ":1234",
    "Network":                      "10.0.0.0/24",

    "MobilityManager":              "StaticUniformPositions",
    "MobilityManagerParameters":    {
        "Spacing": 5000,
        "Shape": "Linear"
    },

    "September":                    "September2nd",
    "SeptemberParameters":          {
        "LowestZeroPacketDeliveryDistance": 120000,
        "InterferenceRange":                250000
    }
}
```

`client.conf.json`
```json
{
    "ServerAddress":    "192.168.2.254:1234",
    "TunInterfaceName": "i.am.not.real",
    "Identity":         1
}
```

Configuration files are in `json` format. If there's anything wrong with the file, squirrel will complain about it and print help messages on parameters.

Tips for `client.conf.json`:

1. You may need to change `ServerAddress` field to your actual IP address of master node.
2. `Identity` field needs to be unique for each client. You could copy it to another host, and change `1` into `2`.

To run the master node:
```bash
$ squirrel -m -f master.conf.json
```

To run a client node:
```bash
$ sudo squirrel -c -f client.conf.json
```

Make sure the two clients are not on the same host. Everything is OK if master node prints something like this:
```
10.0.0.1 joined
10.0.0.2 joined
```

You may check the interface with command `ip addr`:
```
3: i.am.not.real: <POINTOPOINT,MULTICAST,NOARP,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UNKNOWN qlen 500
    link/none 
    inet 10.0.0.1/24 scope global i.am.not.real
```
That's the interface created by squirrel client.

Now you can use various tools to test the network:

**ping**
```bash
$ # on 10.0.0.1
$ ping -b 10.0.0.255
WARNING: pinging broadcast address
PING 10.0.0.255 (10.0.0.255) 56(84) bytes of data.
64 bytes from 10.0.0.1: icmp_req=1 ttl=64 time=0.044 ms
64 bytes from 10.0.0.2: icmp_req=1 ttl=64 time=1.85 ms (DUP!)
64 bytes from 10.0.0.1: icmp_req=2 ttl=64 time=0.051 ms
64 bytes from 10.0.0.2: icmp_req=2 ttl=64 time=1.99 ms (DUP!)
^C
--- 10.0.0.255 ping statistics ---
2 packets transmitted, 2 received, +2 duplicates, 0% packet loss, time 1001ms
rtt min/avg/max/mdev = 0.044/0.984/1.991/0.938 ms
```
_* You probably need this on each client to enable broadcast ping: `sudo bash -c "echo 1 > /proc/sys/net/ipv4/icmp_echo_ignore_broadcasts"`_

**iperf**
```bash
$ # on 10.0.0.2
$ iperf -s
------------------------------------------------------------
Server listening on TCP port 5001
TCP window size: 85.3 KByte (default)
------------------------------------------------------------
```
```bash
$ # on 10.0.0.1
$ iperf -c 10.0.0.2 -i1 -t5
------------------------------------------------------------
Client connecting to 10.0.0.2, TCP port 5001
TCP window size: 23.5 KByte (default)
------------------------------------------------------------
[  3] local 10.0.0.1 port 58812 connected with 10.0.0.2 port 5001
[ ID] Interval       Transfer     Bandwidth
[  3]  0.0- 1.0 sec  1.62 MBytes  13.6 Mbits/sec
[  3]  1.0- 2.0 sec   512 KBytes  4.19 Mbits/sec
[  3]  2.0- 3.0 sec  1.38 MBytes  11.5 Mbits/sec
[  3]  3.0- 4.0 sec  1.00 MBytes  8.39 Mbits/sec
[  3]  4.0- 5.0 sec   896 KBytes  7.34 Mbits/sec
[  3]  0.0- 5.3 sec  5.50 MBytes  8.68 Mbits/sec
```

You may change `Spacing` parameter in the mobility manager to make nodes farther away from each other.

## Future Work in Plan
- A mobility manager in which nodes are actually mobile
- Mobility managers that can read SUMO traces
- Performance tuning
- More `September` models
- Larger scale tests
- Supporting olsrd, xorp, etc.

## License
`squirrel` is licensed under GNU GENERAL PUBLIC LICENSE Version 3
