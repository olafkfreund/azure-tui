package tfbicep

import (
	"os/exec"
)

// Terraform
func TerraformInit(dir string) error {
	cmd := exec.Command("terraform", "init")
	cmd.Dir = dir
	return cmd.Run()
}

func TerraformPlan(dir string) error {
	cmd := exec.Command("terraform", "plan")
	cmd.Dir = dir
	return cmd.Run()
}

func TerraformApply(dir string) error {
	cmd := exec.Command("terraform", "apply", "-auto-approve")
	cmd.Dir = dir
	return cmd.Run()
}

func TerraformDestroy(dir string) error {
	cmd := exec.Command("terraform", "destroy", "-auto-approve")
	cmd.Dir = dir
	return cmd.Run()
}

// Bicep
func BicepBuild(file string) error {
	return exec.Command("bicep", "build", file).Run()
}

func BicepDeploy(file, group string) error {
	return exec.Command("az", "deployment", "group", "create", "--resource-group", group, "--template-file", file).Run()
}
