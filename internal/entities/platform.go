package entities

// Platform представляет платформу в игре
type Platform struct {
	X, Y          float64 // Позиция платформы
	Width, Height float64 // Размеры платформы
}

// NewPlatform создает новую платформу
func NewPlatform(x, y, width, height float64) *Platform {
	return &Platform{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}
