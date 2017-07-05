// network子包提供网络相关的一些小函数
package network

import (
	"errors"
	"net"
	"strings"
)

// 获取本地IP地址，除loopback地址(127.0.0.1)
func GetLocalIPAddrs() (ipaddrs []string, err error) {
	var (
		addrs []net.Addr
		addr  []net.Addr
	)

	// 获取所有网络接口
	i, err := net.Interfaces()
	if err != nil {
		return
	}

	// 循环每个网络接口,如果一个接口获取addrs失败，不返回，而是继续
	// 这样可以保证获取可用addrs
	for _, iface := range i {
		if strings.Contains(strings.ToLower(iface.Flags.String()), "loopback") {
			continue
		}
		addr, err = iface.Addrs()
		if err != nil {
			continue
		}
		addrs = append(addrs, addr...)
	}

	if len(addrs) == 0 {
		err = errors.New("can not get any addrs")
	}

	ipaddrs = make([]string, 0)
	for _, addr := range addrs {
		inet, ok := addr.(*net.IPNet)
		if ok {
			ip := inet.IP.To4()
			if ip != nil {
				ipaddrs = append(ipaddrs, ip.String())
			}
		}
	}
	return
}
