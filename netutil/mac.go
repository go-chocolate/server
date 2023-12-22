package netutil

import "net"

func GetMacAddress() ([]string, error) {
	var macAddrs []string
	netInterfaces, err := net.Interfaces()
	if err != nil {
		//fmt.Printf("fail to get net interfaces: %v", err)
		return macAddrs, err
	}

	for _, netInterface := range netInterfaces {
		macAddr := netInterface.HardwareAddr.String()
		if len(macAddr) == 0 {
			continue
		}

		macAddrs = append(macAddrs, macAddr)
	}
	return macAddrs, nil
}
