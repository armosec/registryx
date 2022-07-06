package quay

import "fmt"

func (res *QuayCatalogResponse) Transform(maxSize int) []string {
	repos := make([]string, 0, maxSize)
	for i := range res.Repositories {
		if maxSize == 0 || len(repos) < maxSize {
			repos = append(repos, fmt.Sprintf("%s/%s", res.Repositories[i].Namespace, res.Repositories[i].Name))
		}
	}

	return repos
}
