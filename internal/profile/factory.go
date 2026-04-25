package profile

func Factory() []Profile {
	return []Profile{
		{ID: "factory:0.96", Name: "0.96\" Display", Factory: true, Width: 80, Height: 160, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
		{ID: "factory:1.05", Name: "1.05\" Display", Factory: true, Width: 120, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
		{ID: "factory:1.14", Name: "1.14\" Display", Factory: true, Width: 135, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
		{ID: "factory:1.90", Name: "1.90\" Display", Factory: true, Width: 120, Height: 240, FPS: 20, Quality: 9, Saturation: 2.5, Gamma: 0.8, Scaler: "lanczos"},
	}
}
