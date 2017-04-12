package libtr

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/kyokomi/emoji"
)

func isTerraformTarget(name string) bool {
	info, err := ioutil.ReadDir(name)
	if err != nil {
		panic(err)
	}
	for _, item := range info {
		if !item.IsDir() {
			if strings.HasSuffix(item.Name(), ".tf") {
				return true
			}
		}
	}
	return false
}

func findTerraformTargets() []string {
	info, err := ioutil.ReadDir(".")
	if err != nil {
		panic(err)
	}
	res := []string{}
	for _, item := range info {
		if item.IsDir() {
			if isTerraformTarget(item.Name()) {
				res = append(res, item.Name())
			}
		}
	}
	return res
}

func (vault *Vault) InitRemoteState(target string) {
	emoji.Println(":pray: init remote state")

	bin, err := exec.LookPath("terraform")
	if err != nil {
		panic(err)
	}

	if err := RimRaf(".terraform"); err != nil {
		panic(err)
	}

	cmd := exec.Command(bin, "remote", "config",
		"-backend", "s3",
		"-backend-config", fmt.Sprintf("bucket=%s", vault.stateBucket()),
		"-backend-config", fmt.Sprintf("key=%s-%s.tfstate", vault.EnvName(), target),
		"-backend-config", fmt.Sprintf("region=%s", vault.awsRegion()),
		"-backend-config", fmt.Sprintf("access_key=%s", vault.awsKey()),
		"-backend-config", fmt.Sprintf("secret_key=%s", vault.awsSecret()))

	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	if err := cmd.Wait(); err != nil {
		panic(err)
	}
	emoji.Println(":ok_hand: remote state ready")
}
