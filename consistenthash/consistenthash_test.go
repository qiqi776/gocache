package consistenthash

import (
	"strconv"
	"strings"
	"testing"
)

func TestHashing(t *testing.T) {
	hashFunc := func(data []byte) uint32 {
		s := string(data)
		if idx := strings.Index(s, "-"); idx != -1 {
			nodePart := s[:idx]
			replicaPart := s[idx+1:]
			nodeVal, _ := strconv.Atoi(nodePart)
			replicaVal, _ := strconv.Atoi(replicaPart)
			return uint32(nodeVal + replicaVal*10)
		}
		val, _ := strconv.Atoi(s)
		return uint32(val)
	}
	cfg := &Config{
		HashFunc:        hashFunc,
		DefaultReplicas: 3,
		MinReplicas:     1,
		MaxReplicas:     10,
	}
	hash := New(WithConfig(cfg))
	hash.Add("6", "4", "2")
	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}
	for k, v := range testCases {
		if node := hash.Get(k); node != v {
			t.Errorf("Asking for %s, should have yielded %s, but got %s", k, v, node)
		}
	}
	hash.Add("8")
	testCases["27"] = "8"
	for k, v := range testCases {
		if node := hash.Get(k); node != v {
			t.Errorf("Asking for %s, should have yielded %s, but got %s", k, v, node)
		}
	}
}