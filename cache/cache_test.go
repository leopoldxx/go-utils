package cache_test

import (
	"testing"
	"time"

	. "github.com/leopoldxx/go-utils/cache"
)

func TestCache(t *testing.T) {
	count := 0
	cb := func(key Key, value Value) {
		t.Log(key, value)
		count++
	}

	cache := NewCacheWithConfig(Config{MaxLen: 2, Callback: cb})

	testCases := []struct {
		op           string
		key          string
		value        string
		expectExists bool
		expectLen    int
		expectValue  string
	}{
		{
			"put",
			"testkey1",
			"testvalue1",
			true,
			1,
			"testvalue1",
		},
		{
			"put",
			"testkey2",
			"testvalue2",
			true,
			2,
			"testvalue2",
		},
		{
			"get",
			"testkey1",
			"",
			true,
			2,
			"testvalue1",
		},
		{
			"get",
			"testkey2",
			"",
			true,
			2,
			"testvalue2",
		},
		{
			"put",
			"testkey3",
			"testvalue3",
			true,
			2,
			"testvalue3",
		},
		{
			"get",
			"testkey1",
			"",
			false,
			2,
			"",
		},
		{
			"get",
			"testkey2",
			"",
			true,
			2,
			"testvalue2",
		},
		{
			"del",
			"testkey2",
			"",
			false,
			1,
			"",
		},
		{
			"put",
			"testkey4",
			"testvalue4",
			true,
			2,
			"testvalue4",
		},
		{
			"put",
			"testkey5",
			"testvalue5",
			true,
			2,
			"testvalue5",
		},
		{
			"get",
			"testkey3",
			"",
			false,
			2,
			"",
		},
	}
	for _, tc := range testCases {
		switch tc.op {
		case "get":
			val, ok := cache.Get(tc.key)
			if ok != tc.expectExists {
				t.Fatalf("test key %s exist status failed, expect %v, got %v", tc.key, tc.expectExists, ok)
			}
			if tc.expectExists && val != tc.expectValue {
				t.Fatalf("test key %s value failed, expect %v, got %v", tc.key, tc.expectValue, val)
			}
			if tc.expectLen != cache.Len() {
				t.Fatalf("test key %s len failed, expect %v, got %v", tc.key, tc.expectLen, cache.Len())
			}
		case "put":
			cache.Put(tc.key, tc.value)
			val, ok := cache.Get(tc.key)
			if ok != tc.expectExists {
				t.Fatalf("test key %s exist status failed, expect %v, got %v", tc.key, tc.expectExists, ok)
			}
			if tc.expectExists && val != tc.expectValue {
				t.Fatalf("test key %s value failed, expect %v, got %v", tc.key, tc.expectValue, val)
			}
			if tc.expectLen != cache.Len() {
				t.Fatalf("test key %s len failed, expect %v, got %v", tc.key, tc.expectLen, cache.Len())
			}
		case "del":
			cache.Del(tc.key)
			val, ok := cache.Get(tc.key)
			if ok != tc.expectExists {
				t.Fatalf("test key %s exist status failed, expect %v, got %v", tc.key, tc.expectExists, ok)
			}
			if tc.expectExists && val != tc.expectValue {
				t.Fatalf("test key %s value failed, expect %v, got %v", tc.key, tc.expectValue, val)
			}
			if tc.expectLen != cache.Len() {
				t.Fatalf("test key %s len failed, expect %v, got %v", tc.key, tc.expectLen, cache.Len())
			}

		}

	}

}
func TestCacheTime(t *testing.T) {
	count := 0
	cb := func(key Key, value Value) {
		t.Log(key, value)
		count++
	}

	cache := NewCacheWithConfig(Config{MaxLen: 2, Callback: cb, CacheTime: time.Second})

	testCases := []struct {
		op           string
		key          string
		value        string
		st           time.Duration
		expectExists bool
		expectLen    int
		expectValue  string
	}{
		{
			"put",
			"testkey1",
			"testvalue1",
			0,
			true,
			1,
			"testvalue1",
		},
		{
			"put",
			"testkey2",
			"testvalue2",
			time.Second,
			false,
			1,
			"testvalue2",
		},
		{
			"get",
			"testkey1",
			"",
			0,
			false,
			0,
			"",
		},
	}
	for _, tc := range testCases {
		switch tc.op {
		case "get":
			if tc.st != 0 {
				time.Sleep(tc.st)
			}
			val, ok := cache.Get(tc.key)
			if ok != tc.expectExists {
				t.Fatalf("test key %s exist status failed, expect %v, got %v", tc.key, tc.expectExists, ok)
			}
			if tc.expectExists && val != tc.expectValue {
				t.Fatalf("test key %s value failed, expect %v, got %v", tc.key, tc.expectValue, val)
			}
			if tc.expectLen != cache.Len() {
				t.Fatalf("test key %s len failed, expect %v, got %v", tc.key, tc.expectLen, cache.Len())
			}
		case "put":
			cache.Put(tc.key, tc.value)
			if tc.st != 0 {
				time.Sleep(tc.st)
			}
			val, ok := cache.Get(tc.key)
			if ok != tc.expectExists {
				t.Fatalf("test key %s exist status failed, expect %v, got %v", tc.key, tc.expectExists, ok)
			}
			if tc.expectExists && val != tc.expectValue {
				t.Fatalf("test key %s value failed, expect %v, got %v", tc.key, tc.expectValue, val)
			}
			if tc.expectLen != cache.Len() {
				t.Fatalf("test key %s len failed, expect %v, got %v", tc.key, tc.expectLen, cache.Len())
			}
		}
	}
}
