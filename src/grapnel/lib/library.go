package lib

/*
Copyright (c) 2014 Eric Anderton <eric.t.anderton@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

import (
	"fmt"
	"go/build"
	log "grapnel/log"
	"io"
	"os"
	"path"
	"path/filepath"
)

// contains resolved factors from the parent depdendency specification
type Library struct {
	Dependency
	Version      *Version
	TempDir      string
	Provides     []string // imports provided by this library
	Dependencies []*Dependency
}

func NewLibrary(dep *Dependency) *Library {
	result := &Library{
		Provides:     make([]string, 0),
		Dependencies: make([]*Dependency, 0),
	}
	result.Dependency = *dep
	return result
}

func (self *Library) Install(installRoot string) error {
	// set up root target dir
	importPath := filepath.Join(installRoot, self.Import)
	log.Debug("installing to: %s", importPath)
	if err := os.MkdirAll(importPath, 0755); err != nil {
		log.Info("%s", err.Error())
		return fmt.Errorf("Could not create target directory: '%s'", importPath)
	}

	// move everything over
	if err := CopyFileTree(importPath, self.TempDir); err != nil {
		log.Info("%s", err.Error())
		return fmt.Errorf("Error while walking dependency file tree")
	}
	return nil
}

func (self *Library) Destroy() error {
	return os.Remove(self.TempDir)
}

func (self *Library) AddDependencies() error {
	if self.TempDir == "" {
		return nil // do nothing if there's nothing to search
	}

	// get dependencies via lockfile or grapnelfile
	if deplist, err := LoadGrapnelDepsfile(
		path.Join(self.TempDir, "grapnel-lock.toml"),
		path.Join(self.TempDir, "grapnel.toml")); err != nil {
		return err
	} else if deplist != nil {
		self.Dependencies = append(self.Dependencies, deplist...)
		return nil
	}

	// figure out the provided modules in this library
	if importPaths, err := GetDirectories(self.TempDir); err != nil {
		return err
	} else {
		// fully qualify the set of paths
		for _, path := range importPaths {
			self.Provides = append(self.Provides, self.Import+"/"+path)
		}
	}

	// attempt get dependencies via raw import statements instead
	pkg, err := build.ImportDir(self.TempDir, 0)
	if err != nil {
		log.Debug("Failed to get go imports for %v", err)
		log.Warn("No Go imports to process for %v", self.Import)
		return nil
	}

	// add all non std libs as dependencies of this lib
	for _, importName := range pkg.Imports {
		if IsStandardDependency(importName) {
			log.Debug("Ignoring import: %v", importName)
		} else {
			log.Warn("Adding secondary import: %v", importName)
			dep, err := NewDependency(importName, "", "")
			if err != nil {
				return err
			}
			self.Dependencies = append(self.Dependencies, dep)
		}
	}
	return nil
}

func (self *Library) ToToml(writer io.Writer) {
	fmt.Fprintf(writer, "\n[[dependencies]]\n")
	if self.Version.Major > 0 {
		fmt.Fprintf(writer, "version = \"%v\"\n", self.Version)
	} else {
		fmt.Fprintf(writer, "# Unversioned\n")
	}
	if self.Type != "" {
		fmt.Fprintf(writer, "type = \"%s\"\n", self.Type)
	}
	if self.Import != "" {
		fmt.Fprintf(writer, "import = \"%s\"\n", self.Import)
	}
	if self.Url != nil {
		fmt.Fprintf(writer, "url = \"%s\"\n", self.Url.String())
	}
	if self.Branch != "" {
		fmt.Fprintf(writer, "branch = \"%s\"\n", self.Branch)
	}
	if self.Tag != "" {
		// TODO: repair notification
		//if self.Dependency.Tag == "" && self.Version.Major == 0 {
		//  fmt.Fprintf(writer, "# Pinned to recent tip/head of repository\n")
		//}
		fmt.Fprintf(writer, "tag = \"%s\"\n", self.Tag)
	}
}

func (self *Library) ToDsd(writer io.Writer) {
	//TODO
}
