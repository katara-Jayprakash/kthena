/*
Copyright The Volcano Authors.

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
	"net"
	"os/exec"
	"time"
)

// SetupPortForward sets up a kubectl port-forward and waits for it to be ready.
// It returns a cleanup function that should be called to kill the port-forward process.
func SetupPortForward(namespace, service, localPort, remotePort string) (func(), error) {
	// Setup port-forward
	pfCmd := exec.Command("kubectl", "port-forward", "-n", namespace, "--address", "127.0.0.1", fmt.Sprintf("svc/%s", service), fmt.Sprintf("%s:%s", localPort, remotePort))
	if err := pfCmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start port-forward: %v", err)
	}

	cleanup := func() {
		if pfCmd.Process != nil {
			_ = pfCmd.Process.Kill()
		}
	}

	// Start a goroutine to wait for the command to finish to avoid zombie processes.
	// The error is expected if the process is killed during cleanup.
	go func() {
		if err := pfCmd.Wait(); err != nil {
			// Ignore error as it's expected when the process is killed during cleanup
		}
	}()

	// Wait for port-forward to be ready
	address := fmt.Sprintf("127.0.0.1:%s", localPort)
	timeout := 15 * time.Second
	interval := 1 * time.Second
	start := time.Now()

	for {
		conn, err := net.DialTimeout("tcp", address, 1*time.Second)
		if err == nil {
			conn.Close()
			return cleanup, nil
		}
		if time.Since(start) > timeout {
			cleanup()
			return nil, fmt.Errorf("timeout waiting for port-forward on %s to become available", address)
		}
		time.Sleep(interval)
	}
}
