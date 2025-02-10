package pvc

import (
	"fmt"
)

// GeneratePVCName generates a PVC name based on the VM name.
func GeneratePVCName(vmName string) string {
	return fmt.Sprintf("%s-disk", vmName)
}
