package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoundFloat(t *testing.T) {
	// Test rounding to 2 decimal places
	assert.Equal(t, 3.14, RoundFloat(3.14159, 2))
	assert.Equal(t, 0.00, RoundFloat(0.001, 2))
	assert.Equal(t, 5.00, RoundFloat(4.999, 2))
	assert.Equal(t, -1.23, RoundFloat(-1.234, 2))

	// Test rounding to 0 decimal places
	assert.Equal(t, 3.0, RoundFloat(3.14159, 0))
	assert.Equal(t, 0.0, RoundFloat(0.001, 0))
	assert.Equal(t, 5.0, RoundFloat(4.999, 0))

	// Test rounding to 4 decimal places
	assert.Equal(t, 3.1416, RoundFloat(3.14159, 4))
}

func TestGetTotalMemory(t *testing.T) {
	// Just verify that it returns a reasonable value
	totalMemory, _ := GetTotalMemory()
	assert.Greater(t, totalMemory.Total, uint64(0))
}

func TestStrToInt32(t *testing.T) {
	// Test with valid input
	assert.Equal(t, int32(123), StrToInt32("123"))

	// Test with invalid input
	assert.Equal(t, int32(0), StrToInt32("invalid"))
}

func TestInt32toStr(t *testing.T) {
	// Test with valid input
	assert.Equal(t, "123", Int32toStr(123))
}

func TestSortSlice(t *testing.T) {
	// Test with valid input
	sorted := SortSlice([]int32{3, 1, 2})
	assert.Equal(t, []int32{1, 2, 3}, sorted)
}

func TestContains(t *testing.T) {
	// Test with valid input
	slice := []string{"a", "b", "c"}
	assert.True(t, Contains(slice, "b"))
	assert.False(t, Contains(slice, "d"))
}

func TestGetScreenWidth(t *testing.T) {
	// Just verify that it returns a reasonable width
	width := GetScreenWidth()
	assert.GreaterOrEqual(t, width, 40) // Most terminals are at least 40 columns wide
}

func TestTruncateString(t *testing.T) {
	// Test with valid input
	assert.Equal(t, "1234567890", TruncateString("1234567890", 10))
	assert.Equal(t, "123456789", TruncateString("1234567890", 9))
	assert.Equal(t, "123456789", TruncateString("123456789", 9))
	assert.Equal(t, "123456789", TruncateString("123456789", 10))
	assert.Equal(t, "123456789", TruncateString("123456789", 10))
}

func TestUserExists(t *testing.T) {
	// Test with a user that should exist on most systems
	assert.True(t, UserExists("root"))

	// Test with a user that should not exist
	assert.False(t, UserExists("nonexistentuser123456789"))
}

func TestByteConverter(t *testing.T) {
	// Test with valid input
	assert.Equal(t, "1.00 KiB", ByteConverter(1024))
	assert.Equal(t, "1.00 MiB", ByteConverter(1048576))
	assert.Equal(t, "1.00 GiB", ByteConverter(1073741824))
	assert.Equal(t, "1.00 TiB", ByteConverter(1099511627776))
	assert.Equal(t, "1.00 PiB", ByteConverter(1125899906842624))
	assert.Equal(t, "1.00 EiB", ByteConverter(1152921504606847000))
}

func TestBtoI(t *testing.T) {
	assert.Equal(t, 1, BtoI(true))
	assert.Equal(t, 0, BtoI(false))
}

func TestStoI(t *testing.T) {
	assert.Equal(t, 1, StoI("test"))
	assert.Equal(t, 0, StoI(""))
}
func TestGetUnixTimestamp(t *testing.T) {
	// Just verify that it returns a reasonable timestamp
	timestamp := GetUnixTimestamp()
	assert.Greater(t, timestamp, int64(1609459200)) // Jan 1, 2021
}

func TestDetermineUsername(t *testing.T) {
	// Just verify that it returns a non-empty string
	username := DetermineUsername()
	assert.NotEmpty(t, username)
}

func TestFindDuration(t *testing.T) {
	duration := FindDuration(10)
	assert.Equal(t, int64(10), duration.Seconds)

	duration = FindDuration(60)
	assert.Equal(t, int64(1), duration.Minutes)

	duration = FindDuration(3600)
	assert.Equal(t, int64(1), duration.Hours)

	duration = FindDuration(86400)
	assert.Equal(t, int64(1), duration.Days)

	duration = FindDuration(2592000)
	assert.Equal(t, int64(30), duration.Days)

}

func TestDeleteSliceElement(t *testing.T) {
	// Test deleting an element from the middle
	slice := []string{"a", "b", "c", "d"}
	result := DeleteSliceElement(slice, 1)
	assert.Equal(t, []string{"a", "c", "d"}, result)

	// Test deleting the first element
	slice = []string{"a", "b", "c"}
	result = DeleteSliceElement(slice, 0)
	assert.Equal(t, []string{"b", "c"}, result)

	// Test deleting the last element
	slice = []string{"a", "b", "c"}
	result = DeleteSliceElement(slice, 2)
	assert.Equal(t, []string{"a", "b"}, result)

	// Test with an invalid index (should return the original slice)
	slice = []string{"a", "b", "c"}
	result = DeleteSliceElement(slice, 5)
	assert.Equal(t, []string{"a", "b", "c"}, result)

	// Test with an empty slice
	slice = []string{}
	result = DeleteSliceElement(slice, 0)
	assert.Equal(t, []string{}, result)
}
