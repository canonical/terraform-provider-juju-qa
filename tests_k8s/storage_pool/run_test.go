package qa

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	utils "github.com/juju/terraform-provider-juju-qa"
)

func TestQA_StoragePool(t *testing.T) {
	info := utils.GetMainControllerInfo(t)
	if info.CloudType != "k8s" {
		t.Skip("Skipping test on non-k8s cloud")
	}

	// arrange
	tfOpts := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",
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
	utils.JujuWaitForApplication(t, "db")
}
