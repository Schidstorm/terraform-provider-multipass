package main

import (
	"testing"
)

func TestVMInfoRunning(t *testing.T) {
	sampleContent := `{
		"errors": [
		],
		"info": {
			"test": {
				"disks": {
					"sda1": {
						"total": "5019643904",
						"used": "1051099136"
					}
				},
				"image_hash": "3b2e3aaaebf2bc364da70fbc7e9619a7c0bb847932496a1903cd4913cf9b1a26",
				"image_release": "18.04 LTS",
				"ipv4": [
					"10.191.95.144"
				],
				"load": [
					0,
					0,
					0
				],
				"memory": {
					"total": 1032937472,
					"used": 78082048
				},
				"mounts": {
				},
				"release": "Ubuntu 18.04.4 LTS",
				"state": "Running"
			}
		}
	}`

	vm, err := vmInfo([]byte(sampleContent))

	if err != nil {
		t.Error(err)
	}

	if info, ok := vm.Info["test"]; ok {
		assertEqual(t, "5019643904", info.Disks["sda1"].Total, "info.disks.sda1.total differs")
		assertEqual(t, "3b2e3aaaebf2bc364da70fbc7e9619a7c0bb847932496a1903cd4913cf9b1a26", info.ImageHash, "info.image_hash differs")
		assertEqual(t, "18.04 LTS", info.ImageRelease, "info.image_release differs")
		assertEqual(t, "10.191.95.144", info.Ipv4[0], "info.ipv4[0] differs")
		assertEqual(t, 1, len(info.Ipv4), "len(info.ipv4) differs")
		assertEqual(t, uint64(1032937472), info.Memory.Total, "info.memory.total differs")
		assertEqual(t, "Ubuntu 18.04.4 LTS", info.Release, "info.release differs")
		assertEqual(t, "Running", info.State, "info.state differs")
	} else {
		t.Error("vm.info.test missing")
	}
}

func TestVMInfoStopped(t *testing.T) {
	sampleContent := `{
		"errors": [
		],
		"info": {
			"test": {
				"disks": {
					"sda1": {
					}
				},
				"image_hash": "3b2e3aaaebf2bc364da70fbc7e9619a7c0bb847932496a1903cd4913cf9b1a26",
				"image_release": "18.04 LTS",
				"ipv4": [
				],
				"load": [
				],
				"memory": {
				},
				"mounts": {
					"/tmp": {
						"gid_mappings": [
							"300:400",
							"100:200"
						],
						"source_path": "/tmp",
						"uid_mappings": [
							"100:200"
						]
					},
					"/tmp/abs": {
						"gid_mappings": [
							"300:400",
							"100:200"
						],
						"source_path": "/tmp",
						"uid_mappings": [
							"100:200"
						]
					}
				},
				"release": "",
				"state": "Stopped"
			}
		}
	}`

	vm, err := vmInfo([]byte(sampleContent))

	if err != nil {
		t.Error(err)
	}

	if info, ok := vm.Info["test"]; ok {
		assertEqual(t, "3b2e3aaaebf2bc364da70fbc7e9619a7c0bb847932496a1903cd4913cf9b1a26", info.ImageHash, "info.image_hash differs")
		assertEqual(t, "18.04 LTS", info.ImageRelease, "info.image_release differs")
		assertEqual(t, 0, len(info.Ipv4), "len(info.ipv4) differs")
		assertEqual(t, "Stopped", info.State, "info.state differs")

		//"/tmp" mapping
		assertEqual(t, "/tmp", info.Mounts["/tmp"].SourcePath, "info.mounts./tmp.source_path differs")
		assertEqual(t, "300:400", info.Mounts["/tmp"].GIDMappings[0], "info.mounts./tmp.gid_mappings[0] differs")
		assertEqual(t, "100:200", info.Mounts["/tmp"].GIDMappings[1], "info.mounts./tmp.gid_mappings[1] differs")
		assertEqual(t, 2, len(info.Mounts["/tmp"].GIDMappings), "len(info.mounts./tmp.gid_mappings) differs")
		assertEqual(t, "100:200", info.Mounts["/tmp"].UIDMappings[0], "info.mounts./tmp.uid_mappings[0] differs")
		assertEqual(t, 1, len(info.Mounts["/tmp"].UIDMappings), "len(info.mounts./tmp.uid_mappings) differs")

		//"/tmp/abs" mapping
		assertEqual(t, "/tmp", info.Mounts["/tmp/abs"].SourcePath, "info.mounts./tmp/abs.source_path differs")
		assertEqual(t, "300:400", info.Mounts["/tmp/abs"].GIDMappings[0], "info.mounts./tmp/abs.gid_mappings[0] differs")
		assertEqual(t, "100:200", info.Mounts["/tmp/abs"].GIDMappings[1], "info.mounts./tmp/abs.gid_mappings[1] differs")
		assertEqual(t, 2, len(info.Mounts["/tmp/abs"].GIDMappings), "len(info.mounts./tmp/abs.gid_mappings) differs")
		assertEqual(t, "100:200", info.Mounts["/tmp/abs"].UIDMappings[0], "info.mounts./tmp/abs.uid_mappings[0] differs")
		assertEqual(t, 1, len(info.Mounts["/tmp/abs"].UIDMappings), "len(info.mounts./tmp/abs.uid_mappings) differs")
	} else {
		t.Error("vm.info.test missing")
	}
}

func assertEqual(t *testing.T, should, is interface{}, message string) {
	if should != is {
		t.Error(message)
	}
}
