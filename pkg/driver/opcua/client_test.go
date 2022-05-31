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

package opcua

import (
	"testing"
)

func TestReadWithoutAuth(t *testing.T) {
	/*
		c := &OPCUAConfig{URL: "opc.tcp://localhost:4840"}

		client, err := NewClient(*c)
		assert.Nil(t, err)
		results, err := client.Get("ns=2;i=3")
		if err != nil {
			fmt.Println("Read error: ", err)
			return
		}
		fmt.Println("result: ", results)
	*/
}

func TestReadWithUserPass(t *testing.T) {
	/*
		c := &OPCUAConfig{URL: "opc.tcp://localhost:4840",
			User:           "testuser",
			Passwordfile:   "/home/wei/ca/pass",
			SecurityPolicy: "None",
			SecurityMode:   "None",
		}
		client, err := NewClient(*c)
		assert.Nil(t, err)
		results, err := client.Get("ns=2;i=3")
		if err != nil {
			fmt.Println("Read error: ", err)
			return
		}
		fmt.Println("result: ", results)
	*/
}

func TestReadWithCert(t *testing.T) {
	/*
		c := &OPCUAConfig{URL: "opc.tcp://localhost:4840",
			SecurityPolicy: "Basic256Sha256",
			SecurityMode:   "Sign",
			Certfile:       "/home/wei/ca/clientcert.pem",
			Keyfile:        "/home/wei/ca/clientkey.pem",
			RemoteCertfile: "/home/wei/ca/servercert.pem",
		}
		client, err := NewClient(*c)
		assert.Nil(t, err)
		results, err := client.Get("ns=2;i=3")
		if err != nil {
			fmt.Println("Read error: ", err)
			return
		}
		fmt.Println("result: ", results)
	*/
}

func TestWrite(t *testing.T) {
	/*
		c := &OPCUAConfig{URL: "opc.tcp://localhost:4840",
			User:           "",
			Passwordfile:   "",
			SecurityPolicy: "None",
			SecurityMode:   "None",
			Certfile:       "",
			Keyfile:        ""}

		client, err := NewClient(*c)
		assert.Nil(t, err)

		results, err := client.Set("ns=2;i=3", "true")
		if err != nil {
			fmt.Println("Write error: ", err)
			return
		}
		fmt.Println("result: ", results)
	*/
}
