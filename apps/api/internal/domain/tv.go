package domain

// type TVShow struct {
// 	ID          int64
// 	Title       string
// 	Overview    string
// 	ReleaseDate string
// 	PosterPath  string
// 	Popularity  float32
// 	VoteAverage float32
// }

// type UserTVShow struct {
// 	TVShowID int64
// 	TVShow   *TVShow
// 	Status   WatchStatus
// 	Rating   *int
// }

type TVShowDetail struct {
	ID                     int64
	Title                  string
	Overview               string
	ReleaseDate            string
	LastEpisodeReleaseDate string
	NextEpisodeReleaseDate string
	PosterPath             string
	Popularity             float32
	VoteAverage            float32
	VoteCount              int64
	UserRating             *int
	UserStatus             WatchStatus
	Genres                 []Genre
	OriginalLanguage       string
	OriginCountry          []string
	ProductionCountries    []string
	OriginalTitle          string
	Status                 string
	Videos                 []Video
	NumberOfSeasons        int
	NumberOfEpisodes       int
	Seasons                []Season
}

type Season struct {
	ID           int64
	ReleaseDate  string
	EpisodeCount int
	Title        string
	Overview     string
	PosterPath   string
	SeasonNumber int
	VoteAverage  float32
	VoteCount    int64
	IsWatched    *bool
}

type Episode struct {
	ID            int64
	ReleaseDate   string
	EpisodeNumber int
	Title         string
	Overview      string
	Runtime       int
	SeasonNumber  int
	TVShowId      int64
	PosterPath    string
	VoteAverage   float32
	VoteCount     int64
	IsWatched     *bool
}
