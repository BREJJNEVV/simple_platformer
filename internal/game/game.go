package game

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"platformer/internal/config"
	"platformer/internal/entities"
	"platformer/internal/network"
	"platformer/internal/physics"
	"platformer/internal/renderer"
)

// Camera представляет камеру, которая следует за игроком
type Camera struct {
	X, Y float64 // Позиция камеры в игровом мире
}

// Mode определяет режим игры.
type Mode string

const (
	ModeLocal  Mode = "local"
	ModeHost   Mode = "host"
	ModeClient Mode = "client"
)

// Options описывает параметры запуска игры.
type Options struct {
	Mode    Mode
	Address string
}

// Update обновляет позицию камеры, чтобы она следовала за игроком
func (c *Camera) Update(playerX, playerY float64) {
	// Центрируем камеру на игроке
	// Камера должна показывать игрока в центре экрана (или немного смещена вперед)
	targetX := playerX - config.ScreenWidth/2 + config.PlayerWidth/2

	// Ограничиваем камеру границами мира
	// Камера не должна выходить за левую границу мира
	if targetX < 0 {
		targetX = 0
	}
	// Камера не должна выходить за правую границу мира
	if targetX > config.WorldWidth-config.ScreenWidth {
		targetX = config.WorldWidth - config.ScreenWidth
	}

	// Плавно перемещаем камеру к целевой позиции
	// Это создает более плавное движение камеры
	c.X += (targetX - c.X) * 0.1

	// Камера по Y всегда центрирована на игроке (или можно сделать фиксированной)
	c.Y = playerY - config.ScreenHeight/2 + config.PlayerHeight/2
}

// Game представляет основное состояние игры
type Game struct {
	player    *entities.Player     // Игровой персонаж
	platforms []*entities.Platform // Список всех платформ на уровне (пустой, но оставляем для совместимости)
	bullets   []*entities.Bullet   // Список всех активных пуль на экране
	npcs      []*entities.NPC      // Список всех NPC на карте
	camera    Camera               // Камера, следующая за игроком
	remote    *entities.Player     // Удаленный игрок
	enemyFire []*entities.Bullet   // Пули удаленного игрока
	net       *network.Manager     // Менеджер сетевого подключения
	options   Options              // Опции запуска

	// Отслеживание состояния клавиш для одноразовых нажатий
	// Храним предыдущее состояние клавиш стрельбы
	prevShootKeyPressed bool // Предыдущее состояние клавиши стрельбы
}

// NewGame создает новую игру с начальными параметрами
func NewGame() *Game {
	gameInstance, err := NewGameWithOptions(Options{Mode: ModeLocal})
	if err != nil {
		panic(err)
	}
	return gameInstance
}

// NewGameWithOptions создает новую игру с заданными опциями.
func NewGameWithOptions(opts Options) (*Game, error) {
	// Создаем персонажа в начальной позиции
	player := entities.NewPlayer(100, 100)

	// Создаем пустую карту (все платформы убраны)
	platforms := createLevel()

	// Создаем NPC на карте
	npcs := []*entities.NPC{
		entities.NewNPC(500, config.WorldHeight-100, 40, 40), // NPC в центре карты
		entities.NewNPC(600, config.WorldHeight-100, 40, 40), // NPC дальше
		entities.NewNPC(650, config.WorldHeight-100, 40, 40), // NPC еще дальше
	}

	gameInstance := &Game{
		player:              player,
		platforms:           platforms,
		bullets:             make([]*entities.Bullet, 0), // Инициализируем пустой список пуль
		npcs:                npcs,                        // Добавляем NPC
		camera:              Camera{X: 0, Y: 0},          // Инициализируем камеру
		prevShootKeyPressed: false,                       // Инициализируем состояние клавиши стрельбы
		enemyFire:           make([]*entities.Bullet, 0),
		options:             opts,
	}

	if opts.Mode != ModeLocal {
		manager, err := startNetwork(opts)
		if err != nil {
			return nil, err
		}
		if manager != nil {
			gameInstance.net = manager
			gameInstance.remote = entities.NewPlayer(player.X, player.Y)
		}
	}

	return gameInstance, nil
}

func startNetwork(opts Options) (*network.Manager, error) {
	switch opts.Mode {
	case ModeLocal, Mode(""):
		return nil, nil
	case ModeHost:
		return network.Host(opts.Address)
	case ModeClient:
		return network.Join(opts.Address)
	default:
		return nil, fmt.Errorf("unknown game mode: %s", opts.Mode)
	}
}

// createLevel создает пустую карту без платформ
func createLevel() []*entities.Platform {
	// Возвращаем пустой список платформ
	// Оставляем только пол на всю ширину мира для того, чтобы персонаж не падал в бесконечность
	platforms := make([]*entities.Platform, 0)
	platforms = append(platforms, entities.NewPlatform(0, config.WorldHeight-60, config.WorldWidth, 1000))
	return platforms
}

// Update обновляет логику игры каждый кадр
func (g *Game) Update() error {
	// Обрабатываем ввод с клавиатуры
	g.handleInput()

	// Применяем гравитацию к персонажу
	g.applyGravity()

	// Обновляем позицию персонажа на основе скорости
	g.updatePlayerPosition()

	// Проверяем коллизии с платформами
	g.checkCollisions()

	// Обновляем все пули
	g.updateBullets()

	// Обновляем камеру, чтобы она следовала за игроком
	g.camera.Update(g.player.X, g.player.Y)

	// Синхронизируем состояние с удаленным игроком
	if err := g.updateNetwork(); err != nil {
		return err
	}

	return nil
}

// handleInput обрабатывает нажатия клавиш и управляет персонажем
func (g *Game) handleInput() {
	player := g.player

	// Проверяем нажатие клавиш движения влево/вправо
	// ebiten.IsKeyPressed проверяет, нажата ли клавиша в данный момент
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		// Движение влево - уменьшаем скорость по X
		player.VelocityX = -config.MoveSpeed
		player.FacingRight = false // Персонаж смотрит влево
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		// Движение вправо - увеличиваем скорость по X
		player.VelocityX = config.MoveSpeed
		player.FacingRight = true // Персонаж смотрит вправо
	} else {
		// Если клавиши не нажаты, применяем трение для замедления
		player.VelocityX *= config.Friction
		// Если скорость стала очень маленькой, останавливаем персонажа
		if math.Abs(player.VelocityX) < 0.1 {
			player.VelocityX = 0
		}
	}

	// Проверяем нажатие клавиши прыжка (пробел или стрелка вверх)
	// Прыгать можно только если персонаж стоит на платформе
	if (ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW)) && player.OnGround {
		// Применяем силу прыжка (отрицательное значение, так как Y растет вниз)
		player.VelocityY = config.JumpStrength
		// Помечаем, что персонаж больше не на земле
		player.OnGround = false
	}

	// Проверяем нажатие клавиши стрельбы (J или Enter)
	// Отслеживаем одноразовое нажатие, чтобы предотвратить непрерывную стрельбу
	// Проверяем, нажата ли клавиша сейчас
	shootKeyPressed := ebiten.IsKeyPressed(ebiten.KeyJ) || ebiten.IsKeyPressed(ebiten.KeyEnter)

	// Если клавиша нажата сейчас, но не была нажата в предыдущем кадре,
	// значит это новое нажатие - стреляем
	if shootKeyPressed && !g.prevShootKeyPressed {
		g.shoot() // Вызываем функцию стрельбы
	}

	// Сохраняем текущее состояние клавиши для следующего кадра
	g.prevShootKeyPressed = shootKeyPressed
}

// applyGravity применяет гравитацию к персонажу
func (g *Game) applyGravity() {
	player := g.player

	// Если персонаж не на земле, применяем гравитацию
	if !player.OnGround {
		// Увеличиваем скорость падения
		player.VelocityY += config.Gravity

		// Ограничиваем максимальную скорость падения
		// Это предотвращает слишком быстрое падение
		if player.VelocityY > config.MaxFallSpeed {
			player.VelocityY = config.MaxFallSpeed
		}
	}
}

// updatePlayerPosition обновляет позицию персонажа на основе его скорости
func (g *Game) updatePlayerPosition() {
	player := g.player

	// Обновляем позицию по X (горизонтальное движение)
	player.X += player.VelocityX

	// Обновляем позицию по Y (вертикальное движение)
	player.Y += player.VelocityY

	// Предотвращаем выход персонажа за границы мира по горизонтали
	if player.X < 0 {
		player.X = 0
		player.VelocityX = 0
	} else if player.X+config.PlayerWidth > config.WorldWidth {
		player.X = config.WorldWidth - config.PlayerWidth
		player.VelocityX = 0
	}

	// Если персонаж упал за нижнюю границу экрана, возвращаем его наверх
	if player.Y > config.ScreenHeight {
		player.Y = 100
		player.X = 100
		player.VelocityY = 0
		player.VelocityX = 0
	}
}

// checkCollisions проверяет столкновения персонажа с платформами
func (g *Game) checkCollisions() {
	player := g.player
	player.OnGround = false // Предполагаем, что персонаж не на земле

	// Проверяем каждую платформу
	for _, platform := range g.platforms {
		// Проверяем, пересекается ли персонаж с платформой
		if physics.IsColliding(player, platform, config.PlayerWidth, config.PlayerHeight) {
			// Вычисляем, с какой стороны произошло столкновение
			// Это нужно для правильной обработки коллизий

			// Вычисляем центр персонажа и платформы
			playerCenterX := player.X + config.PlayerWidth/2
			playerCenterY := player.Y + config.PlayerHeight/2
			platformCenterX := platform.X + platform.Width/2
			platformCenterY := platform.Y + platform.Height/2

			// Вычисляем расстояния между центрами
			dx := playerCenterX - platformCenterX
			dy := playerCenterY - platformCenterY

			// Вычисляем минимальное расстояние для разделения
			minDistX := (config.PlayerWidth + platform.Width) / 2
			minDistY := (config.PlayerHeight + platform.Height) / 2

			// Определяем, с какой стороны произошло столкновение
			overlapX := minDistX - math.Abs(dx)
			overlapY := minDistY - math.Abs(dy)

			// Если перекрытие по Y меньше, чем по X, значит столкновение вертикальное
			if overlapY < overlapX {
				// Вертикальное столкновение
				if dy < 0 {
					// Персонаж сверху платформы - ставим его на платформу
					player.Y = platform.Y - config.PlayerHeight
					player.VelocityY = 0
					player.OnGround = true
				} else {
					// Персонаж снизу платформы - останавливаем движение вверх
					player.Y = platform.Y + platform.Height
					player.VelocityY = 0
				}
			} else {
				// Горизонтальное столкновение
				if dx < 0 {
					// Персонаж слева от платформы
					player.X = platform.X - config.PlayerWidth
					player.VelocityX = 0
				} else {
					// Персонаж справа от платформы
					player.X = platform.X + platform.Width
					player.VelocityX = 0
				}
			}
		}
	}
}

// shoot создает новую пулю и добавляет ее в список пуль
func (g *Game) shoot() {
	player := g.player

	// Вычисляем начальную позицию пули
	// Пуля появляется в центре персонажа по вертикали
	// И с края персонажа по горизонтали (в зависимости от направления взгляда)
	var bulletX float64
	bulletY := player.Y + config.PlayerHeight/2 - config.BulletHeight/2

	// Если персонаж смотрит вправо, пуля появляется справа от персонажа
	if player.FacingRight {
		bulletX = player.X + config.PlayerWidth
	} else {
		// Если персонаж смотрит влево, пуля появляется слева от персонажа
		bulletX = player.X - config.BulletWidth
	}

	// Определяем направление скорости пули
	velocityX := config.BulletSpeed
	if !player.FacingRight {
		velocityX = -config.BulletSpeed
	}

	// Создаем новую пулю
	bullet := entities.NewBullet(bulletX, bulletY, velocityX, config.BulletWidth, config.BulletHeight)

	// Добавляем пулю в список активных пуль
	g.bullets = append(g.bullets, bullet)
}

// updateBullets обновляет позиции всех пуль и удаляет те, что вышли за границы экрана
func (g *Game) updateBullets() {
	// Создаем новый список для хранения активных пуль
	activeBullets := make([]*entities.Bullet, 0)

	// Проходим по всем пулям
	for _, bullet := range g.bullets {
		// Обновляем позицию пули на основе ее скорости
		bullet.Update()

		// Проверяем, не вышла ли пуля за границы мира
		// Если пуля еще в мире, добавляем ее в список активных
		if bullet.X > -config.BulletWidth && bullet.X < config.WorldWidth+config.BulletWidth {
			// Проверяем коллизии пули с платформами
			hitPlatform := false
			for _, platform := range g.platforms {
				if physics.IsBulletColliding(bullet, platform) {
					// Если пуля попала в платформу, помечаем ее для удаления
					hitPlatform = true
					break
				}
			}

			// Если пуля не попала в платформу, оставляем ее активной
			if !hitPlatform {
				activeBullets = append(activeBullets, bullet)
			}
		}
		// Если пуля вышла за границы экрана или попала в платформу, она не добавляется в activeBullets
		// и таким образом удаляется из игры
	}

	// Заменяем старый список пуль на новый (без удаленных пуль)
	g.bullets = activeBullets
}

// updateNetwork синхронизирует состояние игры между игроками.
func (g *Game) updateNetwork() error {
	if g.net == nil {
		return nil
	}

	if state, ok := g.net.LatestState(); ok {
		g.applyRemoteState(state)
	}

	if err := g.net.Send(g.buildLocalState()); err != nil {
		return err
	}

	if err := g.net.Err(); err != nil {
		return err
	}

	return nil
}

func (g *Game) buildLocalState() network.StateMessage {
	player := g.player

	msg := network.StateMessage{
		Player: network.PlayerState{
			X:           player.X,
			Y:           player.Y,
			VelocityX:   player.VelocityX,
			VelocityY:   player.VelocityY,
			OnGround:    player.OnGround,
			FacingRight: player.FacingRight,
		},
		Bullets: make([]network.BulletState, 0, len(g.bullets)),
	}

	for _, bullet := range g.bullets {
		msg.Bullets = append(msg.Bullets, network.BulletState{
			X:         bullet.X,
			Y:         bullet.Y,
			VelocityX: bullet.VelocityX,
		})
	}

	return msg
}

func (g *Game) applyRemoteState(state network.StateMessage) {
	if g.remote == nil {
		g.remote = entities.NewPlayer(state.Player.X, state.Player.Y)
	}

	g.remote.X = state.Player.X
	g.remote.Y = state.Player.Y
	g.remote.VelocityX = state.Player.VelocityX
	g.remote.VelocityY = state.Player.VelocityY
	g.remote.OnGround = state.Player.OnGround
	g.remote.FacingRight = state.Player.FacingRight

	if g.enemyFire == nil {
		g.enemyFire = make([]*entities.Bullet, 0, len(state.Bullets))
	} else {
		g.enemyFire = g.enemyFire[:0]
	}

	for _, bullet := range state.Bullets {
		g.enemyFire = append(g.enemyFire, entities.NewBullet(
			bullet.X,
			bullet.Y,
			bullet.VelocityX,
			config.BulletWidth,
			config.BulletHeight,
		))
	}
}

// Draw отрисовывает все объекты игры на экране
func (g *Game) Draw(screen *ebiten.Image) {
	// Очищаем экран, заливая его цветом неба
	screen.Fill(color.RGBA{R: 135, G: 206, B: 235, A: 255}) // Светло-голубой цвет

	// Рисуем все платформы с учетом позиции камеры
	for _, platform := range g.platforms {
		// Проверяем, видна ли платформа на экране (оптимизация отрисовки)
		if platform.X+platform.Width > g.camera.X && platform.X < g.camera.X+config.ScreenWidth {
			renderer.DrawPlatformWithCamera(screen, platform, g.camera.X, g.camera.Y)
		}
	}

	// Рисуем удаленного игрока и его пули, если он подключен
	if g.remote != nil {
		if g.remote.X+config.PlayerWidth > g.camera.X && g.remote.X < g.camera.X+config.ScreenWidth {
			renderer.DrawPlayerWithCamera(screen, g.remote, g.camera.X, g.camera.Y)
		}
		for _, bullet := range g.enemyFire {
			if bullet.X+bullet.Width > g.camera.X && bullet.X < g.camera.X+config.ScreenWidth {
				renderer.DrawBulletWithCamera(screen, bullet, g.camera.X, g.camera.Y)
			}
		}
	}

	// Рисуем персонажа с учетом позиции камеры
	renderer.DrawPlayerWithCamera(screen, g.player, g.camera.X, g.camera.Y)

	// Рисуем все пули с учетом позиции камеры
	for _, bullet := range g.bullets {
		// Проверяем, видна ли пуля на экране (оптимизация отрисовки)
		if bullet.X+bullet.Width > g.camera.X && bullet.X < g.camera.X+config.ScreenWidth {
			renderer.DrawBulletWithCamera(screen, bullet, g.camera.X, g.camera.Y)
		}
	}

	// Рисуем всех NPC с учетом позиции камеры
	for _, npc := range g.npcs {
		// Проверяем, виден ли NPC на экране (оптимизация отрисовки)
		if npc.X+npc.Width > g.camera.X && npc.X < g.camera.X+config.ScreenWidth {
			renderer.DrawNPCWithCamera(screen, npc, g.camera.X, g.camera.Y)
		}
	}

	// Выводим отладочную информацию
	renderer.DrawDebugInfo(screen, g.player, len(g.bullets))
}

// Layout возвращает размеры игрового экрана
// Эта функция требуется интерфейсом ebiten.Game
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}
