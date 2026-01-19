package validation

import (
    "testing"
)

// BenchmarkIsValidUsername is a performance benchmark test case.
func BenchmarkIsValidUsername(b *testing.B) {
    // Define a set of usernames for testing, including valid and invalid inputs.
    testUsernames := []string{
        "valid_user123",         // Valid, regular input
        "user_too_long_example", // Length exceeds 20
        "sh",                    // Length less than 3
        "in*valid",              // Contains invalid characters
        "12345678901234567890",  // Valid, exactly 20 characters
    }

    // Reset the timer.
    b.ResetTimer()

    // Performance benchmark test.
    for i := 0; i < b.N; i++ {
        // Simulate different test cases.
        for _, username := range testUsernames {
            isValidUsername(username)
        }
    }
}
