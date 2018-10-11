package ssh

import (
	"errors"
)

func ParseClient(pattern string) ([]string, error) {
	clis := make([]string, 0)
	if len(pattern) == 0 {
		return clis, errors.New("hosts cannot have empty value")
	}
	//hostList := strings.Split(pattern, ",")
	return clis, nil
}

func ParseModule() {}

func ParseArgs() {}
