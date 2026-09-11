package zalo

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"
)

// Message is a Zalo OA / ZNS outbound unit. Live send requires Stephen-provisioned
// OA credentials (R23); without them providers must refuse.
type Message struct {
	UserOAID string
	Template string
	Body     string
}

type SendResult int

const (
	ResultSent SendResult = iota
	ResultRetry
	ResultFailed
)

type SendOutcome struct {
	Result            SendResult
	RetryAfter        time.Duration
	ProviderMessageID string
}

type Provider interface {
	Send(ctx context.Context, msg Message) (SendOutcome, error)
}

// LogProvider is the default fail-closed path: log intent, never claim delivery.
type LogProvider struct {
	log  *slog.Logger
	name string
}

func NewLogProvider(log *slog.Logger, name string) LogProvider {
	if log == nil {
		log = slog.Default()
	}
	if name == "" {
		name = "noop"
	}
	return LogProvider{log: log, name: name}
}

func (p LogProvider) Send(_ context.Context, msg Message) (SendOutcome, error) {
	p.log.Info("zalo noop provider refused message",
		"provider", p.name,
		"user_oa_id", msg.UserOAID,
		"template", msg.Template,
	)
	return SendOutcome{Result: ResultFailed, ProviderMessageID: "noop"}, nil
}

// Configured reports whether live Zalo OA/ZNS secrets appear in the environment.
// Used by main to choose noop vs (future) live client — never invents a send path.
func Configured() bool {
	oa := strings.TrimSpace(os.Getenv("ZALO_OA_ID"))
	secret := strings.TrimSpace(os.Getenv("ZALO_OA_SECRET"))
	token := strings.TrimSpace(os.Getenv("ZALO_OA_ACCESS_TOKEN"))
	return oa != "" && (secret != "" || token != "")
}
