package domain

type Collection struct {
	ID         int64
	Name       string
	Overview   string
	PosterPath string
	Parts      []CollectionPart
}

type CollectionPart struct {
	ID          int64
	Title       string
	Overview    string
	PosterPath  string
	MediaType   string
	ReleaseDate string
	VoteAverage float32
}
