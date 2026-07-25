package proxy

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type Vault interface {
	ProxyCreds(tier Tier) ProxyCreds
}

type ProxyCreds struct {
	User string
	Pass string
}

type Pool struct {
	mu          sync.Mutex
	guard       *CostGuard
	vault       Vault
	bannedIPs   map[string]time.Time
	cooldown    time.Duration
	rotateIndex int
	proxies     []string
}

func NewPool(guard *CostGuard, vault Vault) *Pool {
	proxyList := os.Getenv("PROXY_LIST")
	var proxies []string
	if proxyList != "" {
		proxies = strings.Split(proxyList, ",")
	}

	return &Pool{
		guard:     guard,
		vault:     vault,
		bannedIPs: make(map[string]time.Time),
		cooldown:  5 * time.Minute, // ví dụ 5 phút cooldown
		proxies:   proxies,
	}
}

func (p *Pool) Acquire(ctx context.Context, tier Tier, country string) (ProxySession, error) {
	decision, err := p.guard.Evaluate(ctx, time.Now())
	if err != nil {
		return ProxySession{}, err
	}

	if decision == BlockCold {
		// Ở đây ta giả định Acquire không biết tier của job (hot/cold)
		// Thực tế orchestrator sẽ hỏi CanProceed trước.
		// Ta có thể trả về một lỗi cụ thể.
		return ProxySession{}, fmt.Errorf("cost-guard: block cold")
	}

	if decision == DowngradeTier {
		// Hạ tier nếu có thể
		if tier == TierEnterprise {
			tier = TierMid
		} else if tier == TierMid {
			tier = TierBudget
		}
	}

	ip := p.rotate(tier, country)
	creds := p.vault.ProxyCreds(tier)
	return ProxySession{URL: ip, User: creds.User, Pass: creds.Pass, Country: country, IP: ip}, nil
}

func (p *Pool) rotate(tier Tier, country string) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	var baseIP string
	if len(p.proxies) > 0 {
		p.rotateIndex++
		baseIP = p.proxies[p.rotateIndex%len(p.proxies)]
	} else {
		// Fallback to generated IP if PROXY_LIST is empty
		p.rotateIndex++
		baseIP = fmt.Sprintf("http://proxy.%s.%s:%d", tier, country, 8000+p.rotateIndex%100)
	}

	// Tránh IP bị ban
	for {
		if bannedAt, ok := p.bannedIPs[baseIP]; ok {
			if time.Since(bannedAt) < p.cooldown {
				p.rotateIndex++
				if len(p.proxies) > 0 {
					baseIP = p.proxies[p.rotateIndex%len(p.proxies)]
				} else {
					baseIP = fmt.Sprintf("http://proxy.%s.%s:%d", tier, country, 8000+p.rotateIndex%100)
				}
				continue
			}
			// Hết cooldown, xóa khỏi map
			delete(p.bannedIPs, baseIP)
		}
		break
	}

	return baseIP
}

func (p *Pool) MarkBanned(ip string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.bannedIPs[ip] = time.Now()
}
