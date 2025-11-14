package renderer

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"platformer/internal/config"
	"platformer/internal/entities"
)

var (
	playerSprite *ebiten.Image // Кэшированный спрайт персонажа
	npcSprite    *ebiten.Image // Кэшированный спрайт NPC
)

// init инициализирует спрайты при загрузке пакета
func init() {
	// Создаем спрайт персонажа (простой пиксельный арт)
	playerSprite = createPlayerSprite()
	// Создаем спрайт NPC
	npcSprite = createNPCSprite()
}

// createPlayerSprite создает простой спрайт персонажа программно
func createPlayerSprite() *ebiten.Image {
	img := ebiten.NewImage(config.PlayerWidth, config.PlayerHeight)

	// Рисуем простой спрайт персонажа
	// Голова (верхняя часть)
	headColor := color.RGBA{R: 255, G: 200, B: 150, A: 255} // Цвет кожи
	for y := 0; y < 12; y++ {
		for x := 8; x < 32; x++ {
			img.Set(x, y, headColor)
		}
	}

	// Глаза
	eyeColor := color.RGBA{R: 0, G: 0, B: 0, A: 255} // Черный
	img.Set(14, 6, eyeColor)
	img.Set(25, 6, eyeColor)

	// Тело (средняя часть)
	bodyColor := color.RGBA{R: 0, G: 100, B: 255, A: 255} // Синий
	for y := 12; y < 28; y++ {
		for x := 6; x < 34; x++ {
			img.Set(x, y, bodyColor)
		}
	}

	// Руки
	armColor := color.RGBA{R: 255, G: 200, B: 150, A: 255} // Цвет кожи
	for y := 14; y < 26; y++ {
		img.Set(4, y, armColor)
		img.Set(5, y, armColor)
		img.Set(34, y, armColor)
		img.Set(35, y, armColor)
	}

	// Ноги (нижняя часть)
	legColor := color.RGBA{R: 100, G: 50, B: 0, A: 255} // Коричневый
	for y := 28; y < 40; y++ {
		for x := 10; x < 18; x++ {
			img.Set(x, y, legColor)
		}
		for x := 22; x < 30; x++ {
			img.Set(x, y, legColor)
		}
	}

	return img
}

// createNPCSprite создает простой спрайт NPC программно
func createNPCSprite() *ebiten.Image {
	img := ebiten.NewImage(40, 40)

	// Рисуем простой спрайт NPC (зеленый персонаж)
	// Голова
	headColor := color.RGBA{R: 150, G: 255, B: 150, A: 255} // Светло-зеленый
	for y := 0; y < 12; y++ {
		for x := 8; x < 32; x++ {
			img.Set(x, y, headColor)
		}
	}

	// Глаза
	eyeColor := color.RGBA{R: 0, G: 0, B: 0, A: 255} // Черный
	img.Set(14, 6, eyeColor)
	img.Set(25, 6, eyeColor)

	// Тело
	bodyColor := color.RGBA{R: 0, G: 200, B: 0, A: 255} // Зеленый
	for y := 12; y < 28; y++ {
		for x := 6; x < 34; x++ {
			img.Set(x, y, bodyColor)
		}
	}

	// Руки
	armColor := color.RGBA{R: 150, G: 255, B: 150, A: 255} // Светло-зеленый
	for y := 14; y < 26; y++ {
		img.Set(4, y, armColor)
		img.Set(5, y, armColor)
		img.Set(34, y, armColor)
		img.Set(35, y, armColor)
	}

	// Ноги
	legColor := color.RGBA{R: 0, G: 150, B: 0, A: 255} // Темно-зеленый
	for y := 28; y < 40; y++ {
		for x := 10; x < 18; x++ {
			img.Set(x, y, legColor)
		}
		for x := 22; x < 30; x++ {
			img.Set(x, y, legColor)
		}
	}

	return img
}

// DrawPlayer рисует персонажа на экране
func DrawPlayer(screen *ebiten.Image, player *entities.Player) {
	// Создаем изображение для персонажа
	playerImg := ebiten.NewImage(config.PlayerWidth, config.PlayerHeight)

	// Заливаем персонажа цветом (красный квадрат)
	playerImg.Fill(color.RGBA{R: 255, G: 0, B: 0, A: 255})

	// Создаем опции для позиционирования
	op := &ebiten.DrawImageOptions{}

	// Устанавливаем позицию, где нужно нарисовать персонажа
	op.GeoM.Translate(player.X, player.Y)

	// Рисуем персонажа на экране
	screen.DrawImage(playerImg, op)
}

// DrawPlayerWithCamera рисует персонажа на экране с учетом позиции камеры
func DrawPlayerWithCamera(screen *ebiten.Image, player *entities.Player, cameraX, cameraY float64) {
	// Используем предзагруженный спрайт персонажа
	if playerSprite == nil {
		// Если спрайт не загружен, создаем его
		playerSprite = createPlayerSprite()
	}

	// Создаем опции для позиционирования
	op := &ebiten.DrawImageOptions{}

	// Если персонаж смотрит влево, отражаем спрайт по горизонтали
	if !player.FacingRight {
		op.GeoM.Scale(-1, 1)                     // Отражаем по горизонтали
		op.GeoM.Translate(config.PlayerWidth, 0) // Смещаем после отражения
	}

	// Вычисляем позицию на экране с учетом камеры
	// Вычитаем позицию камеры, чтобы объект отображался в правильном месте на экране
	screenX := player.X - cameraX
	screenY := player.Y - cameraY

	// Устанавливаем позицию, где нужно нарисовать персонажа
	op.GeoM.Translate(screenX, screenY)

	// Рисуем спрайт персонажа на экране
	screen.DrawImage(playerSprite, op)
}

// DrawPlatform рисует платформу на экране
func DrawPlatform(screen *ebiten.Image, platform *entities.Platform) {
	// Создаем изображение для платформы
	platformImg := ebiten.NewImage(int(platform.Width), int(platform.Height))

	// Заливаем платформу коричневым цветом
	platformImg.Fill(color.RGBA{R: 0, G: 255, B: 0, A: 255})

	// Создаем опции для позиционирования
	op := &ebiten.DrawImageOptions{}

	// Устанавливаем позицию платформы
	op.GeoM.Translate(platform.X, platform.Y)

	// Рисуем платформу на экране
	screen.DrawImage(platformImg, op)
}

// DrawPlatformWithCamera рисует платформу на экране с учетом позиции камеры
func DrawPlatformWithCamera(screen *ebiten.Image, platform *entities.Platform, cameraX, cameraY float64) {
	// Создаем изображение для платформы
	platformImg := ebiten.NewImage(int(platform.Width), int(platform.Height))

	// Заливаем платформу коричневым цветом
	platformImg.Fill(color.RGBA{R: 139, G: 69, B: 19, A: 255})

	// Создаем опции для позиционирования
	op := &ebiten.DrawImageOptions{}

	// Вычисляем позицию на экране с учетом камеры
	// Вычитаем позицию камеры, чтобы объект отображался в правильном месте на экране
	screenX := platform.X - cameraX
	screenY := platform.Y - cameraY

	// Устанавливаем позицию платформы
	op.GeoM.Translate(screenX, screenY)

	// Рисуем платформу на экране
	screen.DrawImage(platformImg, op)
}

// DrawBullet рисует пулю на экране
func DrawBullet(screen *ebiten.Image, bullet *entities.Bullet) {
	// Создаем изображение для пули
	bulletImg := ebiten.NewImage(int(bullet.Width), int(bullet.Height))

	// Заливаем пулю желтым цветом для лучшей видимости
	bulletImg.Fill(color.RGBA{R: 255, G: 255, B: 0, A: 255})

	// Создаем опции для позиционирования
	op := &ebiten.DrawImageOptions{}

	// Устанавливаем позицию пули
	op.GeoM.Translate(bullet.X, bullet.Y)

	// Рисуем пулю на экране
	screen.DrawImage(bulletImg, op)
}

// DrawBulletWithCamera рисует пулю на экране с учетом позиции камеры
func DrawBulletWithCamera(screen *ebiten.Image, bullet *entities.Bullet, cameraX, cameraY float64) {
	// Создаем изображение для пули
	bulletImg := ebiten.NewImage(int(bullet.Width), int(bullet.Height))

	// Заливаем пулю желтым цветом для лучшей видимости
	bulletImg.Fill(color.RGBA{R: 255, G: 255, B: 0, A: 255})

	// Создаем опции для позиционирования
	op := &ebiten.DrawImageOptions{}

	// Вычисляем позицию на экране с учетом камеры
	// Вычитаем позицию камеры, чтобы объект отображался в правильном месте на экране
	screenX := bullet.X - cameraX
	screenY := bullet.Y - cameraY

	// Устанавливаем позицию пули
	op.GeoM.Translate(screenX, screenY)

	// Рисуем пулю на экране
	screen.DrawImage(bulletImg, op)
}

// DrawDebugInfo выводит отладочную информацию на экран
func DrawDebugInfo(screen *ebiten.Image, player *entities.Player, bulletCount int) {
	// Выводим информацию для отладки (FPS, позиция персонажа)
	ebitenutil.DebugPrint(screen, "Платформер на Go!")
	ebitenutil.DebugPrintAt(screen,
		"Управление: Стрелки/WASD - движение, Пробел - прыжок, J/Enter - стрельба",
		0, 20)
	ebitenutil.DebugPrintAt(screen,
		"Позиция: X="+formatFloat(player.X)+" Y="+formatFloat(player.Y),
		0, 40)
	ebitenutil.DebugPrintAt(screen,
		"Скорость: VX="+formatFloat(player.VelocityX)+" VY="+formatFloat(player.VelocityY),
		0, 60)
	if player.OnGround {
		ebitenutil.DebugPrintAt(screen, "На земле: Да", 0, 80)
	} else {
		ebitenutil.DebugPrintAt(screen, "На земле: Нет", 0, 80)
	}
	// Выводим количество активных пуль
	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("Пули: %d", bulletCount),
		0, 100)
}

// DrawNPCWithCamera рисует NPC на экране с учетом позиции камеры
func DrawNPCWithCamera(screen *ebiten.Image, npc *entities.NPC, cameraX, cameraY float64) {
	// Используем предзагруженный спрайт NPC
	if npcSprite == nil {
		// Если спрайт не загружен, создаем его
		npcSprite = createNPCSprite()
	}

	// Создаем опции для позиционирования
	op := &ebiten.DrawImageOptions{}

	// Если NPC смотрит влево, отражаем спрайт по горизонтали
	if !npc.FacingRight {
		op.GeoM.Scale(-1, 1)            // Отражаем по горизонтали
		op.GeoM.Translate(npc.Width, 0) // Смещаем после отражения
	}

	// Вычисляем позицию на экране с учетом камеры
	// Вычитаем позицию камеры, чтобы объект отображался в правильном месте на экране
	screenX := npc.X - cameraX
	screenY := npc.Y - cameraY

	// Устанавливаем позицию NPC
	op.GeoM.Translate(screenX, screenY)

	// Рисуем спрайт NPC на экране
	screen.DrawImage(npcSprite, op)
}

// formatFloat форматирует число с плавающей точкой для вывода
func formatFloat(f float64) string {
	return fmt.Sprintf("%.1f", f)
}
