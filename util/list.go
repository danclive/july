package util

type ListParams struct {
	Start  int    `form:"start" json:"start"`
	Limit  int    `form:"limit" json:"limit"`
	Search string `form:"search" json:"search"`
	SortBy string `form:"sort_by" json:"sort_by"`
	Asc    bool   `form:"asc" json:"asc"`
	Label  string `form:"label" json:"label"`
}
