package foundation

import "testing"

func TestSanitizeName(t *testing.T) {
	var cases = []struct {
		name     string
		in       string
		expected string
	}{
		{
			name:     "valid name should not change",
			in:       "valid_name",
			expected: "valid_name",
		},
		{
			name:     "name with space should be changed with underscore",
			in:       "name with a space",
			expected: "name_with_a_space",
		},
		{
			name:     "name with dot should be changed with underscore",
			in:       "name.with.a.dot",
			expected: "name_with_a_dot",
		},
		{
			name:     "name with mix should be changed with underscore",
			in:       "name with-a--mix@set",
			expected: "name_with_a__mix_set",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			newName := sanitizeName(tc.in)
			if newName != tc.expected {
				t.Errorf("expected %s but got %s", tc.expected, newName)
			}
		})
	}
}
