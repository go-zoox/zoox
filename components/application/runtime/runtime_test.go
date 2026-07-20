package runtime

import (
	"testing"
)

func TestDisk(t *testing.T) {
	r := New(nil).(*runtime)

	free, total := r.Disk()

	if total <= 0 {
		t.Errorf("total disk space should be > 0, got %f", total)
	}
	if free < 0 {
		t.Errorf("free disk space should be >= 0, got %f", free)
	}
	if free > total {
		t.Errorf("free disk space (%f) should not exceed total (%f)", free, total)
	}
}

func TestInfo_DiskFields(t *testing.T) {
	r := New(nil).(*runtime)

	info := r.Info()

	if info.DiskTotal <= 0 {
		t.Errorf("DiskTotal should be > 0, got %f", info.DiskTotal)
	}
	if info.DiskUsed < 0 {
		t.Errorf("DiskUsed should be >= 0, got %f", info.DiskUsed)
	}
	if info.DiskUsed > info.DiskTotal {
		t.Errorf("DiskUsed (%f) should not exceed DiskTotal (%f)", info.DiskUsed, info.DiskTotal)
	}

	diskFree, diskTotal := r.Disk()
	expectedUsed := diskTotal - diskFree
	if info.DiskUsed != expectedUsed {
		t.Errorf("DiskUsed (%f) should equal total - free (%f)", info.DiskUsed, expectedUsed)
	}
}

func TestMemory(t *testing.T) {
	r := New(nil).(*runtime)

	used, total := r.Memory()

	if total <= 0 {
		t.Errorf("total memory should be > 0, got %d", total)
	}
	if used < 0 {
		t.Errorf("used memory should be >= 0, got %d", used)
	}
	if used > total {
		t.Errorf("used memory (%d) should not exceed total (%d)", used, total)
	}
}
