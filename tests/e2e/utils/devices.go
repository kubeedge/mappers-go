/*
Copyright 2021 The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	log "k8s.io/klog/v2"
	"os/exec"
)

func MakeDeviceImage(makeDevice string, checkDeviceImage string) error {
	cmd := exec.Command("sh", "-c", makeDevice)
	if err := PrintCmdOutput(cmd); err != nil {
		return err
	}
	cmd = exec.Command("sh", "-c", checkDeviceImage)
	if err := PrintCmdOutput(cmd); err != nil {
		return err
	}
	return nil
}

func DeleteDeviceImage(deleteDeviceImage string) error {
	cmd := exec.Command("sh", "-c", deleteDeviceImage)
	if err := PrintCmdOutput(cmd); err != nil {
		return err
	}
	return nil
}

// run device container
func StartDevices(runDeviceContainer string, checkDeviceRun string) error {
	cmd := exec.Command("sh", "-c", runDeviceContainer)
	if err := PrintCmdOutput(cmd); err != nil {
		return err
	}

	// confirm whether the device can start normally
	cmd = exec.Command("sh", "-c", checkDeviceRun)
	if err := PrintCmdOutput(cmd); err != nil {
		return err
	}
	return nil
}

//stop mapper container
func StopAndDeleteDevice(getContainerID string) error {
	log.Info("stop device running")
	cmd := exec.Command("sh", "-c", getContainerID)
	result, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	containerID := string(result[:12])

	log.Info("stop and delete device container")
	cmd = exec.Command("sh", "-c", "docker stop "+containerID)
	if err = PrintCmdOutput(cmd); err != nil {
		return err
	}
	cmd = exec.Command("sh", "-c", "docker rm "+containerID)
	if err = PrintCmdOutput(cmd); err != nil {
		return err
	}
	return nil
}
