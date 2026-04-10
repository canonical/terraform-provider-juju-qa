package qa

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	utils "github.com/juju/terraform-provider-juju-qa"
)

func TestQA_CanonicalK8S(t *testing.T) {
	// arrange
	info := utils.GetMainControllerInfo(t)

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
	utils.JujuWaitFor(t, "k8s")
}
