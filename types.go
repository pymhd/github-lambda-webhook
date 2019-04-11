package main

type Repo struct {
	ID        int    `json:"id"`
	FullName  string `json:"full_name"`
	ShortName string `json:"name"`
}

type Branch struct {
	Ref  string `json:"ref"`
	Sha  string `json:"sha"`
	Repo Repo   `json:"repo"`
}

type User struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Type  string `json:"type"`
}

type Label struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type PullRequestPayload struct {
	Action      string `json:"action"`
	Number      int    `json:"number"`
	PullRequest struct {
		Number int    `json:"number"`
		State  string `json:"state"`
		User   User   `json:"user"`
		Base   Branch `json:"base"`
		Head   Branch `json:"head"`
		Sender User   `json:"sender"`
		Label  Label  `json:"label"`
	} `json:"pull_request"`
	Label  Label `json:"label"`
	Sender User  `json:"sender"`
}
