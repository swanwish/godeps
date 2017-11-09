package bash

import (
	"os/exec"

	"github.com/swanwish/go-common/logs"
)

func ExecuteCmd(command string) (string, error) {
	logs.Debugf("Execute command %s", command)
	cmd := exec.Command("bash", "-c", command)

	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		logs.Errorf("Failed to combiled output, the error is %v", err)
		return "", err
	} else {
		return string(cmdOutput), nil
	}
}
