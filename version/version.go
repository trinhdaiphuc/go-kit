package version

import (
	versionsdk "github.com/hashicorp/go-version"
)

func CompareVersion(version string, pivotVersion string) bool {
	ver, err := versionsdk.NewVersion(version)
	if err != nil {
		return false
	}

	pivotVer, err := versionsdk.NewVersion(pivotVersion)
	if err != nil {
		return false
	}
	return ver.GreaterThanOrEqual(pivotVer)
}
