package x

import "strings"

func SpaceSplit(s string) []string {
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
