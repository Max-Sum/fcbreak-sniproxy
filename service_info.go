package main

import (
	"encoding/json"
	"os"

	"github.com/Max-Sum/fcbreak"
)

func readServiceInfo(path string) (*fcbreak.ServiceInfo, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	info := &fcbreak.ServiceInfo{}
	err = json.Unmarshal(bytes, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}