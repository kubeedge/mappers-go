/*
Copyright 2024 The KubeEdge Authors.
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

package missions

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	klog "k8s.io/klog/v2"

	"github.com/kubeedge/mappers-go/mappers/windows-virtual-exec/internal/core/model"
	"github.com/kubeedge/mappers-go/mappers/windows-virtual-exec/internal/core/mqtt"
	"github.com/kubeedge/mappers-go/mappers/windows-virtual-exec/internal/core/store"
	"github.com/kubeedge/mappers-go/mappers/windows-virtual-exec/internal/utils/encode"
)

const (
	StatusOK      = "ok"
	StatusError   = "error"
	StatusWaiting = "waiting"
	StatusWorking = "working"
)

type Mission struct {
	exec   *Command
	Config MissionConfig `json:"config"`
	Status string        `json:"status"` // ok, error, waiting, working
	Output string        `json:"output"`
}

type MissionConfig struct {
	UniqueName       string `json:"uniqueName"` // use as device id
	Command          string `json:"command"`
	FileContent      string `json:"fileContent"`
	FileName         string `json:"fileName"`
	WorkingDirectory string `json:"workingDirectory"`
}

var cache = sync.Map{}
var createMutex sync.Mutex

// NewMission add a new mission in memory cache
func NewMission(config MissionConfig) (client *Mission, err error) {
	defer func() {
		go client.Run()
	}()
	// load cached mission
	if data, ok := cache.Load(config.UniqueName); ok {
		klog.Info("Get mission from cache: ", config.UniqueName)
		return data.(*Mission), nil
	}

	// prevent New operation simutaneously in different goroutines
	createMutex.Lock()
	defer createMutex.Unlock()

	// not in memory, add a new mission
	client = &Mission{
		exec:   &Command{},
		Config: config,
	}

	var mission model.Mission
	if store.DB.Model(&mission).Where("unique_name = ?", config.UniqueName).Find(&mission).RowsAffected > 0 && mission.Status != StatusWorking {
		client.Config.UniqueName = mission.UniqueName
		client.Config.Command = mission.Command
		client.Config.FileContent = mission.FileContent
		client.Config.FileName = mission.FileName
		client.Config.WorkingDirectory = mission.WorkingDirectory
		client.Status = mission.Status
		client.Output = mission.Output
		klog.Info("Get mission from db: ", config.UniqueName)
		return client, nil
	}

	client.exec.Cmd = exec.Command("powershell", "-c", client.Config.Command)
	client.exec.Cmd.Dir = client.Config.WorkingDirectory
	client.Status = StatusWaiting

	if mission.Status == StatusWorking {
		client.UpdateDB()
	} else {
		client.InsertDB()
	}

	cache.Store(config.UniqueName, client)
	client.ReportDeviceStatus()
	klog.Info("New mission: ", config.UniqueName)
	return client, nil
}

func RemoveMission(id string) {
	createMutex.Lock()
	defer createMutex.Unlock()

	var mission model.Mission
	if store.DB.Model(&mission).Where("unique_name = ?", id).Find(&mission).RowsAffected == 0 {
		return
	}

	cache.Delete(id)
}

func (c *Mission) InsertDB() {
	err := store.DB.Create(&model.Mission{
		UniqueName:       c.Config.UniqueName,
		Command:          c.Config.Command,
		FileContent:      c.Config.FileContent,
		FileName:         c.Config.FileName,
		WorkingDirectory: c.Config.WorkingDirectory,
		Status:           c.Status,
		Output:           c.Output,
	}).Error
	if err != nil {
		klog.Error("InsertDB error: ", err.Error())
	}
}

func (c *Mission) UpdateDB() {
	err := store.DB.Model(&model.Mission{}).Where("unique_name = ?", c.Config.UniqueName).UpdateColumns(map[string]interface{}{
		"status":            c.Status,
		"output":            c.Output,
		"command":           c.Config.Command,
		"file_content":      c.Config.FileContent,
		"file_name":         c.Config.FileName,
		"working_directory": c.Config.WorkingDirectory,
	}).Error
	if err != nil {
		klog.Error("InsertDB error: ", err.Error())
	}
}

func (c *Mission) Run() {
	if c.Status == StatusOK || c.Status == StatusError || c.Status == StatusWorking {
		klog.Info("Mission status is not waiting, skip with current status ", c.Status)
		return
	}

	defer func() {
		c.UpdateDB()
		c.ReportMissionStatus()
		klog.Info("Mission finished: ", c.Config.UniqueName, " result: ", c.Status)
	}()

	c.Status = StatusWorking
	c.ReportMissionStatus()
	klog.Info("Mission start: ", c.Config.UniqueName, " status: ", c.Status, " output: ", c.Output)

	// clean working directory in  windows
	dir := c.Config.WorkingDirectory
	os.RemoveAll(dir)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		klog.Errorf("Failed to make workdir %s: %v", dir, err)
		return
	}

	file, err := os.Create(filepath.Join(dir, c.Config.FileName))
	if err != nil {
		klog.Error("Create file error: ", err)
		c.Status = StatusError
		c.Output = err.Error()
		return
	}
	_, err = file.WriteString(encode.DecodeBase64(c.Config.FileContent))
	file.Close()

	if err != nil {
		klog.Error("Write file error: ", err)
		c.Status = StatusError
		c.Output = err.Error()
		return
	}

	err = c.exec.Exec()
	if err != nil {
		klog.Error("Exec error: ", err)
		c.Status = StatusError
		c.Output = fmt.Sprintf("【msg】%s\n【err】%s\n", err.Error(), string(c.exec.StdErr))
		return
	}

	c.Status = StatusOK
	c.Output = string(c.exec.StdOut)
}

func (c *Mission) ReportDeviceStatus() {
	var payload []byte
	var err error
	if payload, err = mqtt.CreateMessageState("OK"); err != nil {
		klog.Errorf("Create message state failed: %v", err)
		return
	}
	if err = mqtt.GetClient().Publish(fmt.Sprintf(mqtt.TopicPubDeviceStateUpdateRequest, c.Config.UniqueName), payload); err != nil {
		klog.Errorf("Publish failed: %v", err)
		return
	}
}

func (c *Mission) ReportMissionStatus() {
	var payload []byte
	var err error
	if payload, err = mqtt.CreateMessageTwinUpdate(map[string]string{
		"status":            c.Status,
		"output":            c.Output,
		"exec-command":      c.Config.Command,
		"exec-file-name":    c.Config.FileName,
		"exec-file-content": c.Config.FileContent,
	}); err != nil {
		klog.Errorf("Create message state failed: %v", err)
		return
	}
	if err = mqtt.GetClient().Publish(fmt.Sprintf(mqtt.TopicPubTwinUpdateRequest, c.Config.UniqueName), payload); err != nil {
		klog.Errorf("Publish failed: %v", err)
		return
	}
}
