package tmdb

import (
	"context"
	"fmt"
	"net/http"
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"
	"golang.org/x/time/rate"

	"github.com/xromen/movietracker/internal/domain"
)

const (
	maxTmdbPagesCount = 500
)

type translationData struct {
	Title    string
	Overview string
}

type Paginated[T any] struct {
	Items      []T
	TotalPages int
	TotalItems int
}

type Client interface {
	//Movies
	SearchMovies(ctx context.Context, query string, page int) (*Paginated[domain.Media], error)
	GetMovie(ctx context.Context, tmdbId int64) (*domain.Media, error)
	GetMovieNowPlaying(ctx context.Context, page int) (*Paginated[domain.Media], error)
	GetMoviePopular(ctx context.Context, page int) (*Paginated[domain.Media], error)
	GetMovieTopRated(ctx context.Context, page int) (*Paginated[domain.Media], error)
	GetMovieUpcoming(ctx context.Context, page int) (*Paginated[domain.Media], error)
	GetMovieDetails(ctx context.Context, tmdbId int64) (*domain.MovieDetail, error)
	GetMovieRecommendations(ctx context.Context, tmdbId int64, page int) (*Paginated[domain.Media], error)

	//Collection
	GetCollectionDetails(ctx context.Context, id int64) (*domain.Collection, error)

	//TVShows
	SearchTVShows(ctx context.Context, query string, page int) (*Paginated[domain.Media], error)
	GetTVShow(ctx context.Context, tmdbId int64) (*domain.Media, error)
	GetTVShowAiringToday(ctx context.Context, page int) (*Paginated[domain.Media], error)
	GetTVShowOnTheAir(ctx context.Context, page int) (*Paginated[domain.Media], error)
	GetTVShowPopular(ctx context.Context, page int) (*Paginated[domain.Media], error)
	GetTVShowTopRated(ctx context.Context, page int) (*Paginated[domain.Media], error)
	GetTVShowDetails(ctx context.Context, tmdbId int64) (*domain.TVShowDetail, error)
	GetTVShowRecommendations(ctx context.Context, tmdbId int64, page int) (*Paginated[domain.Media], error)
	GetTvSeasonEpisodes(ctx context.Context, tvId int64, seasonNumber int, page int) (*Paginated[domain.Episode], error)

	//Search
	SearchMulti(ctx context.Context, query string, page int) (*Paginated[domain.Media], error)
}

type Config struct {
	BaseURL       string
	ImagesBaseURL string
	BearerToken   string
	Timeout       time.Duration
	RPM           int
	Burst         int
}

type client struct {
	bearerToken   string
	imagesBaseURL string
	timeout       time.Duration
	transport     http.RoundTripper
	RPM           int
	Burst         int
}

func NewClient(cfg Config) (Client, error) {
	c, err := gotmdb.InitV4(cfg.BearerToken)
	if err != nil {
		return nil, fmt.Errorf("init tmdb client: %w", err)
	}

	c.SetCustomBaseURL(cfg.BaseURL)

	return &client{
		bearerToken:   cfg.BearerToken,
		imagesBaseURL: cfg.ImagesBaseURL,
		timeout:       cfg.Timeout,
		transport:     http.DefaultTransport,
		RPM:           cfg.RPM,
		Burst:         cfg.Burst,
	}, nil
}

func (c *client) requestClient(ctx context.Context) (*gotmdb.Client, error) {
	tmdbClient, err := gotmdb.InitV4(c.bearerToken)
	if err != nil {
		return nil, fmt.Errorf("init tmdb request client: %w", err)
	}

	contextTransport := contextTransport{
		ctx:  ctx,
		base: c.transport,
	}

	limiter := rate.NewLimiter(rate.Limit(c.RPM), c.Burst)

	rateLimitedTransport := &rateLimitedTransport{
		base:    contextTransport,
		limiter: limiter,
		ctx:     ctx,
	}

	tmdbClient.SetClientConfig(http.Client{
		Timeout:   c.timeout,
		Transport: rateLimitedTransport,
	})

	return tmdbClient, nil
}

type contextTransport struct {
	ctx  context.Context
	base http.RoundTripper
}

type rateLimitedTransport struct {
	base    http.RoundTripper
	limiter *rate.Limiter
	ctx     context.Context
}

func (t contextTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := t.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}

	return base.RoundTrip(req.Clone(ctx))
}

func (t *rateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	err := t.limiter.Wait(t.ctx)
	if err != nil {
		return nil, fmt.Errorf("rate limit wait: %w", err)
	}

	return t.base.RoundTrip(req)
}

func (c *client) getPosterPath(relative string) string {
	if relative == "" {
		return ""
	}
	return fmt.Sprintf("%s/t/p/w500%s", c.imagesBaseURL, relative)
}

func getDefaultOpts() map[string]string {
	return map[string]string{
		"language":      "ru",
		"region":        "ru",
		"include_adult": "true",
	}
}

func toDomainGenres(genres []gotmdb.Genre) []domain.Genre {
	domainGenres := make([]domain.Genre, 0, len(genres))
	for _, genre := range genres {
		domainGenres = append(domainGenres, domain.Genre{
			ID:   genre.ID,
			Name: genre.Name,
		})
	}
	return domainGenres
}

func toDomainViedos(result *gotmdb.VideoResults) []domain.Video {
	var videos = make([]domain.Video, 0, len(result.Results))
	for _, result := range result.Results {
		videos = append(videos, domain.Video{
			ID:          result.ID,
			Iso639_1:    result.Iso639_1,
			Iso3166_1:   result.Iso3166_1,
			Name:        result.Name,
			Official:    result.Official,
			PublishedAt: result.PublishedAt,
			Site:        result.Site,
			Size:        result.Size,
			Type:        result.Type,
			Key:         result.Key,
		})
	}

	return videos
}
