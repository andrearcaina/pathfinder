package pathfinder

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

func scanDependencies(rootPath string, flags Config) ([]DependencyFile, error) {
	var depFiles []DependencyFile
	var mu sync.Mutex
	var wg sync.WaitGroup

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// if the user didn't set recursive flag, skip subdirectories
		if !flags.RecursiveFlag && path != flags.PathFlag && d.IsDir() {
			return filepath.SkipDir
		}

		// if the user set max depth, skip directories deeper than max depth
		if flags.RecursiveFlag && flags.MaxDepthFlag != -1 {
			relPath, _ := filepath.Rel(flags.PathFlag, path)
			depth := len(strings.Split(relPath, string(filepath.Separator)))
			if depth > flags.MaxDepthFlag {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		name := d.Name()

		if !flags.HiddenFlag && strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if excludeDir(name) {
				return filepath.SkipDir
			}
			return nil
		}

		if isBinary(name) {
			return nil
		}

		ext := hasNoExt(name)
		if ext == "" {
			return nil
		}

		var depType string
		var shouldProcess bool

		switch {
		case name == "go.mod":
			depType = "Go Modules"
			shouldProcess = true
		case name == "go.sum":
			return nil

		case name == "package.json":
			depType = "npm/yarn"
			shouldProcess = true

		case name == "requirements.txt":
			depType = "pip"
			shouldProcess = true

		case name == "pom.xml":
			depType = "Maven"
			shouldProcess = true

		case strings.HasSuffix(name, ".csproj"):
			depType = ".NET/NuGet"
			shouldProcess = true
		default:
			return nil
		}

		if shouldProcess {
			wg.Add(1)
			go func(filePath, fileType string) {
				defer wg.Done()

				var deps []string
				var err error

				switch {
				case strings.HasSuffix(filePath, "go.mod"):
					deps, err = scanGoMod(filePath)
				case strings.HasSuffix(filePath, "package.json"):
					deps, err = scanPackageJSON(filePath)
				case strings.HasSuffix(filePath, "requirements.txt"):
					deps, err = scanRequirementsTxt(filePath)
				case strings.HasSuffix(filePath, "pom.xml"):
					deps, err = scanPomXML(filePath)
				case strings.HasSuffix(filePath, ".csproj"):
					deps, err = scanCsproj(filePath)
				}

				if err != nil {
					return // continue on error
				}

				if len(deps) > 0 {
					mu.Lock()
					depFiles = append(depFiles, DependencyFile{
						Path:         filePath,
						Type:         fileType,
						Dependencies: deps,
					})
					mu.Unlock()
				}
			}(path, depType)
		}

		return nil
	})

	wg.Wait()
	return depFiles, err
}

func scanGoMod(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var deps []string
	scanner := bufio.NewScanner(file)
	inRequireBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true
			continue
		}

		if inRequireBlock && line == ")" {
			inRequireBlock = false
			continue
		}

		if strings.HasPrefix(line, "require ") && !inRequireBlock {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				deps = append(deps, parts[1])
			}
		} else if inRequireBlock && line != "" && !strings.HasPrefix(line, "//") {
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				deps = append(deps, parts[0])
			}
		}
	}

	return deps, scanner.Err()
}

func scanPackageJSON(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.NewDecoder(file).Decode(&pkg); err != nil {
		return nil, err
	}

	var deps []string
	for name := range pkg.Dependencies {
		deps = append(deps, name)
	}
	for name := range pkg.DevDependencies {
		deps = append(deps, name)
	}

	return deps, nil
}

func scanRequirementsTxt(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var deps []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		re := regexp.MustCompile(`^([a-zA-Z0-9_-]+)`)
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			deps = append(deps, matches[1])
		}
	}

	return deps, scanner.Err()
}

func scanPomXML(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var pom struct {
		Dependencies struct {
			Dependency []struct {
				GroupID    string `xml:"groupId"`
				ArtifactID string `xml:"artifactId"`
			} `xml:"dependency"`
		} `xml:"dependencies"`
	}

	if err := xml.NewDecoder(file).Decode(&pom); err != nil {
		return nil, err
	}

	var deps []string
	for _, dep := range pom.Dependencies.Dependency {
		if dep.GroupID != "" && dep.ArtifactID != "" {
			deps = append(deps, dep.GroupID+":"+dep.ArtifactID)
		}
	}

	return deps, nil
}

func scanCsproj(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var project struct {
		ItemGroup []struct {
			PackageReference []struct {
				Include string `xml:"Include,attr"`
			} `xml:"PackageReference"`
		} `xml:"ItemGroup"`
	}

	if err := xml.NewDecoder(file).Decode(&project); err != nil {
		return nil, err
	}

	var deps []string
	for _, group := range project.ItemGroup {
		for _, pkg := range group.PackageReference {
			if pkg.Include != "" {
				deps = append(deps, pkg.Include)
			}
		}
	}

	return deps, nil
}
