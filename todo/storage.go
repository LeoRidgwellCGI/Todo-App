package todo

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
)

// Save writes the list of items to a JSON file.
func Save(list []Item, path string) error {
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Load reads items from a JSON file.
// If the file does not exist, returns an empty list.
func Load(path string) ([]Item, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []Item{}, nil
		}
		return nil, err
	}
	if len(b) == 0 {
		return []Item{}, nil
	}
	var list []Item
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, err
	}
	return list, nil
}
