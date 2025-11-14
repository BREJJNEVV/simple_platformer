package entities

// NPC представляет неигрового персонажа
type NPC struct {
	// Позиция NPC на экране
	X, Y float64

	// Размеры NPC
	Width, Height float64

	// Направление взгляда NPC
	// true = смотрит вправо, false = смотрит влево
	FacingRight bool
}

// NewNPC создает нового NPC с заданными параметрами
func NewNPC(x, y, width, height float64) *NPC {
	return &NPC{
		X:           x,
		Y:           y,
		Width:       width,
		Height:      height,
		FacingRight: true, // По умолчанию смотрит вправо
	}
}
