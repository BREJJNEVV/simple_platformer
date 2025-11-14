package entities

// Bullet представляет пулю, выпущенную персонажем
type Bullet struct {
	X, Y          float64 // Позиция пули на экране
	VelocityX     float64 // Скорость пули по горизонтали (положительная = вправо, отрицательная = влево)
	Width, Height float64 // Размеры пули
}

// NewBullet создает новую пулю
func NewBullet(x, y, velocityX, width, height float64) *Bullet {
	return &Bullet{
		X:         x,
		Y:         y,
		VelocityX: velocityX,
		Width:     width,
		Height:    height,
	}
}

// Update обновляет позицию пули
func (b *Bullet) Update() {
	b.X += b.VelocityX
}
