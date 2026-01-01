package x

import (
	"strings"

	"github.com/google/uuid"
)

func RandomUUID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}

func SplitSpace(s string) []string {
	return RemoveEmpty(strings.Split(s, " "))
}

func RemoveEmpty(args []string) (ret []string) {
	for _, v := range args {
		v = strings.TrimSpace(v)
		if v != "" {
			ret = append(ret, v)
		}
	}
	return
}
