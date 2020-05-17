package main

type argumentMount struct {
	SourcePath  string
	TargetPath  string
	GIDMappings []string
	UIDMappings []string
}

type argumentMounts = []argumentMount

func convertArgumentMounts(mounts [](interface{})) argumentMounts {
	resultMounts := argumentMounts{}
	type mountType = map[string](interface{})
	for _, imount := range mounts {
		mount := imount.(map[string]interface{})
		iuidMaps := mount["uid_mappings"].([]interface{})
		uidMaps := []string{}
		for _, m := range iuidMaps {
			uidMaps = append(uidMaps, m.(string))
		}

		igidMaps := mount["gid_mappings"].([]interface{})
		gidMaps := []string{}
		for _, m := range igidMaps {
			gidMaps = append(gidMaps, m.(string))
		}

		resultMounts = append(resultMounts, argumentMount{
			SourcePath:  mount["source_path"].(string),
			TargetPath:  mount["target_path"].(string),
			UIDMappings: uidMaps,
			GIDMappings: gidMaps,
		})
	}

	return resultMounts
}
