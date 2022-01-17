package sourcekey

import (
	"fmt"
	"strings"

	"github.com/ernestosuarez/itertools"

	// TODO: Fails with unknown flag when executing `make test` if not registered.
	_ "github.com/appuio/appuio-cloud-reporting/pkg/db/flag"
)

const elementSeparator = ":"

type SourceKey struct {
	Query     string
	Zone      string
	Tenant    string
	Namespace string

	Class string
}

func Parse(raw string) (SourceKey, error) {
	parts := strings.Split(raw, elementSeparator)
	if len(parts) == 4 {
		return SourceKey{parts[0], parts[1], parts[2], parts[3], ""}, nil
	} else if len(parts) == 5 {
		return SourceKey{parts[0], parts[1], parts[2], parts[3], parts[4]}, nil
	}

	return SourceKey{}, fmt.Errorf("expected key with 4 to 5 elements separated by `%s` got %d", elementSeparator, len(parts))
}

func (k SourceKey) String() string {
	elements := []string{k.Query, k.Zone, k.Tenant, k.Namespace}
	if k.Class != "" {
		elements = append(elements, k.Class)
	}
	return strings.Join(elements, elementSeparator)
}

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

func combinations(s []int, r int) [][]int {
	collected := make([][]int, 0)
	for e := range itertools.CombinationsInt(s, r) {
		collected = append(collected, e)
	}
	return collected
}

func reverse(s [][]int) [][]int {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
