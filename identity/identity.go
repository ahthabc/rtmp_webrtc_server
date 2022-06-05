package identity

import (
	"fmt"
	"strings"

	"github.com/toolkits/pkg/file"
	"github.com/toolkits/pkg/sys"
)

type Identity struct {
	IP    shell `yaml:"ip"`
	Ident shell `yaml:"ident"`
	SN    shell `yaml:"sn"`
}

type shell struct {
	Specify string `yaml:"specify"`
	Shell   string `yaml:"shell"`
}

var (
	config Identity
	SN     string
)

func Parse() error {
	yml := getIdentityYmlFile()
	if yml == "" {
		return fmt.Errorf("etc/identity[.local].yml not found")
	}

	var i Identity
	if err := file.ReadYaml(yml, &i); err != nil {
		return err
	}

	config = i
	return nil
}

func GetIP() (string, error) {
	return getByShell(config.IP)
}

func GetIdent() (string, error) {
	return getByShell(config.Ident)
}

// func GetSN() (string, error) {
// 	return getByShell(config.SN)
// }
func GetSN() (string, error) {
	var err error
	if SN == "" {
		SN, err = exec(config.SN.Shell)
		if err != nil {
			return "", fmt.Errorf("cannot get sn: %s", err)
		}

	}
	return SN, nil
}
func getByShell(s shell) (string, error) {
	if s.Specify != "" {
		return s.Specify, nil
	}

	out, err := sys.CmdOutTrim("sh", "-c", s.Shell)
	if err != nil {
		return "", err
	}

	if strings.Contains(out, " ") {
		return "", fmt.Errorf("output: %s invalid", out)
	}

	return out, nil
}
func exec(shell string) (string, error) {
	out, err := sys.CmdOutTrim("sh", "-c", shell)
	if err != nil {
		return "", fmt.Errorf("cannot exec `%s', error: %v", shell, err)
	}

	return out, nil
}
func getIdentityYmlFile() string {
	yml := "etc/identity.local.yml"
	if file.IsExist(yml) {
		return yml
	}

	yml = "./etc/identity.yml"
	if file.IsExist(yml) {
		return yml
	}

	return ""
}
