package middleware

import "testing"

func TestResolveResource(t *testing.T) {
	cases := []struct {
		path     string
		expected string
	}{
		{"/api/v1/users", "users"},
		{"/api/v1/users/abc-id-1", "users"},
		{"/api/v1/customers", "customers"},
		{"/api/v1/retail/contracts", "retail_contracts"},
		{"/api/v1/retail/contracts/abc-id-2", "retail_contracts"},
		{"/api/v1/retail/packages", "retail_packages"},
		{"/api/v1/scheduler/jobs/abc-id-3/trigger", "scheduler"},
		{"/api/v1/analytics/alerts", "analytics"},
		// 长前缀优先：retail_contracts 应比 retail_packages 不冲突，
		// 但要确保 /api/v1/retail 不会错误匹配为 retail_packages
		{"/api/v1/retail", ""}, // 没有任何 prefix 完全等于 /api/v1/retail，应返回 nil
	}

	for _, c := range cases {
		got := resolveResource(c.path)
		if c.expected == "" {
			if got != nil {
				t.Errorf("path %q 应返回 nil，得到 %q", c.path, *got)
			}
			continue
		}
		if got == nil {
			t.Errorf("path %q 应返回 %q，得到 nil", c.path, c.expected)
			continue
		}
		if *got != c.expected {
			t.Errorf("path %q 应返回 %q，得到 %q", c.path, c.expected, *got)
		}
	}
}

func TestUUIDExtractor(t *testing.T) {
	cases := map[string]string{
		"/api/v1/customers/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee":         "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		"/api/v1/customers/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee/edit":    "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		"/api/v1/customers":                                              "",
		"/api/v1/customers/not-a-uuid":                                   "",
	}
	for path, want := range cases {
		got := uuidRe.FindString(path)
		if got != want {
			t.Errorf("path %q 期望 %q，得到 %q", path, want, got)
		}
	}
}
