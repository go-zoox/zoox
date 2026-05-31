package cron

import (
	"testing"
)

func TestCronJobManagementAfterAdd(t *testing.T) {
	c := New()

	if c.HasJob("job-1") {
		t.Fatal("expected no jobs before add")
	}

	if err := c.AddJob("job-1", "@every 1h", func() error { return nil }); err != nil {
		t.Fatalf("add job: %v", err)
	}

	if !c.HasJob("job-1") {
		t.Fatal("expected job to exist after add")
	}

	if err := c.RemoveJob("job-1"); err != nil {
		t.Fatalf("remove job: %v", err)
	}

	if c.HasJob("job-1") {
		t.Fatal("expected job to be removed")
	}
}

func TestCronClearJobsAfterAdd(t *testing.T) {
	c := New()

	if err := c.AddJob("job-1", "@every 1h", func() error { return nil }); err != nil {
		t.Fatalf("add job-1: %v", err)
	}

	if err := c.AddJob("job-2", "@every 2h", func() error { return nil }); err != nil {
		t.Fatalf("add job-2: %v", err)
	}

	if err := c.ClearJobs(); err != nil {
		t.Fatalf("clear jobs: %v", err)
	}

	if c.HasJob("job-1") || c.HasJob("job-2") {
		t.Fatal("expected all jobs to be cleared")
	}
}

func TestCronNotStartedBeforeAdd(t *testing.T) {
	c := New()

	if err := c.RemoveJob("missing"); err == nil {
		t.Fatal("expected error when removing job before cron starts")
	}

	if err := c.ClearJobs(); err == nil {
		t.Fatal("expected error when clearing jobs before cron starts")
	}
}
