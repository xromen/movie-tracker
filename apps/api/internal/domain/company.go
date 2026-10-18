package domain

type CompanyDetails struct {
	TMDBID        int64 `json:"ID"`
	Name          string
	Description   string
	LogoPath      string
	OriginCountry string
	Homepage      string
	Headquarters  string
}
