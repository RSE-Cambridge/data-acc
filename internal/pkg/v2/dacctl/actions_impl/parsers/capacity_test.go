package parsers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseSize(t *testing.T) {
	size, err := ParseSize("10GiB")
	assert.Nil(t, err)
	assert.Equal(t, 10737418240, size)

	size, err = ParseSize("10GB")
	assert.Nil(t, err)
	assert.Equal(t, 10000000000, size)

	size, err = ParseSize("10 GB")
	assert.Nil(t, err)
	assert.Equal(t, 10000000000, size)

	size, err = ParseSize("10B")
	assert.Equal(t, "unable to parse size: 10B", err.Error())

	size, err = ParseSize("10.1234567MB")
	assert.Nil(t, err)
	assert.Equal(t, 10123457, size)

	size, err = ParseSize("1TiB")
	assert.Nil(t, err)
	assert.Equal(t, 1099511627776, size)

	size, err = ParseSize("1TB")
	assert.Nil(t, err)
	assert.Equal(t, 1000000000000, size)

	size, err = ParseSize("1MiB")
	assert.Nil(t, err)
	assert.Equal(t, 1048576, size)

	size, err = ParseSize("AMiB")
	assert.Equal(t, "strconv.ParseFloat: parsing \"A\": invalid syntax", err.Error())
	assert.Equal(t, 0, size)
}

func TestParseCapacityBytes(t *testing.T) {
	pool, size, err := ParseCapacityBytes("foo:1MB")
	assert.Nil(t, err)
	assert.Equal(t, "foo", pool)
	assert.Equal(t, 1000000, size)

	pool, size, err = ParseCapacityBytes("foo : 1 MB")
	assert.Equal(t, "foo", pool)
	assert.Equal(t, 1000000, size)
	assert.Nil(t, err)

	pool, size, err = ParseCapacityBytes("foo1MB")
	assert.Equal(t, "must format capacity correctly and include pool", err.Error())
	assert.Equal(t, "", pool)
	assert.Equal(t, 0, size)

	pool, size, err = ParseCapacityBytes("foo:1B")
	assert.Equal(t, "unable to parse size: 1B", err.Error())
	assert.Equal(t, "", pool)
	assert.Equal(t, 0, size)
}
