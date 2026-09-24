package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/benjuh/stew/util"
	"time"
)

const (
	green       = "\033[32m"
	red         = "\033[31m"
	property    = "\033[36m"
	description = "\033[35m"
	pathColor   = "\033[33m"
	reset       = "\033[0m"
)

// Stew is a struct that represents a stew
type Stew struct {
	Name        string
	Description string
	Path        string
	CreatedAt   time.Time
	Source      string
	SourceType  string
	Version     string
	Checksum    string
	Revision    string
	UpdatedAt   time.Time
}

// Stews is a slice of Stew structs
type Stews []Stew

// Len returns the length of the slice
func (st Stews) Len() int {
	return len(st)
}

// Print prints the stew
func (st Stew) Print() {
	fmt.Printf("\n%vName%v        -> %s\n", property, reset, st.Name)
	fmt.Printf("%vDescription%v -> '%v%s%v'\n", property, reset, description, st.Description, reset)

	if util.CheckIfDirExists(st.Path) {
		fmt.Printf("%vPath%v        -> '%v%s%v'\n", property, reset, pathColor, st.Path, reset)
	} else {
		fmt.Printf("%vPath%v        -> '%v%s%v'\n", property, reset, red, st.Path, reset)
		fmt.Printf("^^^ This stew location has been deleted\nEither edit the path or remove the stew\n")
	}
	fmt.Printf("%vCreated At%v  -> %s\n\n", property, reset, util.FormatTime(st.CreatedAt))
}

// PrintAdded prints the stew with the added properties
func (st Stew) PrintAdded() {
	fmt.Printf("\n`%v%s%v` has been added to your stews 🎉\n\n", green, st.Name, reset)
	fmt.Printf(" %vName%v        -> %s\n", property, reset, st.Name)
	fmt.Printf(" %vDescription%v -> '%v%s%v'\n", property, reset, description, st.Description, reset)
	fmt.Printf(" %vPath%v        -> '%v%s%v'\n", property, reset, pathColor, st.Path, reset)
	fmt.Printf(" %vCreated At%v  -> %s\n\n", property, reset, util.FormatTime(st.CreatedAt))
}

// PrintRemoved prints the stew with the removed properties
func (st Stew) PrintRemoved() {
	fmt.Printf("\n`%v%s%v` has been removed from your stews 👋\n", red, st.Name, reset)
}

// Add adds a stew to the slice
func (st *Stews) Add(name string, description string, path string) error {
	if name == "" {
		return errors.New("stew name cannot be empty")
	}
	var err error
	path, err = util.GetPath(path)
	if err != nil {
		return err
	}
	if st.doesStewExist(path) {
		return fmt.Errorf("stew path already exists: %s", path)
	}

	if st.doesStewExistWithName(name) {
		return fmt.Errorf("stew name already exists: %s", name)
	}

	stew := Stew{
		Name:        name,
		Description: description,
		Path:        path,
		CreatedAt:   time.Now(),
	}
	*st = append(*st, stew)

	stew.PrintAdded()
	return nil
}

// doesStewExist checks if a stew already exists in the slice
func (st *Stews) doesStewExist(path string) bool {
	for _, s := range *st {
		if s.Path == path {
			return true
		}
	}
	return false
}

func (st *Stews) doesStewExistWithName(name string) bool {
	for _, s := range *st {
		if s.Name == name {
			return true
		}
	}
	return false
}

// Remove removes a stew from the slice
func (st *Stews) Remove(i int) error {
	ls := *st
	if i < 0 || i >= len(ls) {
		return errors.New("invalid index")
	}

	ls[i].PrintRemoved()
	*st = append(ls[:i], ls[i+1:]...)
	return nil
}

// RemoveByName removes a stew from the slice by name
func (st *Stews) RemoveByName(name string) error {
	for i, s := range *st {
		if s.Name == name {
			return st.Remove(i)
		}
	}
	err := fmt.Sprintf("\n%vstew with name `%s` not found%v", red, name, reset)
	return errors.New(err)
}

// Get returns a stew from the slice by name
func (st Stews) Get(i int) (*Stew, error) {
	if i < 0 || i >= len(st) {
		err := fmt.Sprintf("\n%vstew of index %d not found%v", red, i, reset)
		return &Stew{}, errors.New(err)
	}

	return &st[i], nil
}

// GetByName returns a stew from the slice by name
func (st Stews) GetByName(name string) (*Stew, error) {
	for i, t := range st {
		if t.Name == name {
			return st.Get(i)
		}
	}
	err := fmt.Sprintf("\n%vstew with name `%s` not found%v", red, name, reset)
	return &Stew{}, errors.New(err)
}

// Edit edits a stew's name, description, or path
func (st *Stew) Edit(name, description, path string) error {

	if name != "" {
		fmt.Printf("\nName Changed: %s -> %v%s%v\n", st.Name, green, name, reset)
		st.Name = name
	}

	if description != "" {
		fmt.Printf("\nDescription Changed: %s -> %v%s%v\n", st.Description, green, description, reset)
		st.Description = description
	}

	if path != "" {
		path, err := util.GetPath(path)
		if err != nil {
			err := fmt.Sprintf("%v", err)
			return errors.New(err)
		}
		fmt.Printf("\nPath Changed: %s -> %v%s%v\n", st.Path, green, path, reset)
		st.Path = path
	}

	return nil
}

// List prints the stews
func (st *Stews) List() {
	// create a table of the stews aligned in columns of ID, Name, Description, CreatedAt
	if len(*st) == 0 {
		fmt.Println("\nNo stews found")
		return
	}

	fmt.Printf("\n%v%v%v\n", property, "Stews:", reset)
	for i, s := range *st {
		fmt.Println()
		fmt.Printf(" %v%v%v [%v%v%v]", green, s.Name, reset, property, i, reset)
		fmt.Printf(": '%v%s%v'\n", description, s.Description, reset)

		if util.CheckIfDirExists(s.Path) {
			fmt.Printf(" %v%s%v\n", pathColor, s.Path, reset)
		} else {
			fmt.Printf(" %v%s%v\n", red, s.Path, reset)
			fmt.Printf(" ^^^ This stew location has been deleted\n Either edit the path or remove the stew\n")
		}
		fmt.Printf(" %s\n", util.FormatTime(s.CreatedAt))
	}

}

// Load loads stews from a file
func (st *Stews) Load(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	if len(file) == 0 {
		return nil
	}

	err = json.Unmarshal(file, st)
	if err != nil {
		return err
	}

	return nil

}

// Save saves stews to a file
func (st *Stews) Save(path string) error {
	file, err := json.Marshal(st)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".stews-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(0644); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(file); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
