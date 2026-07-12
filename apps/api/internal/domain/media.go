package domain

type Media struct {
	ID          int64
	Title       string
	Overview    string
	ReleaseDate string
	PosterPath  string
	VoteAverage float32
	VoteCount   int64
	Type        MediaType
}

type UserMedia struct {
	Media  *Media
	Status WatchStatus
	Rating *int
}

type Video struct {
	ID          string
	Iso639_1    string
	Iso3166_1   string
	Key         string
	Name        string
	Official    bool
	PublishedAt string
	Site        string
	Size        int
	Type        string
}

type MediaType string

const (
	MediaTypeTV    MediaType = "tv"
	MediaTypeMovie MediaType = "movie"
)

func (s MediaType) IsValid() bool {
	switch s {
	case MediaTypeTV, MediaTypeMovie:
		return true
	}
	return false
}

type WatchStatus string

const (
	WatchStatusWatched     WatchStatus = "watched"
	WatchStatusWantToWatch WatchStatus = "want_to_watch"
	WatchStatusFavorite    WatchStatus = "favorite"
)

func (s WatchStatus) IsValid() bool {
	switch s {
	case WatchStatusWatched, WatchStatusWantToWatch, WatchStatusFavorite:
		return true
	}
	return false
}
