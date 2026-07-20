//go:build !windows

package runtime

import "syscall"

func (r *runtime) Disk() (free, total float64) {
	var diskStat syscall.Statfs_t
	err := syscall.Statfs(".", &diskStat)
	if err != nil {
		return 0, 0
	}

	free = float64(diskStat.Bavail*uint64(diskStat.Bsize)) / (1024 * 1024 * 1024)
	total = float64(diskStat.Blocks*uint64(diskStat.Bsize)) / (1024 * 1024 * 1024)
	return
}
