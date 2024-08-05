package zoox

import "testing"

func TestGroupMatchPath(t *testing.T) {
	testcases := []map[string]any{
		{
			"path":   "/",
			"prefix": "/",
			"expect": true,
		},
		{
			"path":   "/",
			"prefix": "/api",
			"expect": false,
		},
		{
			"path":   "/api",
			"prefix": "/",
			"expect": true,
		},
		{
			"path":   "/v1/containers/d0ac6213f33620362e59cc1b855658f9792377335087c2f3ba1d43639466dd8a/terminal",
			"prefix": "/v1/containers/:id",
			"expect": true,
		},
	}

	for _, testcase := range testcases {
		group := &RouterGroup{
			prefix: testcase["prefix"].(string),
		}

		if got := group.matchPath(testcase["path"].(string)); got != testcase["expect"] {
			t.Fatalf("expected %v, got %v (path: %s, group: %s)", testcase["expect"], got, testcase["path"], testcase["prefix"])
		}
	}
}
