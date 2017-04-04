This guide is based on Ubuntu 16.04. But it should be adaptable to other modern
Linux distrubutions as well. Alternatively, for users more familiar with
containers, [CoreOS](https://coreos.com/) is a
[great option](https://github.com/songgao/dissertation) to run Squirrel
experiments.

## Install necessary packages

```
sudo apt-get update
sudo apt-get install docker.io etcd
```

The `apt-get` install, if successful, should have installed docker and etcd
packages, and also started and enabled the systemd units for them. Use
`systemctl` command (`systemctl status docker` and `systemctl status etcd`) to
verify.

## Configure Go environment

We are using a binary release from Go's official download source, so if you've
installed Go before using `apt-get`, and are not relying on that particular
version for anything, consider removing it first:

```
sudo apt-get --purge autoremove golang golang-go
```

Find a release download link at https://golang.org/dl/, then download the
tarball, untar, and place `go` in `$HOME`. 

```
cd $HOME
curl 'https://storage.googleapis.com/golang/go1.8.linux-amd64.tar.gz' | tar zx
```

This results in a `go` directory in `$HOME`. This should be your `$GOROOT`.

Create a file `.rc` in `$HOME` with following content for environment variables:

```
export GOROOT="$HOME/go"
export GOPATH="$HOME/gopath"
export PATH="$GOROOT/bin:$GOPATH/bin:$PATH"
```

Then add a line `source $HOME/.rc` to `$HOME/.bashrc` or whatever init script
your shell is using.

Open a new shell to make these variables effective. You can use `go get` to
install a package to verify Go is configured properly. For example,
`go get github.com/songgao/colorgo` should successfully install a `colorgo`
binary into your `$GOPATH/bin` directory, and `type colorgo` should print its
path.

## Get source code for squirrel and build binaries

```
go get -u github.com/squirrel-land/squirrel/squirrel-master
go get -u github.com/squirrel-land/squirrel/squirrel-worker
```

If the above two commands succeeded, you should have two binaries
`squirrel-master` and `squirrel-worker` in `$GOPATH/bin`. We won't be using
`squirrel-worker` binary directly, but use the `Makefile` described below to
build one suitable for running in containers.

## Add youself to docker group

You need to be in the group `docker` to have permissio to interact with the
docker daemon. Use the following command to add yourself to the group, and make
it effective:

```
sudo usermod -aG docker $USER
newgrp docker
```

This only needs to be done once. It'll persist throughout reboots. You can run
`docker ps` to verify it's working.


## Build docker images for squirrel worker

Each squirrel worker instance needs its own network namespace. We can achieve
it by using either a virtual machine for each worker node, or a Linux container
for each worker node. We are going with the latter approach here, using docker.
To run squirrel and other software in a container, we need to build a docker
image that has `squirrel-worker` binary.

There are many existing docker images that we can build on top of. At minimum,
a blank docker image that only includes the statically linked squirrel-worker,
and any programs you need to run on squirrel, would be sufficient. However,
it's not convinient to develop and design your experiment. For the purpose of
this guide, we build our docker image on top of an existing image (`alpine`),
which has a package manager built-in for installing various software packages.

See [docker's doc](https://docs.docker.com/engine/getstarted/step_four/#step-3-learn-about-the-build-process)
for more details on how to build an image. For the purpose of this guide, an
example `Dockerfile` and a `Makefile` are included in this directory. Use it to
build an image:

```
cd $GOPATH/src/github.com/squirrel-land/squirrel/getting-started
make
```

## Allow docker containers to access etcd

By default, etcd only allows access from localhost. Docker containers run on a
virtual bridge configured to a different network. We need to configure etcd to
allow access from this network. More specifically, etcd needs to listen on
the IP address that is in the network on the docker bridge. To find out this
address, use the `ip addr` command, and look for the IP address under
`bridge0`. If you haven't tweaked docker configuration, the default address for
the bridge is normally `172.17.0.1`.

Etcd access bind address is determined by two environmental variables,
`ETCD_LISTEN_CLIENT_URLS` and `ETCD_ADVERTISE_CLIENT_URLS`. Add following lines
to `/etc/default/etcd` to override them. Make sure not to leave out the
localhost:

```
ETCD_LISTEN_CLIENT_URLS="http://172.17.0.1:2379,http://172.17.0.1:4001,http://localhost:4001"
ETCD_ADVERTISE_CLIENT_URLS="http://172.17.0.1:2379,http://172.17.0.1:4001,http://localhost:4001"
```

Restart `etcd` to make it effective:

```
sudo systemctl restart etcd
```

## Run some iperf tests in squirrel

Before running squirrel, we need to populate configurations in etcd. Two
scripts `pop-etcd.passthrough` and `pop-etcd.csmaca` are included in the
getting-started guide. The former uses a passthrough model which does not
regulate traffic at all. It's useful to varify your configuration is correct,
and to probe the setup to find out the amoung of traffic it can handle. The
latter one configures it to a CSMA/CA model, which regulates the traffic based
on an internal model that mimics CSMA/CA.

### First let's use the passthrough:


1. Populate etcd configuration:

     ```
     cd $GOPATH/src/github.com/squirrel-land/squirrel/getting-started
     ./pop-etcd.passthrough
     ```

2. Then open a new shell (call it shell #2), and run:

     ```
     SQUIRREL_ENDPOINT=http://172.17.0.1:4001 squirrel-master
     ```
     
     This runs the squirrel-master on the machine natively (not in container).
     The `SQUIRREL_ENDPOINT` environmental variable tells squirrel-master which
     etcd endpoint endpoint to connect to.

3. Open a second shell (call it shell #3), and run:

     ```
     docker run --privileged --env SQUIRREL_ENDPOINT=http://172.17.0.1:4001 --detach squirrel-worker
     ```
     
     This runs the `squirrel-worker` image (that is built in previous steps) in
     a new docker container. Note that several flags are used here:
     `--privileged` tells docker to give this container higher privileges,
     among which `CAP_NET_ADMIN` is necessary to create a TUN/TAP interface;
     `--env` passes in an environment variable, in this case, the
     `SQUIRREL_ENDPOINT`, which tells squirrel-worker which etcd endpoint to
     connect to; The `--detach` flag tells docker to run this container in
     background, and causes docker to print an ID of the container that was
     created. Let's call this id [container-id-1].
     
     Now in the same shell (shell #3), run:
     
     ```
     docker exec --interactive --tty [container-id-1] /bin/bash
     ```
     
     (apparently replace [container-id-1] with proper ID from the last docker
     command)
     
     This gives you a bash shell inside the container. Since the container's
     entry point has been configured to `/bin/squirrel-worker` (see
     `Dockerfile`), `squirrel-worker` has already been running. So if you type
     `ip addr`, you'll see a `tap0` interface with an IP in `10.0.128.0/24`
     subnet (e.g.  `10.0.128.1`). This is the virtual address from our
     emulation. Let's call this address [ip-address-1].

4. Open a new shell (call it shell #4), and do the same thing from step 3. This
   gets you a new squirrel-worker container running ([container-id-2], with a
   shell inside the container, with [ip-address-2] on the `tap0` interface.

5. In shell #3, run `iperf -sui1`. This starts an iperf process in server mode
   (the `-s` flag), listening on UDP (the `-u` flag), and tells iperf to report
   statistics every 1 second (the `-i1` flag).

6. In shell #4, run `iperf -uc [ip-address-1] -b 300M`. This starts an iperf
   process in client mode connecting to the server running at [ip-address-1],
   i.e. the container running in shell #3, (the `-c [ip-address-1]` flag), and
   flood UDP traffic (the `-u` flag) at 300 Mbps (the `-b 300M` flag).

Now you'll see outputs from shell #3 printing the throughput at which it's
receiving the UDP traffic.

### Change the model to CSMA/CA

Stop all conatiners and process from the passthrough one, and run following:

```
cd $GOPATH/src/github.com/squirrel-land/squirrel/getting-started
./pop-etcd.csmaca
```

to configure squirrel to use the CSMA/CA model, and follow the rest steps (step
2-6) from the passthrough section. You should notice the traffic throughput
that iperf server is reporting is much lower than passthrough, due to the
emulated CSMA/CA model.

You may change parameters in the `pop-etcd.csmaca` script to emulate different
802.11 setups.
