package path

import (
	"fmt"
	"os"
)

func DataDir() string {
	if os.Geteuid() == 0 {
		return "/var/lib/mqtt-broker"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		panic("failed to find user data dir: " + err.Error())
	}
	return fmt.Sprintf("%s/.local/share/mqtt-broker", home)
}

func ServiceDataDir(nodeID string, service string) string {
	path := fmt.Sprintf("%s/%s/%s", DataDir(), nodeID, service)
	err := os.MkdirAll(path, 0750)
	if err != nil {
		panic("failed to build data dir: " + err.Error())
	}
	return path
}
