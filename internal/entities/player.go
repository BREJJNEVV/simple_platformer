package entities

// Player представляет игрового персонажа
type Player struct {
	// Позиция персонажа на экране
	X, Y float64

	// Скорость персонажа (для физики)
	VelocityX, VelocityY float64

	// Состояние персонажа
	OnGround bool // Находится ли персонаж на платформе

	// Направление взгляда персонажа (для стрельбы)
	// true = смотрит вправо, false = смотрит влево
	FacingRight bool
}

// NewPlayer создает нового персонажа с начальными параметрами
func NewPlayer(x, y float64) *Player {
	return &Player{
		X:           x,
		Y:           y,
		FacingRight: true, // По умолчанию персонаж смотрит вправо
	}
}
