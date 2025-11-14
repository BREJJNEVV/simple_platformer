package physics

import "platformer/internal/entities"

// IsColliding проверяет, пересекается ли персонаж с платформой
// Используется алгоритм AABB (Axis-Aligned Bounding Box) для проверки коллизий
func IsColliding(player *entities.Player, platform *entities.Platform, playerWidth, playerHeight float64) bool {
	// Проверяем, не пересекаются ли прямоугольники
	// Два прямоугольника пересекаются, если:
	// - левая сторона одного не правее правой стороны другого
	// - правая сторона одного не левее левой стороны другого
	// - верхняя сторона одного не ниже нижней стороны другого
	// - нижняя сторона одного не выше верхней стороны другого

	return player.X < platform.X+platform.Width &&
		player.X+playerWidth > platform.X &&
		player.Y < platform.Y+platform.Height &&
		player.Y+playerHeight > platform.Y
}

// IsBulletColliding проверяет, пересекается ли пуля с платформой
func IsBulletColliding(bullet *entities.Bullet, platform *entities.Platform) bool {
	// Используем тот же алгоритм AABB, что и для персонажа
	return bullet.X < platform.X+platform.Width &&
		bullet.X+bullet.Width > platform.X &&
		bullet.Y < platform.Y+platform.Height &&
		bullet.Y+bullet.Height > platform.Y
}
