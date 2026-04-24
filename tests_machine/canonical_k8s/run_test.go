package qa

import (
	"os/exec"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	utils "github.com/juju/terraform-provider-juju-qa"
)

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
	cmd := exec.Command(
		"bash", "-e", "-x", "-c", "./setup-cloud.sh",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to set up k8s cloud: %s", out)
	}

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
