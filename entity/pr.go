package entity

type SimplePRList struct {
	Total int
	Items []*SimplePR
}
type SimplePR struct {
	RepoOwner string
	RepoName  string
	Number    int
	Title     string
	URL       string
	Author    string
	BaseRef   string
	HeadRef   string
}
