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

// This application needs OPCUA server and device.
// Please edit by demand for testing.

package onvif

/*
func TestReadWithoutAuth(t *testing.T) {
	c := OnvifConfig{Name: "camera0",
		URL:          "192.168.168.64:80",
		User:         "admin",
		Passwordfile: "/home/wei/ca/pass"}

	client, err := NewClient(c)
	assert.Nil(t, err)
	fmt.Printf("streamuri: %s", client.Config.StreamURI)
}

func TestGetResource(t *testing.T) {
	c := OnvifConfig{Name: "camera1",
		URL:          "192.168.168.64:80",
		User:         "admin",
		Passwordfile: "/home/wei/ca/pass"}

	_, err := NewClient(c)
	assert.Nil(t, err)
	resources := GetOnvifResources()
	for _, r := range resources.Resources {
		fmt.Println(r)
	}
}

func TestSaveFrame(t *testing.T) {
	c := OnvifConfig{Name: "camera1",
		URL:          "192.168.168.64:80",
		User:         "admin",
		Passwordfile: "/home/wei/ca/pass"}

	client, err := NewClient(c)
	assert.Nil(t, err)
	IfSaveFrame = true
	streamURI := client.GetStream()
	err = SaveFrame(streamURI, "/home/wei/output", "jpg", 50, 1000000000)
	assert.Nil(t, err)
}

func TestSaveVideo(t *testing.T) {
	c := OnvifConfig{Name: "camera1",
		URL:          "192.168.168.64:80",
		User:         "admin",
		Passwordfile: "/home/wei/ca/pass"}

	client, err := NewClient(c)
	assert.Nil(t, err)
	IfSaveVideo = true
	streamURI := client.GetStream()
	err = SaveVideo(streamURI, "/home/wei/output", "mp4", 500)
	assert.Nil(t, err)
}
*/
