package metatube

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"time"

	"github.com/metatube-community/metatube-sdk-go/engine"
	"github.com/metatube-community/metatube-sdk-go/engine/providerid"
	sdkmodel "github.com/metatube-community/metatube-sdk-go/model"
)

type Client struct {
	engine *engine.Engine
}

func New(eng *engine.Engine) *Client {
	return &Client{engine: eng}
}

type ProvidersResponse struct {
	ActorProviders map[string]string `json:"actor_providers"`
	MovieProviders map[string]string `json:"movie_providers"`
}

func (c *Client) GetProviders() (*ProvidersResponse, error) {
	movie := c.engine.GetMovieProviders()
	actor := c.engine.GetActorProviders()
	resp := &ProvidersResponse{
		MovieProviders: make(map[string]string, len(movie)),
		ActorProviders: make(map[string]string, len(actor)),
	}
	for _, p := range movie {
		if u := p.URL(); u != nil {
			resp.MovieProviders[p.Name()] = u.String()
		} else {
			resp.MovieProviders[p.Name()] = p.Name()
		}
	}
	for _, p := range actor {
		if u := p.URL(); u != nil {
			resp.ActorProviders[p.Name()] = u.String()
		} else {
			resp.ActorProviders[p.Name()] = p.Name()
		}
	}
	return resp, nil
}

func (c *Client) SearchMovies(query, provider string, fallback bool) ([]MovieSearchResult, error) {
	var sdks []*sdkmodel.MovieSearchResult
	var err error
	if provider == "" {
		sdks, err = c.engine.SearchMovieAll(query, fallback)
	} else {
		sdks, err = c.engine.SearchMovie(query, provider, fallback)
	}
	if err != nil {
		return nil, err
	}
	out := make([]MovieSearchResult, 0, len(sdks))
	for _, r := range sdks {
		out = append(out, convertMovieSearchResult(r))
	}
	return out, nil
}

func (c *Client) GetMovieInfo(provider, id string) (*MovieInfo, error) {
	info, err := c.engine.GetMovieInfoByProviderID(providerid.ProviderID{
		Provider: provider,
		ID:       id,
	}, false)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, fmt.Errorf("movie info not found: %s/%s", provider, id)
	}
	out := convertMovieInfo(info)
	return &out, nil
}

func (c *Client) SearchActors(query, provider string, fallback bool) ([]ActorSearchResult, error) {
	var sdks []*sdkmodel.ActorSearchResult
	var err error
	if provider == "" {
		sdks, err = c.engine.SearchActorAll(query, fallback)
	} else {
		sdks, err = c.engine.SearchActor(query, provider, fallback)
	}
	if err != nil {
		return nil, err
	}
	out := make([]ActorSearchResult, 0, len(sdks))
	for _, r := range sdks {
		out = append(out, convertActorSearchResult(r))
	}
	return out, nil
}

func (c *Client) GetActorInfo(provider, id string) (*ActorInfo, error) {
	info, err := c.engine.GetActorInfoByProviderID(providerid.ProviderID{
		Provider: provider,
		ID:       id,
	}, false)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, fmt.Errorf("actor info not found: %s/%s", provider, id)
	}
	out := convertActorInfo(info)
	return &out, nil
}

func (c *Client) HealthCheck() error {
	return nil
}

func (c *Client) GetImageThumb(provider, id string) ([]byte, error) {
	img, err := c.engine.GetMovieThumbImage(providerid.ProviderID{
		Provider: provider,
		ID:       id,
	})
	if err != nil || img == nil {
		return nil, err
	}
	return encodeJPEG(img)
}

// GetImageCover returns the uncropped cover, matching the original
// metatube-web-manager behaviour (its cover route served the raw thumb).
func (c *Client) GetImageCover(provider, id string) ([]byte, error) {
	img, err := c.engine.GetMovieBackdropImage(providerid.ProviderID{
		Provider: provider,
		ID:       id,
	})
	if err != nil || img == nil {
		return nil, err
	}
	return encodeJPEG(img)
}

// GetImagePoster returns the portrait 2:3 poster for grid cards. The -1/-1
// sentinels let the engine apply face-aware positioning when the number calls
// for it.
func (c *Client) GetImagePoster(provider, id string) ([]byte, error) {
	img, err := c.engine.GetMoviePrimaryImage(providerid.ProviderID{
		Provider: provider,
		ID:       id,
	}, -1, -1)
	if err != nil || img == nil {
		return nil, err
	}
	return encodeJPEG(img)
}

func (c *Client) GetImageBigThumb(provider, id string) ([]byte, error) {
	img, err := c.engine.GetMovieBackdropImage(providerid.ProviderID{
		Provider: provider,
		ID:       id,
	})
	if err != nil || img == nil {
		return nil, err
	}
	return encodeJPEG(img)
}

func (c *Client) GetImageBigCover(provider, id string) ([]byte, error) {
	img, err := c.engine.GetMovieBackdropImage(providerid.ProviderID{
		Provider: provider,
		ID:       id,
	})
	if err != nil || img == nil {
		return nil, err
	}
	return encodeJPEG(img)
}

func encodeJPEG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func pqToStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	out := make([]string, len(s))
	copy(out, s)
	return out
}

func dateToString(d time.Time) string {
	if d.IsZero() {
		return ""
	}
	return d.Format("2006-01-02")
}

func convertMovieSearchResult(r *sdkmodel.MovieSearchResult) MovieSearchResult {
	return MovieSearchResult{
		ID:          r.ID,
		Number:      r.Number,
		Title:       r.Title,
		Provider:    r.Provider,
		Homepage:    r.Homepage,
		ThumbURL:    r.ThumbURL,
		CoverURL:    r.CoverURL,
		Score:       r.Score,
		Actors:      pqToStrings(r.Actors),
		ReleaseDate: dateToString(time.Time(r.ReleaseDate)),
	}
}

func convertActorSearchResult(r *sdkmodel.ActorSearchResult) ActorSearchResult {
	return ActorSearchResult{
		ID:       r.ID,
		Name:     r.Name,
		Provider: r.Provider,
		Homepage: r.Homepage,
		Aliases:  pqToStrings(r.Aliases),
		Images:   pqToStrings(r.Images),
	}
}

type MovieSearchResult struct {
	ID          string   `json:"id"`
	Number      string   `json:"number"`
	Title       string   `json:"title"`
	Provider    string   `json:"provider"`
	Homepage    string   `json:"homepage"`
	ThumbURL    string   `json:"thumb_url"`
	CoverURL    string   `json:"cover_url"`
	Score       float64  `json:"score"`
	Actors      []string `json:"actors"`
	ReleaseDate string   `json:"release_date"`
}

type MovieInfo struct {
	ID                 string   `json:"id"`
	Number             string   `json:"number"`
	Title              string   `json:"title"`
	Summary            string   `json:"summary"`
	Provider           string   `json:"provider"`
	Homepage           string   `json:"homepage"`
	Director           string   `json:"director"`
	Actors             []string `json:"actors"`
	ThumbURL           string   `json:"thumb_url"`
	BigThumbURL        string   `json:"big_thumb_url"`
	CoverURL           string   `json:"cover_url"`
	BigCoverURL        string   `json:"big_cover_url"`
	PreviewVideoURL    string   `json:"preview_video_url"`
	PreviewVideoHLSURL string   `json:"preview_video_hls_url"`
	PreviewImages      []string `json:"preview_images"`
	Maker              string   `json:"maker"`
	Label              string   `json:"label"`
	Series             string   `json:"series"`
	Genres             []string `json:"genres"`
	Score              float64  `json:"score"`
	Runtime            int      `json:"runtime"`
	ReleaseDate        string   `json:"release_date"`
}

type ActorSearchResult struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Provider string   `json:"provider"`
	Homepage string   `json:"homepage"`
	Aliases  []string `json:"aliases"`
	Images   []string `json:"images"`
}

type ActorInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Provider     string   `json:"provider"`
	Homepage     string   `json:"homepage"`
	Summary      string   `json:"summary"`
	Hobby        string   `json:"hobby"`
	Skill        string   `json:"skill"`
	BloodType    string   `json:"blood_type"`
	CupSize      string   `json:"cup_size"`
	Measurements string   `json:"measurements"`
	Nationality  string   `json:"nationality"`
	Height       int      `json:"height"`
	Aliases      []string `json:"aliases"`
	Images       []string `json:"images"`
	Birthday     string   `json:"birthday"`
	DebutDate    string   `json:"debut_date"`
}

func convertMovieInfo(info *sdkmodel.MovieInfo) MovieInfo {
	return MovieInfo{
		ID:                 info.ID,
		Number:             info.Number,
		Title:              info.Title,
		Summary:            info.Summary,
		Provider:           info.Provider,
		Homepage:           info.Homepage,
		Director:           info.Director,
		Actors:             pqToStrings(info.Actors),
		ThumbURL:           info.ThumbURL,
		BigThumbURL:        info.BigThumbURL,
		CoverURL:           info.CoverURL,
		BigCoverURL:        info.BigCoverURL,
		PreviewVideoURL:    info.PreviewVideoURL,
		PreviewVideoHLSURL: info.PreviewVideoHLSURL,
		PreviewImages:      pqToStrings(info.PreviewImages),
		Maker:              info.Maker,
		Label:              info.Label,
		Series:             info.Series,
		Genres:             pqToStrings(info.Genres),
		Score:              info.Score,
		Runtime:            info.Runtime,
		ReleaseDate:        dateToString(time.Time(info.ReleaseDate)),
	}
}

func convertActorInfo(info *sdkmodel.ActorInfo) ActorInfo {
	return ActorInfo{
		ID:           info.ID,
		Name:         info.Name,
		Provider:     info.Provider,
		Homepage:     info.Homepage,
		Summary:      info.Summary,
		Hobby:        info.Hobby,
		Skill:        info.Skill,
		BloodType:    info.BloodType,
		CupSize:      info.CupSize,
		Measurements: info.Measurements,
		Nationality:  info.Nationality,
		Height:       info.Height,
		Aliases:      pqToStrings(info.Aliases),
		Images:       pqToStrings(info.Images),
		Birthday:     dateToString(time.Time(info.Birthday)),
		DebutDate:    dateToString(time.Time(info.DebutDate)),
	}
}
