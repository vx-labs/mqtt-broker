package consistency

import (
	fmt "fmt"
	"os"
)

func dataDir() string {
	if os.Geteuid() == 0 {
		return "/var/lib/mqtt-broker"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		panic("failed to find user data dir: " + err.Error())
	}
	return fmt.Sprintf("%s/.local/share/mqtt-broker", home)
}

func buildDataDir(id string) string {
	path := fmt.Sprintf("%s/raft-%s", dataDir(), id)
	err := os.MkdirAll(path, 0750)
	if err != nil {
		panic("failed to build data dir: " + err.Error())
	}
	return path
}
