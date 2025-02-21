package rid_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"{{.D.ModuleName}}/internal/pkg/rid"
)

// Mock Salt function used for testing
func Salt() string {
	return "staticSalt"
}

func TestResourceID_String(t *testing.T) {
	// Test converting UserID to a string
	rid := rid.UserID
	assert.Equal(t, "user", rid.String(), "UserID.String() should return 'user'")
}

func TestResourceID_New(t *testing.T) {
	// Test if the generated ID has the correct prefix
	rid := rid.UserID
	uniqueID := rid.New(1)

	assert.True(t, len(uniqueID) > 0, "Generated ID should not be empty")
	assert.Contains(t, uniqueID, "user-", "Generated ID should start with 'user-' prefix")

	// Generate another unique identifier to ensure uniqueness
	anotherID := rid.New(2)
	assert.NotEqual(t, uniqueID, anotherID, "Generated IDs should be unique")
}

func BenchmarkResourceID_New(b *testing.B) {
	// Performance benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rid := rid.UserID
		_ = rid.New(uint64(i))
	}
}

func FuzzResourceID_New(f *testing.F) {
	// Add preset test data
	f.Add(uint64(1))      // Add a seed value `counter` as 1
	f.Add(uint64(123456)) // Add a larger seed value

	f.Fuzz(func(t *testing.T, counter uint64) {
		// Test the New method for UserID
		result := rid.UserID.New(counter)

		// Assert that the result is not empty
		assert.NotEmpty(t, result, "The generated unique identifier should not be empty")

		// Assert that the result contains the correct resource identifier prefix
		assert.Contains(t, result, rid.UserID.String()+"-", "The generated unique identifier should contain the correct prefix")

		// Assert that the prefix does not overlap with the uniqueStr part
		splitParts := strings.SplitN(result, "-", 2)
		assert.Equal(t, rid.UserID.String(), splitParts[0], "The prefix part of the result should correctly match the UserID")

		// Assert that the generated ID has a fixed length (based on NewCode configuration)
		if len(splitParts) == 2 {
			assert.Equal(t, 6, len(splitParts[1]), "The unique identifier part should have a length of 6")
		} else {
			t.Errorf("The format of the generated unique identifier does not meet expectation")
		}
	})
}
