// Copyright (c) 2020-2022 TU Delft & Valentijn van de Beek <v.d.vandebeek@student.tudelft.nl> All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package model stores miscellaneous database entries
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

// ImageSetupMessage is a stripped down version of the ImageSetup
// model which can be used as a JSON response
type ImageSetupMessage struct {
	UUID    string
	Version uint64
	Update  bool
}
