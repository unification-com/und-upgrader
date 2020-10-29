package main

import (
	"bufio"
)

// UpgradeInfo is the details from the regexp
type UpgradeInfo struct {
	Name string
	// Only 1 of Height or Time is non-zero value
	Height int
	Time   string
	Info   string
}

// WaitForUpdate will listen to the scanner until a line matches upgradeRegexp.
// It returns (info, nil) on a matching line
// It returns (nil, err) if the input stream errored
// It returns (nil, nil) if the input closed without ever matching the regexp
func WaitForUpdate(scanner *bufio.Scanner, plans UpgradePlans) (*UpgradeInfo, error) {
	for scanner.Scan() {
		line := scanner.Text()
		for i := 0; i < len(plans.UpgradePlans); i++ {
			plan := plans.UpgradePlans[i]
			if plan.RegEx.MatchString(line) {
				info := UpgradeInfo{
					Name: plan.Version,// subs[1],
					Info: "", // subs[7],
					Height: plan.Height,
				}
				return &info, nil
			}
		}
	}
	return nil, scanner.Err()
}
