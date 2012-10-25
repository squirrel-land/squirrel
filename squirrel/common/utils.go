package common

import (
	"net"
)

func IPFromBigEndian(bytes []byte) net.IP {
	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0]).To4()
}

func IPMaskFromBigEndian(bytes []byte) net.IPMask {
	return net.IPv4Mask(bytes[3], bytes[2], bytes[1], bytes[0])
}

func IPFrom(bytes []byte) net.IP {
	return net.IPv4(bytes[0], bytes[1], bytes[2], bytes[3]).To4()
}

func IPMaskFrom(bytes []byte) net.IPMask {
	return net.IPv4Mask(bytes[0], bytes[1], bytes[2], bytes[3])
}
