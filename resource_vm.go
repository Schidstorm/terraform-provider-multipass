package main

import (
	"errors"
	"log"
	"os/exec"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const checkAliveTTL = 30

func resourceVM() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerCreate,
		Read:   resourceServerRead,
		Update: resourceServerUpdate,
		Delete: resourceServerDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cpus": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"image_hash": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"disk_size": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"image_release": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv4": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
				Optional: true,
			},
			"memory_size": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"release": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"mounts": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_path": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"target_path": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"uid_mappings": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"gid_mappings": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
				Optional: true,
			},
		},
	}
}

func resourceServerCreate(d *schema.ResourceData, m interface{}) error {
	serverName, ok := d.Get("name").(string)
	if !ok || serverName == "" {
		return errors.New("invalid server name")
	}
	d.SetId(serverName)

	parts := []string{"launch", "--name", serverName}
	if val, ok := d.GetOk("cpus"); ok {
		parts = append(parts, "--cpus", string(val.(int)))
	}
	if val, ok := d.GetOk("memory_size"); ok {
		parts = append(parts, "--mem", string(val.(int)))
	}
	if val, ok := d.GetOk("disk_size"); ok {
		parts = append(parts, "--disk", string(val.(string)))
	}

	cmd := exec.Command("multipass", parts...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New("command multipass launch failed: " + string(output))
	}

	ttl := checkAliveTTL
	for vm, _ := getVMInfo(serverName); vm == nil || vm.Info[serverName].State == "Starting"; time.Sleep(1 * time.Second) {
		if ttl <= 0 {
			return errors.New("waiting for server start timed out")
		}
		ttl--
	}

	for _, mount := range convertArgumentMounts(d.Get("mounts").([]interface{})) {
		err := cmdMount(
			serverName,
			mount.SourcePath,
			mount.TargetPath,
			mount.UIDMappings,
			mount.GIDMappings,
		)

		if err != nil {
			return err
		}
	}

	err = resourceServerRead(d, m)
	if err != nil {
		return err
	}

	if val, ok := d.GetOk("state"); ok && val == "Stopped" {
		cmd = exec.Command("multipass", "stop", serverName)
		_, err = cmd.Output()
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceServerRead(d *schema.ResourceData, m interface{}) error {
	serverName, ok := d.Get("name").(string)
	if !ok || serverName == "" {
		return errors.New("invalid server name")
	}

	vm, err := getVMInfo(serverName)
	if err != nil {
		return err
	}

	info := vm.Info[serverName]
	d.Set("image_hash", info.ImageHash)

	return nil
}

func getVMInfo(serverName string) (*VM, error) {
	cmd := exec.Command("multipass", "info", "--format", "json", serverName)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return vmInfo(output)
}

func cmdMount(serverName, source, target string, uidMappings, gidMappings []string) error {
	arguments := []string{
		"mount",
	}
	for _, uid := range uidMappings {
		arguments = append(arguments, "--uid-map", uid)
	}
	for _, gid := range gidMappings {
		arguments = append(arguments, "--gid-map", gid)
	}
	arguments = append(arguments, source, serverName+":"+target)
	cmd := exec.Command("multipass", arguments...)
	return cmd.Run()
}

func cmdUmount(serverName, mount string) error {
	cmd := exec.Command("multipass", "umount", serverName+":"+mount)
	return cmd.Run()
}

func resourceServerUpdate(d *schema.ResourceData, m interface{}) error {
	serverName, ok := d.Get("name").(string)
	if !ok || serverName == "" {
		return errors.New("invalid server name")
	}

	if d.HasChange("state") {
		_, newi := d.GetChange("state")
		new := newi.(string)
		var err error = nil

		if new == "Started" {
			cmd := exec.Command("multipass", "start", serverName)
			err = cmd.Run()
		} else {
			cmd := exec.Command("multipass", "stop", serverName)
			err = cmd.Run()
		}

		if err != nil {
			return err
		}
	}

	if d.HasChange("mounts") {
		old, new := d.GetChange("mounts")

		oldMap := listMountsToMapMounts(convertArgumentMounts(old.([]interface{})))
		newMap := listMountsToMapMounts(convertArgumentMounts(new.([]interface{})))

		mountsToDelete := subtractMountMaps(oldMap, newMap)
		mountsToAdd := subtractMountMaps(newMap, oldMap)
		mountsToUpdate := subtractMountMaps(newMap, mountsToAdd)

		for key := range mountsToUpdate {
			updateOld := oldMap[key]
			updateNew := newMap[key]

			if !reflect.DeepEqual(updateOld, updateOld) {
				mountsToDelete[key] = updateOld
				mountsToAdd[key] = updateNew
			}
		}

		for _, mount := range mountsToDelete {
			err := cmdUmount(serverName, mount.TargetPath)
			log.Println(err)
		}

		for _, mount := range mountsToAdd {
			err := cmdMount(
				serverName,
				mount.SourcePath,
				mount.TargetPath,
				mount.UIDMappings,
				mount.GIDMappings,
			)

			log.Println(err)
		}
	}

	return resourceServerRead(d, m)
}

func listMountsToMapMounts(mounts argumentMounts) map[string]argumentMount {
	mountMap := map[string]argumentMount{}
	for _, mount := range mounts {
		mountMap[mount.TargetPath] = mount
	}

	return mountMap
}

func subtractMountMaps(lhs map[string]argumentMount, rhs map[string]argumentMount) map[string]argumentMount {
	result := map[string]argumentMount{}
	for key, val := range lhs {
		if _, ok := rhs[key]; !ok {
			result[key] = val
		}
	}

	return result
}

func resourceServerDelete(d *schema.ResourceData, m interface{}) error {
	serverName, ok := d.Get("name").(string)
	if !ok || serverName == "" {
		return errors.New("invalid server name")
	}
	cmd := exec.Command("multipass", "delete", "-p", serverName)
	_, err := cmd.Output()
	if err != nil {
		return err
	}

	return nil
}
