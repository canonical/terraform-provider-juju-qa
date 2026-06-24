package qa

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/gruntwork-io/terratest/modules/terraform"
	utils "github.com/juju/terraform-provider-juju-qa"
)

type jujuActionResult struct {
	Results struct {
		Kubeconfig string `json:"kubeconfig"`
	} `json:"results"`
}

func TestQA_CanonicalK8S(t *testing.T) {
	// *** provision k8s cluster
	// arrange
	info := utils.GetMainControllerInfo(t)

	tfOpts := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "./provision",
		EnvVars:      info.Env(),
		Reconfigure:  true,
		NoColor:      true,
	})

	// act
	defer terraform.Destroy(t, tfOpts)
	terraform.InitAndApply(t, tfOpts)

	// assert
	modelName := terraform.Output(t, tfOpts, "model_name")

	utils.JujuSwitch(t, info.Name+":"+modelName)
	utils.JujuWaitForApplication(t, "k8s")

	cmd := exec.Command(
		"juju", "status",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed get juju status: %s", out)
	}

	// *** deploy on k8s cluster
	// arrange
	removeCloud := addCloud(t, info.Name)
	defer removeCloud()

	tfOpts = terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "./deploy",
		EnvVars:      info.Env(),
		Reconfigure:  true,
		NoColor:      true,
		Vars: map[string]any{
			"cloud":      "tfqa-k8s",
			"credential": "tfqa-k8s",
		},
	})

	// act
	defer terraform.Destroy(t, tfOpts)
	terraform.InitAndApply(t, tfOpts)

	// assert
	modelName = terraform.Output(t, tfOpts, "model_name")

	utils.JujuSwitch(t, info.Name+":"+modelName)
	utils.JujuWaitForApplication(t, "source")
	utils.JujuWaitForApplication(t, "sink")
}

func addCloud(t *testing.T, controllerName string) func() {
	config := getKubeconfig(t)
	inner := extractKubeconfig(config)
	if len(inner) == 0 {
		t.Fatalf("kubeconfig key not found in output: %s", string(config))
	}

	f, err := os.CreateTemp(".", "kubeconfig-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp kubeconfig file: %v", err)
	}
	kubeConfigPath, err := filepath.Abs(f.Name())
	if err != nil {
		f.Close()
		t.Fatalf("failed to resolve kubeconfig path: %v", err)
	}
	if _, err := f.Write(inner); err != nil {
		f.Close()
		t.Fatalf("failed to write kubeconfig file: %v", err)
	}
	f.Close()
	defer os.Remove(kubeConfigPath)

	cmd := exec.Command(
		"juju", "add-k8s",
		"tfqa-k8s",
		"--client",
		"--controller="+controllerName,
		"--cluster-name=k8s",
	)
	cmd.Env = append(os.Environ(), "KUBECONFIG="+kubeConfigPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to add k8s cloud: %s", out)
	}

	return func() {
		cmd := exec.Command(
			"juju", "remove-cloud",
			"tfqa-k8s",
			"--client",
			"--controller="+controllerName,
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("failed to remove k8s cloud: %s", out)
		}
	}
}

func getKubeconfig(t *testing.T) []byte {
	cmd := exec.Command(
		"juju", "run",
		"--wait=5m",
		"--format=json",
		"k8s/0", "get-kubeconfig",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to get kubeconfig: %s", stderr.String())
	}

	if config := extractKubeconfigFromJSON(stdout.Bytes()); len(config) > 0 {
		return config
	}

	t.Fatalf("kubeconfig not found in juju run output: %s", stdout.String())
	return nil
}

func extractKubeconfig(raw []byte) []byte {
	if config := extractKubeconfigFromJSON(raw); len(config) > 0 {
		return config
	}

	var wrapper map[string]string
	if err := yaml.Unmarshal(raw, &wrapper); err == nil {
		if inner, ok := wrapper["kubeconfig"]; ok {
			return []byte(inner)
		}
	}

	return nil
}

func extractKubeconfigFromJSON(raw []byte) []byte {
	var direct jujuActionResult
	if err := json.Unmarshal(raw, &direct); err == nil && direct.Results.Kubeconfig != "" {
		return []byte(direct.Results.Kubeconfig)
	}

	var wrapped map[string]jujuActionResult
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil
	}

	for _, result := range wrapped {
		if result.Results.Kubeconfig != "" {
			return []byte(result.Results.Kubeconfig)
		}
	}

	return nil
}
