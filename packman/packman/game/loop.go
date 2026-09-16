package game

import "time"

// GameLoop handles the timing and update cycle of the game
type GameLoop struct {
	lastUpdate      time.Time
	deltaTime       time.Duration
	tickRate        time.Duration
	accumulator     time.Duration
	isRunning       bool
	paused          bool
	pauseStartTime  time.Duration
	totalPauseTime  time.Duration
	startTime       time.Time
}

// NewGameLoop creates a new game loop with the specified tick rate
func NewGameLoop(ticksPerSecond int) *GameLoop {
	tickRate := time.Second / time.Duration(ticksPerSecond)
	return &GameLoop{
		tickRate:   tickRate,
		lastUpdate: time.Now(),
		isRunning:  false,
	}
}

// Start begins the game loop
func (gl *GameLoop) Start() {
	gl.isRunning = true
	gl.paused = false
	gl.startTime = time.Now()
	gl.totalPauseTime = 0
	gl.lastUpdate = time.Now()
}

// Stop ends the game loop
func (gl *GameLoop) Stop() {
	gl.isRunning = false
}

// Pause pauses the game loop
func (gl *GameLoop) Pause() {
	if gl.isRunning && !gl.paused {
		gl.paused = true
		gl.pauseStartTime = time.Since(gl.startTime) - gl.totalPauseTime
	}
}

// Resume resumes the game loop
func (gl *GameLoop) Resume() {
	if gl.isRunning && gl.paused {
		gl.paused = false
		gl.totalPauseTime += time.Since(gl.startTime) - gl.pauseStartTime - gl.totalPauseTime
		gl.lastUpdate = time.Now()
	}
}

// IsPaused returns whether the game is paused
func (gl *GameLoop) IsPaused() bool {
	return gl.paused
}

// IsRunning returns whether the game loop is running
func (gl *GameLoop) IsRunning() bool {
	return gl.isRunning
}

// GetDeltaTime returns the time elapsed since the last update
func (gl *GameLoop) GetDeltaTime() time.Duration {
	return gl.deltaTime
}

// GetDeltaSeconds returns the delta time in seconds as a float64
func (gl *GameLoop) GetDeltaSeconds() float64 {
	return gl.deltaTime.Seconds()
}

// GetElapsedTime returns the total elapsed game time (excluding pauses)
func (gl *GameLoop) GetElapsedTime() time.Duration {
	if !gl.isRunning {
		return 0
	}
	if gl.paused {
		return gl.pauseStartTime
	}
	return time.Since(gl.startTime) - gl.totalPauseTime
}

// Update updates the game loop timing
func (gl *GameLoop) Update() {
	if !gl.isRunning || gl.paused {
		return
	}

	now := time.Now()
	gl.deltaTime = now.Sub(gl.lastUpdate)
	gl.lastUpdate = now
}

// ShouldUpdate returns true if enough time has passed for a game update
func (gl *GameLoop) ShouldUpdate() bool {
	if !gl.isRunning || gl.paused {
		return false
	}

	now := time.Now()
	elapsed := now.Sub(gl.lastUpdate)
	return elapsed >= gl.tickRate
}

// GetTickRate returns the tick rate of the game loop
func (gl *GameLoop) GetTickRate() time.Duration {
	return gl.tickRate
}

// GetTickRateSeconds returns the tick rate in seconds
func (gl *GameLoop) GetTickRateSeconds() float64 {
	return gl.tickRate.Seconds()
}
