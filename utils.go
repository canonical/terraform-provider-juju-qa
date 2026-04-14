package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/shell"
	"gopkg.in/yaml.v3"
)

const defaultMainControllerName = "tfqa"
const defaultOfferingControllerName = "tfqa-offering"

type ControllerInfo struct {
	Name      string
	Addresses string
	Username  string
	Password  string
	CACert    string
	CloudType string
}

type whoamiOutput struct {
	Controller string `json:"controller"`
}

type ctrlDetails struct {
	APIEndpoints []string `json:"api-endpoints"`
	CACert       string   `json:"ca-cert"`
	Cloud        string   `json:"cloud"`
}

type ctrlInfo struct {
	Details ctrlDetails `json:"details"`
}

type cloudInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ctrlCredentials struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type accountsFile struct {
	Controllers map[string]ctrlCredentials `yaml:"controllers"`
}

// GetCurrentControllerName returns the name of the CLI's current controller
func GetCurrentControllerName(t *testing.T) string {
	out := shell.RunCommandAndGetOutput(t, shell.Command{
		Command: "juju",
		Args:    []string{"whoami", "--format", "json"},
	})

	var whoami whoamiOutput
	err := json.Unmarshal([]byte(out), &whoami)
	if err != nil {
		t.Fatalf("failed to unmarshal whoami output: %s", err)
	}
	return whoami.Controller
}

// GetControllerInfo returns connection info for the given controller name
func GetControllerInfo(t *testing.T, controllerName string) ControllerInfo {
	out := shell.RunCommandAndGetOutput(t, shell.Command{
		Command: "juju",
		Args:    []string{"show-controller", controllerName, "--format", "json"},
	})

	var showCtrl map[string]ctrlInfo
	err := json.Unmarshal([]byte(out), &showCtrl)
	if err != nil {
		t.Fatalf("failed to unmarshal show-controller output: %s", err)
	}

	out = shell.RunCommandAndGetOutput(t, shell.Command{
		Command: "juju",
		Args:    []string{"show-cloud", showCtrl[controllerName].Details.Cloud, "--format", "json"},
	})

	var showCloud []cloudInfo
	err = json.Unmarshal([]byte(out), &showCloud)
	if err != nil {
		t.Fatalf("failed to unmarshal show-cloud output: %s", err)
	}

	file, err := os.Open(os.ExpandEnv("$HOME/.local/share/juju/accounts.yaml"))
	if err != nil {
		t.Fatalf("failed to open accounts file: %s", err)
	}
	defer file.Close()

	var accounts accountsFile
	err = yaml.NewDecoder(file).Decode(&accounts)
	if err != nil {
		t.Fatalf("failed to unmarshal accounts file: %s", err)
	}

	addresses := strings.Join(showCtrl[controllerName].Details.APIEndpoints, ",")
	var cloudType string
	for _, c := range showCloud {
		if c.Name == showCtrl[controllerName].Details.Cloud {
			cloudType = c.Type
			break
		}
	}

	return ControllerInfo{
		Name:      controllerName,
		Addresses: addresses,
		Username:  accounts.Controllers[controllerName].User,
		Password:  accounts.Controllers[controllerName].Password,
		CACert:    showCtrl[controllerName].Details.CACert,
		CloudType: cloudType,
	}
}

// GetMainControllerInfo returns connection info for the main controller
func GetMainControllerInfo(t *testing.T) ControllerInfo {
	controllerName := defaultMainControllerName
	if envCtrlname := os.Getenv("TF_JUJU_QA_CTRL"); envCtrlname != "" {
		controllerName = envCtrlname
		fmt.Printf("%s\n", envCtrlname)
	}
	return GetControllerInfo(t, controllerName)
}

// GetOfferingControllerInfo returns connection info for the offering controller
func GetOfferingControllerInfo(t *testing.T) ControllerInfo {
	controllerName := defaultOfferingControllerName
	if envCtrlname := os.Getenv("TF_JUJU_QA_OFFERING_CTRL"); envCtrlname != "" {
		controllerName = envCtrlname
		fmt.Printf("%s\n", envCtrlname)
	}
	return GetControllerInfo(t, controllerName)
}

// Env returns controller info as a map of env vars
func (i ControllerInfo) Env() map[string]string {
	return map[string]string{
		"JUJU_CONTROLLER_ADDRESSES": i.Addresses,
		"JUJU_USERNAME":             i.Username,
		"JUJU_PASSWORD":             i.Password,
		"JUJU_CA_CERT":              i.CACert,
	}
}

// OfferingVars returns controller info as a map of terraform vars
func (i ControllerInfo) OfferingVars() map[string]any {
	return map[string]any{
		"offering_controller_name":      i.Name,
		"offering_controller_addresses": i.Addresses,
		"offering_controller_username":  i.Username,
		"offering_controller_password":  i.Password,
		"offering_controller_ca_cert":   i.CACert,
	}
}

// JujuSwitch wraps `juju switch`
func JujuSwitch(t *testing.T, to string) {
	cmd := exec.Command(
		"juju", "switch", to,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed juju switch: %s", out)
	}
}

// JujuStatus wraps `juju status --format yaml`
func JujuStatus(t *testing.T) {
	cmd := exec.Command(
		"juju", "status", "--format", "yaml",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed juju status: %s", out)
	} else {
		t.Logf("juju status output: \n%s", out)
	}
}

// JujuWaitFor wraps `juju wait-for application <app> --timeout 60m`
// and will log `juju status` on failure
func JujuWaitFor(t *testing.T, application string) {
	cmd := exec.Command(
		"juju", "wait-for",
		"application", "--timeout", "60m",
		application,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		JujuStatus(t)

		t.Logf("failed juju wait-for: %s", out)
		t.Fail()
	}
}
