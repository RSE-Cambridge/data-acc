package fake

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlugin_PrintLustreInfo(t *testing.T) {
	volume := registry.Volume{Name: "1"}
	brickAllocations := []registry.BrickAllocation{
		{Hostname: "dac1", Device: "nvme0n1", AllocatedIndex: 0},
		{Hostname: "dac1", Device: "nvme2n1", AllocatedIndex: 1},
		{Hostname: "dac1", Device: "nvme3n1", AllocatedIndex: 2},
		{Hostname: "dac2", Device: "nvme2n1", AllocatedIndex: 3},
		{Hostname: "dac2", Device: "nvme3n1", AllocatedIndex: 4},
	}
	result := printLustreInfo(volume, brickAllocations)
	expected := `all:
  children:
    fs1:
      hosts:
        dac1:
          mdts: [nvme0n1]
          osts: {nvme2n1: 1, nvme3n1: 2}
        dac2:
          osts: {nvme2n1: 3, nvme3n1: 4}
      vars: {mgsnode: dac1}
`
	assert.Equal(t, expected, result)
}
