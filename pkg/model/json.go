package model

type GitHubLogin struct {
	Login             string
	Id                int
	NodeId            string
	AvatarUri         string
	GravatarId        string
	Uri               string
	HtmlUri           string
	FollowersUri      string
	FollowingUri      string
	GistsUri          string
	StarredUri        string
	SubscriptionsUri  string
	OrganizationsUri  string
	ReposUri          string
	ReceivedEventsUri string
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
