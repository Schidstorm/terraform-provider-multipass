package main

type VM struct {
	Errors []string
	Info   map[string]VMInfo
}

type VMInfo struct {
	Disks        map[string]VMDisk
	ImageHash    string `json:"image_hash"`
	ImageRelease string `json:"image_release"`
	Ipv4         []string
	Memory       VMMemory
	Mounts       map[string]VMMount
	Release      string
	State        string
}

type VMMount struct {
	SourcePath  string   `json:"source_path"`
	GIDMappings []string `json:"gid_mappings"`
	UIDMappings []string `json:"uid_mappings"`
}

type VMDisk struct {
	Total string
}

type VMMemory struct {
	Total uint64
}
