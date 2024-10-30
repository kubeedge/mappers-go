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
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

type Command struct {
	Cmd      *exec.Cmd
	StdOut   []byte
	StdErr   []byte
	ExitCode int
}

// Exec run command and exit formatted error, callers can print err directly
// Any running error or non-zero exitcode is consider as error
func (cmd *Command) Exec() error {
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Cmd.Stdout = &stdoutBuf
	cmd.Cmd.Stderr = &stderrBuf

	errString := fmt.Sprintf("failed to exec '%s'", cmd.GetCommand())

	err := cmd.Cmd.Start()
	if err != nil {
		errString = fmt.Sprintf("%s, err: %v", errString, err)
		return errors.New(errString)
	}

	err = cmd.Cmd.Wait()
	if err != nil {
		cmd.StdErr = stderrBuf.Bytes()

		if exit, ok := err.(*exec.ExitError); ok {
			cmd.ExitCode = exit.Sys().(syscall.WaitStatus).ExitStatus()
			errString = fmt.Sprintf("%s, err: %s", errString, stderrBuf.Bytes())
		} else {
			cmd.ExitCode = 1
		}

		errString = fmt.Sprintf("%s, err: %v", errString, err)

		return errors.New(errString)
	}

	cmd.StdOut, cmd.StdErr = stdoutBuf.Bytes(), stderrBuf.Bytes()
	return nil
}

func (cmd Command) GetCommand() string {
	return strings.Join(cmd.Cmd.Args, " ")
}

func (cmd Command) GetStdOut() string {
	if len(cmd.StdOut) != 0 {
		return strings.TrimSuffix(string(cmd.StdOut), "\n")
	}
	return ""
}

func (cmd Command) GetStdErr() string {
	if len(cmd.StdErr) != 0 {
		return strings.TrimSuffix(string(cmd.StdErr), "\n")
	}
	return ""
}
