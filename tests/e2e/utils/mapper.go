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
	"fmt"
	log "k8s.io/klog"
	"os/exec"
	"time"
)

func MakeMapperImages(makeMapperImage, checkMapperImage string) error {
	// build mapper image
	cmd := exec.Command("sh", "-c", makeMapperImage)
	if err := PrintCmdOutput(cmd); err != nil {
		return err
	}

	// check images exist
	log.Info("checking mapper images exists")
	cmd = exec.Command("sh", "-c", checkMapperImage)
	if err := PrintCmdOutput(cmd); err != nil {
		return err
	}
	return nil
}

//run mapper
func RunMapper(runMapper, devicesProfilesname string, checkMapperRun string) error {
	log.Info("run mapper image on docker")
	time.Sleep(1 * time.Second)
	runMapper = fmt.Sprintf(runMapper, devicesProfilesname)
	cmd := exec.Command("sh", "-c", runMapper)
	if err := PrintCmdOutput(cmd); err != nil {
		return err
	}

	time.Sleep(5 * time.Second)

	log.Info("check weather mapper run successfully")
	cmd = exec.Command("sh", "-c", checkMapperRun)
	if err := PrintCmdOutput(cmd); err != nil {
		return err
	}
	return nil
}

//stop mapper container
func StopAndDeleteMapper(getContainerID string) error {
	log.Info("stop mapper running")
	cmd := exec.Command("sh", "-c", getContainerID)
	result, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	containerID := string(result[:12])

	log.Info("stop and delete mapper container")
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

//delete mapper image
func DeleteMapperImage(deleteMapperImage string) error {
	cmd := exec.Command("sh", "-c", deleteMapperImage)
	if err := PrintCmdOutput(cmd); err != nil {
		return err
	}
	return nil
}
