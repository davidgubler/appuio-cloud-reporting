package sourcekey

import (
	"fmt"
	"strings"
)

const elementSeparator = ":"

// SourceKey represents a source key to look up dimensions objects (currently queries and products).
// It implements the lookup logic found in https://kb.vshn.ch/appuio-cloud/references/architecture/metering-data-flow.html#_system_idea.
type SourceKey struct {
	Query     string
	Zone      string
	Tenant    string
	Namespace string

	Class string
}

// Parse parses a source key in the format of "query:zone:tenant:namespace:class" or "query:zone:tenant:namespace".
func Parse(raw string) (SourceKey, error) {
	parts := strings.Split(raw, elementSeparator)
	if len(parts) == 4 {
		return SourceKey{parts[0], parts[1], parts[2], parts[3], ""}, nil
	} else if len(parts) == 5 {
		return SourceKey{parts[0], parts[1], parts[2], parts[3], parts[4]}, nil
	}

	return SourceKey{}, fmt.Errorf("expected key with 4 to 5 elements separated by `%s` got %d", elementSeparator, len(parts))
}

// String returns the string representation "query:zone:tenant:namespace:class" of the key.
func (k SourceKey) String() string {
	elements := []string{k.Query, k.Zone, k.Tenant, k.Namespace}
	if k.Class != "" {
		elements = append(elements, k.Class)
	}
	return strings.Join(elements, elementSeparator)
}

// LookupKeys generates lookup keys for a dimension object in the database.
// The logic is described here: https://kb.vshn.ch/appuio-cloud/references/architecture/metering-data-flow.html#_system_idea
func (k SourceKey) LookupKeys() []string {
	return generateSourceKeys(k.Query, k.Zone, k.Tenant, k.Namespace, k.Class)
}

func generateSourceKeys(query, zone, tenant, namespace, class string) []string {
	keys := make([]string, 0)
	base := []string{query, zone, tenant, namespace}
	wildcardPositions := []int{1, 2}

	if class != "" {
		wildcardPositions = append(wildcardPositions, 3)
		base = append(base, class)
	}

	for i := len(base); i > 0; i-- {
		keys = append(keys, strings.Join(base[:i], elementSeparator))

		for j := 1; j < len(wildcardPositions)+1; j++ {
			perms := combinations(wildcardPositions, j)
			for _, wcpos := range reverse(perms) {
				elements := append([]string{}, base[:i]...)
				for _, p := range wcpos {
					elements[p] = "*"
				}
				keys = append(keys, strings.Join(elements, elementSeparator))
			}
		}
		if i > 2 {
			wildcardPositions = wildcardPositions[:len(wildcardPositions)-1]
		}
	}

	return keys
}

func reverse(s [][]int) [][]int {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func combinations(iterable []int, r int) (rt [][]int) {
	pool := iterable
	n := len(pool)

	if r > n {
		return
	}

	indices := make([]int, r)
	for i := range indices {
		indices[i] = i
	}

	result := make([]int, r)
	for i, el := range indices {
		result[i] = pool[el]
	}
	s2 := make([]int, r)
	copy(s2, result)
	rt = append(rt, s2)

	for {
		i := r - 1
		for ; i >= 0 && indices[i] == i+n-r; i -= 1 {
		}

		if i < 0 {
			return
		}

		indices[i] += 1
		for j := i + 1; j < r; j += 1 {
			indices[j] = indices[j-1] + 1
		}

		for ; i < len(indices); i += 1 {
			result[i] = pool[indices[i]]
		}
		s2 = make([]int, r)
		copy(s2, result)
		rt = append(rt, s2)
	}
}
