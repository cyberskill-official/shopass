package captcha

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrCaptchaBudget = errors.New("captcha budget exhausted")

type Solver struct {
	dailyBudgetSolves   int
	maxSolvesPerTarget  int
	
	solvesToday         int
	targetSolves        map[string]int
	lastReset           time.Time
	mu                  sync.Mutex
}

func NewSolver(dailyBudget, maxPerTarget int) *Solver {
	return &Solver{
		dailyBudgetSolves:  dailyBudget,
		maxSolvesPerTarget: maxPerTarget,
		targetSolves:       make(map[string]int),
		lastReset:          time.Now(),
	}
}

// Solve giả lập việc giải CAPTCHA (gọi dịch vụ ngoài hoặc farm behavior).
// Trả về ErrCaptchaBudget nếu vượt trần số lần/ngày hoặc số lần/target.
func (s *Solver) Solve(ctx context.Context, targetURL string, kind CaptchaKind) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	// Reset daily if needed
	if now.YearDay() != s.lastReset.YearDay() || now.Year() != s.lastReset.Year() {
		s.solvesToday = 0
		s.targetSolves = make(map[string]int)
		s.lastReset = now
	}

	if s.solvesToday >= s.dailyBudgetSolves {
		return ErrCaptchaBudget
	}

	if s.targetSolves[targetURL] >= s.maxSolvesPerTarget {
		return ErrCaptchaBudget
	}

	// Thực hiện giải thật ở đây (gọi dịch vụ ngoài)
	// ...

	s.solvesToday++
	s.targetSolves[targetURL]++

	return nil
}
