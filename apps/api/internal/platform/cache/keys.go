package cache

import "fmt"

// Конвенция ключей: {сущность}:{идентификатор}:{уточнение}
func MovieSearchKey(query string, page int) string {
	return fmt.Sprintf("movie:search:%s:%d", query, page)
}
func MovieNowPlayingKey(page int) string {
	return fmt.Sprintf("movie:nowplaying:%d", page)
}
func MoviePopularKey(page int) string {
	return fmt.Sprintf("movie:popular:%d", page)
}
func MovieTopRatedKey(page int) string {
	return fmt.Sprintf("movie:toprated:%d", page)
}
func MovieUpcomingKey(page int) string {
	return fmt.Sprintf("movie:upcoming:%d", page)
}
func MovieDetailKey(tmdbID int64) string {
	return fmt.Sprintf("movie:detail:tmdb:%d", tmdbID)
}
func MovieRecommendationsKey(tmdbID int64, page int) string {
	return fmt.Sprintf("movie:tmdb:%d:recommendations:%d", tmdbID, page)
}
func UserMoviesKey(userID int64, status string, page, perPage int) string {
	return fmt.Sprintf("user:%d:movies:%s:%d:%d", userID, status, page, perPage)
}
func UserMoviesTemplate(userID int64, status string) string {
	return fmt.Sprintf("user:%d:movies:%s:*", userID, status)
}
func MovieUserDetailKey(userID, tmdbID int64) string {
	return fmt.Sprintf("user:%d:moviedetail:tmdb:%d", userID, tmdbID)
}

func TVShowSearchKey(query string, page int) string {
	return fmt.Sprintf("tvshow:search:%s:%d", query, page)
}
func TVShowOnTheAirKey(page int) string {
	return fmt.Sprintf("tvshow:ontheair:%d", page)
}
func TVShowPopularKey(page int) string {
	return fmt.Sprintf("tvshow:popular:%d", page)
}
func TVShowTopRatedKey(page int) string {
	return fmt.Sprintf("tvshow:toprated:%d", page)
}
func TVShowAiringTodayKey(page int) string {
	return fmt.Sprintf("tvshow:airingtoday:%d", page)
}
func TVShowDetailKey(tmdbID int64) string {
	return fmt.Sprintf("tvshow:detail:tmdb:%d", tmdbID)
}
func TVShowRecommendationsKey(tmdbID int64, page int) string {
	return fmt.Sprintf("tvshow:tmdb:%d:recommendations:%d", tmdbID, page)
}
func UserTVShowsKey(userID int64, status string, page, perPage int) string {
	return fmt.Sprintf("user:%d:tvshow:%s:%d:%d", userID, status, page, perPage)
}
func UserTVShowsTemplate(userID int64, status string) string {
	return fmt.Sprintf("user:%d:tvshow:%s:*", userID, status)
}
func TVShowUserDetailKey(userID, tmdbID int64) string {
	return fmt.Sprintf("user:%d:tvshowdetail:tmdb:%d", userID, tmdbID)
}
func TVShowSeasonEpisodesKey(tvShowID int64, seasonNumber int, page int) string {
	return fmt.Sprintf("tvshow:%d:season:%d:episodes:%d", tvShowID, seasonNumber, page)
}

func UserMediaListKey(userID int64, status string, mediaType string, page, perPage int) string {
	return fmt.Sprintf("media:%s:%s:user:%d:%d:%d", mediaType, status, userID, page, perPage)
}
func UserMediaListTemplate(userID int64, status string, mediaType string) string {
	return fmt.Sprintf("media:%s:%s:user:%d:*", mediaType, status, userID)
}

func CollectionKey(id int64) string {
	return fmt.Sprintf("collection:%d", id)
}

func SearchMultiKey(query string, page int) string {
	return fmt.Sprintf("searchmulti:%s:%d", query, page)
}
