package auth

import (
	"testing"
)

func TestCreateHash(t *testing.T) {

	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "pa$$word",
			expected: "$argon2id$v=19$m=65536,t=1,p=12$QyX8MNKm4r8r7jbMWQ3/vA$/Z2JP7/ykKfsEyziGjD6TBTZLnOH8nV89gli9oeltBQ",
		},
		// add more cases here
	}

	for _, c := range cases {
		actual, err := HashPassword(c.input)
		// Check the length of the actual slice against the expected slice
		// if they don't match, use t.Errorf to print an error message
		// and fail the test
		if len(actual) != len(c.expected) {
			t.Errorf("Length of actual: %v does not equal length of expected: %v", actual, c.expected)
		}
		if err != nil {
			t.Errorf("There was an error creating the hash: %v", err)
		}
	}
}
