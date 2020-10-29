package main

import (
	"bufio"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWaitForInfo(t *testing.T) {
	cases := map[string]struct {
		write         []string
		plans         []byte
		expectUpgrade *UpgradeInfo
		expectErr     bool
	}{
		"no match": {
			write: []string{"some", "random\ninfo\n"},
			plans: []byte(`{"upgrades":[{"height":200,"version":"1.4.6"}]}`),
		},
		"match height to first upgrade plan": {
			write: []string{"first line\n", `Committed state      module=state height=200 txs=111 appHash=4694FE87727435B239A33866D3EFCA759A5DE905EB1EDF1D257630D6C1868459`, "\nnext line\n"},
			plans: []byte(`{"upgrades":[{"height":200,"version":"1.4.6"}]}`),
			expectUpgrade: &UpgradeInfo{
				Name:   "1.4.6",
				Height: 200,
				Time: "",
				Info: "",
			},
		},
		"match height to second upgrade plan": {
			write: []string{"first line\n", `Committed state                              module=state height=300 txs=0 appHash=4694FE87727435B239A33866D3EFCA759A5DE905EB1EDF1D257630D6C1868459`, "\nnext line\n"},
			plans: []byte(`{"upgrades":[{"height":200,"version":"1.4.6"},{"height":300,"version":"1.4.7"}]}`),
			expectUpgrade: &UpgradeInfo{
				Name:   "1.4.7",
				Height: 300,
				Time: "",
				Info: "",
			},
		},
		"chunks": {
			write: []string{"first l", "ine\nERROR 2020-02-03T11:22:33Z: Committed state   ", `module=state `, "height=200 txs=999999 appHash=4694FE87727435B239A33866D3EFCA759A5DE905EB1EDF1D257630D6C1868459", "  \n LOG: next line"},
			plans: []byte(`{"upgrades":[{"height":200,"version":"1.4.6"}]}`),
			expectUpgrade: &UpgradeInfo{
				Name:   "1.4.6",
				Height: 200,
				Time: "",
				Info: "",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r, w := io.Pipe()
			scan := bufio.NewScanner(r)

			// write all info in separate routine
			go func() {
				for _, line := range tc.write {
					n, err := w.Write([]byte(line))
					assert.NoError(t, err)
					assert.Equal(t, len(line), n)
				}
				w.Close()
			}()

			upgradePlans, _ := LoadPlanFromJsonBytes(tc.plans)

			// now scan the info
			info, err := WaitForUpdate(scan, upgradePlans)
			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectUpgrade, info)
		})
	}
}
