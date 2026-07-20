//go:build windows

package runtime

import "github.com/shirou/gopsutil/disk"

func (r *runtime) Disk() (free, total float64) {
	usage, err := disk.Usage(".")
	if err != nil {
		return 0, 0
	}

	free = float64(usage.Free) / (1024 * 1024 * 1024)
	total = float64(usage.Total) / (1024 * 1024 * 1024)
	return
}
