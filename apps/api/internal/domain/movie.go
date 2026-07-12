package domain

// type Movie struct {
// 	ID          int64
// 	Title       string
// 	Overview    string
// 	ReleaseDate string
// 	PosterPath  string
// 	Popularity  float32
// 	VoteAverage float32
// }

// type UserMovie struct {
// 	MovieID int64
// 	Movie   *Movie
// 	Status  WatchStatus
// 	Rating  *int
// }

type MovieDetail struct {
	ID                  int64
	Title               string
	Overview            string
	ReleaseDate         string
	PosterPath          string
	UserRating          *int
	UserStatus          WatchStatus
	Genres              []Genre
	OriginalLanguage    string
	OriginCountry       []string
	ProductionCountries []string
	OriginalTitle       string
	Popularity          float32
	Status              string
	Videos              []Video
	VoteAverage         float32
	VoteCount           int64
	Runtime             int
	Budget              int64
	Revenue             int64
	CollectionID        *int64
}
