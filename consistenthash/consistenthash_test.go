package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	hash := New(Replicas(3), HashFunc(func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	}))
	hash.Add("6", "4", "2")
	testCases := map[string]string{
		"2":  "2", // Hash(2)=2,   Match 2
		"11": "2", // Hash(11)=11, Match 12(2)
		"23": "4", // Hash(23)=23, Match 24(4)
		"27": "2", // Hash(27)=27, Match 2(Wrap around)
	}
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}
	hash.Add("8")
	testCases["27"] = "8"
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}
}