package pathfinder

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"os"
	"regexp"
	"strings"
)

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
