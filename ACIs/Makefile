.PHONY: build_all clean %.layout

build_all : squirrel-worker.aci squirrel-worker.aci.asc squirrel-master.aci squirrel-master.aci.asc

clean : 
	rm -f squirrel-worker.aci squirrel-worker.aci.asc squirrel-master.aci squirrel-master.aci.asc
	rm -rf squirrel-worker.layout squirrel-master.layout

%.aci.asc : %.aci
	gpg --armor --detach-sign --yes $<

%.aci : %.layout
	actool build --overwrite $< $@
	rm -rf $<

%.layout : templates/% ../% /bin/busybox
	rm -rf $*.layout
	cp -R templates/$* $@
	CGO_ENABLED=0 GOOS=linux go build -o ./$*.layout/rootfs/bin/$* -a -installsuffix cgo ../$*
	cp /bin/busybox ./$*.layout/rootfs/bin/busybox
	cp /bin/busybox ./$*.layout/rootfs/bin/ip
	cp /bin/busybox ./$*.layout/rootfs/bin/sh
