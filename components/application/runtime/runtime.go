package runtime

import (
	rt "runtime"
	"syscall"

	"github.com/go-zoox/datetime"
	"github.com/go-zoox/logger"
)

// Runtime ...
type Runtime interface {
	CurrentTime() *datetime.DateTime
	OS() string
	Arch() string
	CPUCores() int
	Memory() (allocated, total uint64)
	Disk() (free, total float64)
	GoVersion() string
	GoRoot() string
	Print()
}

type runtime struct {
	logger *logger.Logger
}

// New ...
func New(logger *logger.Logger) Runtime {
	return &runtime{
		logger: logger,
	}
}

func (r *runtime) CurrentTime() *datetime.DateTime {
	return datetime.Now()
}

func (r *runtime) OS() string {
	return rt.GOOS
}

func (r *runtime) Arch() string {
	return rt.GOARCH
}

func (r *runtime) CPUCores() int {
	return rt.NumCPU()
}

func (r *runtime) Memory() (allocated, total uint64) {
	var memStats rt.MemStats
	rt.ReadMemStats(&memStats)
	allocated = memStats.Alloc / 1024 / 1024 // 转换为 MB
	total = memStats.Sys / 1024 / 1024       // 转换为 MB
	return
}

func (r *runtime) Disk() (free, total float64) {
	var diskStat syscall.Statfs_t
	err := syscall.Statfs(".", &diskStat)
	if err != nil {
		return 0, 0
	}

	free = float64(diskStat.Bavail*uint64(diskStat.Bsize)) / (1024 * 1024 * 1024)  // 转换为 GB
	total = float64(diskStat.Blocks*uint64(diskStat.Bsize)) / (1024 * 1024 * 1024) // 转换为 GB
	return
}

func (r *runtime) GoVersion() string {
	return rt.Version()
}

func (r *runtime) GoRoot() string {
	return rt.GOROOT()
}

func (r *runtime) Print() {
	r.logger.Infof("CurrentTime: %s", r.CurrentTime().Format("YYYY-MM-DD HH:mm:ss"))
	r.logger.Infof("OS: %s", r.OS())
	r.logger.Infof("Arch: %s", r.Arch())
	r.logger.Infof("CPU: %d", r.CPUCores())

	memAllocated, memTotal := r.Memory()
	r.logger.Infof("Memory: %d/%d MB (%.2f%%)", memAllocated, memTotal, float64(memAllocated)*100/float64(memTotal))

	diskFree, diskTotal := r.Disk()
	r.logger.Infof("Disk: %.2f/%.2f GB (%.2f%%)", diskTotal-diskFree, diskTotal, (diskFree-diskTotal)*100/diskTotal)
}
