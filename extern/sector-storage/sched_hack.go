package sectorstorage

import (
	"fmt"
	"github.com/filecoin-project/go-state-types/abi"
	"io/ioutil"
	"os"
	"path/filepath"
)

func SetSectorPreferredHostname(sectorID abi.SectorNumber, hostname string) {
	path := fmt.Sprintf("/var/tmp/lotus-miner-mapping/%d.txt", sectorID)
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		log.Errorf("[Hack] Creating directory failed: %s", err.Error())
		return
	}

	err = ioutil.WriteFile(path, []byte(hostname), 0644)
	if err != nil {
		log.Errorf("[Hack] Writing file(%s) failed: %s", path, err.Error())
		return
	}

	log.Infof("[Hack] Setting sector %d to host %s", sectorID, hostname)
}

func GetSectorPreferredHostname(sectorID abi.SectorNumber) string {
	path := fmt.Sprintf("/var/tmp/lotus-miner-mapping/%d.txt", sectorID)

	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Debugf("[Hack] Reading file(%s) failed: %s", path, err.Error())
		return ""
	}

	log.Infof("[Hack] Getting sector %d's preferred host %s", sectorID, string(content))
	return string(content)
}

func GetPc1Running(hostname string, exclude abi.SectorNumber) int {
	path := fmt.Sprintf("/var/tmp/lotus-miner-mapping/%s/pc1", hostname)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Errorf("[Hack] Reading directory(%s) failed: %s", path, err.Error())
		return 0
	}

	result := 0
	for _, file := range files {
		if file.Name() != fmt.Sprint(exclude) {
			result = result + 1
		}
	}

	return result
}

func AddPc1Running(hostname string, sectorID abi.SectorNumber) {
	path := fmt.Sprintf("/var/tmp/lotus-miner-mapping/%s/pc1/%d", hostname, sectorID)
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		log.Errorf("[Hack] Creating directory failed: %s", err.Error())
		return
	}

	err = ioutil.WriteFile(path, []byte(""), 0644)
	if err != nil {
		log.Errorf("[Hack] Writing file(%s) failed: %s", path, err.Error())
		return
	}

	log.Infof("[Hack] Adding host %s for pc1 job on sector %d", hostname, sectorID)
}

func RemovePc1Running(hostname string, sectorID abi.SectorNumber) {
	path := fmt.Sprintf("/var/tmp/lotus-miner-mapping/%s/pc1/%d", hostname, sectorID)
	err := os.Remove(path)
	if err != nil {
		log.Errorf("[Hack] Removing file(%s) failed: %s", path, err.Error())
		return
	}

	log.Infof("[Hack] Removing host %s for pc1 job on sector %d", hostname, sectorID)
}