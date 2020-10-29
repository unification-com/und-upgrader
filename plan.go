package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
)

type UpgradePlan struct {
	Height  int   `json:"height"`
    Version string `json:"version"`
	RegEx   *regexp.Regexp
}

type UpgradePlans struct {
	UpgradePlans []UpgradePlan `json:"upgrades"`
}

func LoadPlans(cfg *Config) (UpgradePlans, error) {

	jsonFile, err := os.Open(cfg.Plan())

	if err != nil {
		fmt.Println(err)
		return UpgradePlans{}, err
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		fmt.Println(err)
		return UpgradePlans{}, err
	}

	return LoadPlanFromJsonBytes(byteValue)
}

func LoadPlanFromJsonBytes(plansJson []byte) (UpgradePlans, error) {
	var upgradePlans UpgradePlans

	json.Unmarshal(plansJson, &upgradePlans)

	// generate the regEx for the scanner
	for i := 0; i < len(upgradePlans.UpgradePlans); i++ {
		upgradePlans.UpgradePlans[i].RegEx = regexp.MustCompile(fmt.Sprintf("Committed state\\s+module=state height=%d txs=\\d+ appHash=[A-Z0-9]{64}\\s*$", upgradePlans.UpgradePlans[i].Height))
	}

	return upgradePlans, nil
}