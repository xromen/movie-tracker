package domain

type Collection struct {
	TMDBID     int64 `json:"ID"`
	Name       string
	Overview   string
	PosterPath string
	Parts      []CollectionPart
}

type CollectionPart struct {
	TMDBID      int64 `json:"ID"`
	Title       string
	Overview    string
	PosterPath  string
	MediaType   string
	ReleaseDate string
	VoteAverage float32
	WatchStatus WatchStatus
}
