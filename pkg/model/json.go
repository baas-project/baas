package model

// GitHubLogin represent the JSON structure sent by the GitHub user API
type GitHubLogin struct {
	Login             string
	ID                int
	NodeID            string
	AvatarURI         string
	GravatarID        string
	URI               string
	HTMLURI           string
	FollowersURI      string
	FollowingURI      string
	GistsURI          string
	StarredURI        string
	SubscriptionsURI  string
	OrganizationsURI  string
	ReposURI          string
	ReceivedEventsURI string
	Type              string
	SiteAdmin         bool
	Name              string
	Company           string
	Blog              string
	Location          string
	Email             string
	Hireable          bool
	Bio               string
	TwitterUsername   string
	PublicRepos       string
	PublicGists       string
	Followers         int
	Following         int
	CreatedAt         string
	UpdatedAt         string
}
