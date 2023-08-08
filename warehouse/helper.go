package warehouse

import (
	"errors"
	"net"
	"os"
	"path"
	"strconv"
	"swallow/wrapper"
)

func markSuccess(file string) {
	f, err := os.Create(file)
	if err != nil {
		return
	}
	defer f.Close()
}

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

func listShards(dir string) []*Shard {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	ret := make([]*Shard, 0, len(files))
	for i := 0; i < len(files); i++ {
		if !files[i].IsDir() {
			continue
		}
		ts, err := strconv.ParseInt(files[i].Name(), 10, 64)
		if err != nil {
			continue
		}
		if _, err := os.Stat(path.Join(dir, files[i].Name(), success)); os.IsNotExist(err) {
			dataDir := path.Join(dir, files[i].Name())
			ret = append(ret, &Shard{
				ins:   wrapper.NewInstance(dataDir),
				path:  dataDir,
				start: ts,
			})
		}
	}

	return ret
}
