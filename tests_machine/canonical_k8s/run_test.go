package qa

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/gruntwork-io/terratest/modules/terraform"
	utils "github.com/juju/terraform-provider-juju-qa"
)

const kubeConfigPath = "./kubeconfig"

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
	utils.JujuWaitFor(t, "k8s")
	utils.JujuWaitFor(t, "k8s-worker")

	// *** deploy on k8s cluster
	// arrange
	addCloud(t, info.Name)

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
	utils.JujuWaitFor(t, "source")
	utils.JujuWaitFor(t, "sink")
}

func addCloud(t *testing.T, controllerName string) {
	config := getKubeconfig(t)
	config = unnestKubeconfig(t, config)

	if err := os.WriteFile(kubeConfigPath, config, 0600); err != nil {
		t.Fatalf("failed to write kubeconfig file: %v", err)
	}

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
}

func getKubeconfig(t *testing.T) []byte {
	for range 5 {
		cmd := exec.Command(
			"juju", "run",
			"k8s/0", "get-kubeconfig",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to get kubeconfig: %s", stderr.String())
		}

		if stdout.Len() > 0 {
			return stdout.Bytes()
		}
		time.Sleep(10 * time.Second)
	}

	return nil
}

func unnestKubeconfig(t *testing.T, raw []byte) []byte {
	var wrapper map[string]any
	if err := yaml.Unmarshal(raw, &wrapper); err != nil {
		t.Fatalf("failed to parse kubeconfig yaml: %v", err)
	}

	inner, ok := wrapper["kubeconfig"]
	if !ok {
		t.Fatal("kubeconfig key not found in yaml")
	}

	config, err := yaml.Marshal(inner)
	if err != nil {
		t.Fatalf("failed to serialise kubeconfig: %v", err)
	}

	return config
}
