package profile

import "fmt"

type Profile struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Factory    bool    `json:"factory"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	FPS        int     `json:"fps"`
	Quality    int     `json:"quality"`
	Saturation float64 `json:"saturation"`
	Gamma      float64 `json:"gamma"`
	Scaler     string  `json:"scaler"`
}

func (p Profile) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("name must not be empty")
	}
	if p.Width <= 0 {
		return fmt.Errorf("width must be > 0")
	}
	if p.Height <= 0 {
		return fmt.Errorf("height must be > 0")
	}
	if p.FPS <= 0 {
		return fmt.Errorf("fps must be > 0")
	}
	if p.Quality < 1 || p.Quality > 31 {
		return fmt.Errorf("quality must be in [1,31]")
	}
	if p.Scaler == "" {
		return fmt.Errorf("scaler must not be empty")
	}
	return nil
}
