package models

type AdvancedFilter struct {
	Name     string   `json:"name" bson:"name"`
	Contents []string `json:"contents" bson:"contents"`
}

type ExcludedPathsFilter struct {
	ExcludedPaths []string `json:"excludedPaths" bson:"excludedPaths"`
}
