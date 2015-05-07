package common

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
)

func GetEtcdValue(client *etcd.Client, key string) (value string, err error) {
	var resp *etcd.Response
	resp, err = client.Get(key, false, false)
	if err != nil {
		return
	}
	if resp.Node.Dir {
		err = fmt.Errorf("%s is a Dir (expected a value)", key)
		return
	}
	if resp.Node.Key != key {
		err = fmt.Errorf("Unexpected node returned from etcd. (expected %s, got %s)", key, resp.Node.Key)
		return
	}
	value = resp.Node.Value
	return
}

func IsEtcdNotFoundError(err error) bool {
	etcdErr, ok := err.(*etcd.EtcdError)
	if !ok {
		return false
	}
	// https://github.com/coreos/etcd/blob/f1ed69e8838548e7226250555598a97fd9f9bc52/error/error.go#L78
	return etcdErr.ErrorCode == 100
}
