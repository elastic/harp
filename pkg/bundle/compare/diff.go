package compare

import (
	"fmt"
	"strings"

	bundlev1 "github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/bundle/secret"
	"github.com/elastic/harp/pkg/sdk/security"
)

const (
	// Add describes an operation where the traget path object has been added.
	Add string = "add"
	// Remove describes that entity has been removed.
	Remove string = "remove"
	// Replace describes an operation to replace content of the target path object.
	Replace string = "replace"
)

// DiffItem represents bundle comparison operations.
type DiffItem struct {
	Operation string `json:"op"`
	Type      string `json:"type"`
	Path      string `json:"path"`
	Value     string `json:"value,omitempty"`
}

// -----------------------------------------------------------------------------

// Diff calculates bundle differences.
func Diff(src, dst *bundlev1.Bundle) ([]DiffItem, error) {
	// Check arguments
	if src == nil {
		return nil, fmt.Errorf("unable to diff with a nil source")
	}
	if dst == nil {
		return nil, fmt.Errorf("unable to diff with a nil destination")
	}

	var diffs = []DiffItem{}

	// Index source packages
	var srcIndex = map[string]*bundlev1.Package{}
	for _, srcPkg := range src.Packages {
		srcIndex[srcPkg.Name] = srcPkg
	}

	// Index destination packages
	var dstIndex = map[string]*bundlev1.Package{}
	for _, dstPkg := range dst.Packages {
		dstIndex[dstPkg.Name] = dstPkg
		if _, ok := srcIndex[dstPkg.Name]; !ok {
			// Package has been added
			diffs = append(diffs, DiffItem{
				Operation: Add,
				Type:      "package",
				Path:      dstPkg.Name,
			})

			// Add keys
			for _, s := range dstPkg.Secrets.Data {
				// Unpack secret value
				var data string
				if err := secret.Unpack(s.Value, &data); err != nil {
					return nil, fmt.Errorf("unable to unpack '%s' - '%s' secret value: %w", dstPkg.Name, s.Key, err)
				}

				diffs = append(diffs, DiffItem{
					Operation: Add,
					Type:      "secret",
					Path:      fmt.Sprintf("%s#%s", dstPkg.Name, s.Key),
					Value:     data,
				})
			}
		}
	}

	// Compute package changes
	for n, sp := range srcIndex {
		dp, ok := dstIndex[n]
		if !ok {
			// Not exist in destination bundle
			diffs = append(diffs, DiffItem{
				Operation: Remove,
				Type:      "package",
				Path:      sp.Name,
			})
			continue
		}

		// Index secret data
		srcSecretIndex := map[string]*bundlev1.KV{}
		for _, ss := range sp.Secrets.Data {
			srcSecretIndex[ss.Key] = ss
		}
		dstSecretIndex := map[string]*bundlev1.KV{}
		for _, ds := range dp.Secrets.Data {
			dstSecretIndex[ds.Key] = ds
			oldValue, ok := srcSecretIndex[ds.Key]
			if !ok {
				// Secret has been added
				var data string
				if err := secret.Unpack(ds.Value, &data); err != nil {
					return nil, fmt.Errorf("unable to unpack '%s' - '%s' secret value: %w", dp.Name, ds.Key, err)
				}

				diffs = append(diffs, DiffItem{
					Operation: Add,
					Type:      "secret",
					Path:      fmt.Sprintf("%s#%s", dp.Name, ds.Key),
					Value:     data,
				})
				continue
			}

			// Skip if key does not match
			if !strings.EqualFold(oldValue.Key, ds.Key) {
				continue
			}

			// Compare values
			if !security.SecureCompare(oldValue.Value, ds.Value) {
				// Secret has been replaced
				var data string
				if err := secret.Unpack(ds.Value, &data); err != nil {
					return nil, fmt.Errorf("unable to unpack '%s' - '%s' secret value: %w", dp.Name, ds.Key, err)
				}

				diffs = append(diffs, DiffItem{
					Operation: Replace,
					Type:      "secret",
					Path:      fmt.Sprintf("%s#%s", dp.Name, ds.Key),
					Value:     data,
				})
			}
		}
	}

	// No error
	return diffs, nil
}
