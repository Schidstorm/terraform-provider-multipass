package main

import "encoding/json"

func vmInfo(jsonContent []byte) (*VM, error) {
	vm := new(VM)
	err := json.Unmarshal(jsonContent, vm)
	if err != nil {
		return nil, err
	}
	return vm, nil
}
