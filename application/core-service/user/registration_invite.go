package user

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
)

var (
	errRegistrationNoVipDomainPool = errors.New("no non-official vip domain configured")
	errRegistrationVipDomainBusy   = errors.New("failed to allocate unique vip domain")
)

func isRegistrationInviteConfigErr(err error) bool {
	return errors.Is(err, errRegistrationNoVipDomainPool) ||
		errors.Is(err, errRegistrationVipDomainBusy)
}

// rootDomainFromHost 取 host 最后两段作为根域，如 www.niceboxs.com → niceboxs.com。
func rootDomainFromHost(host string) string {
	host = normalizeRegistrationHost(host)
	if host == "" {
		return ""
	}
	parts := strings.Split(host, ".")
	if len(parts) <= 2 {
		return host
	}
	return parts[len(parts)-2] + "." + parts[len(parts)-1]
}

func normalizeRegistrationHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "" {
		return ""
	}
	if i := strings.Index(host, ":"); i >= 0 {
		host = host[:i]
	}
	return host
}

type registrationInviteProfile struct {
	ParentUserID uint
	VipConfigID  uint
	VipDomain    string
	Apply        bool
}

func (s *UserService) resolveRegistrationInviteProfile(ctx context.Context, registrationHost string) (*registrationInviteProfile, error) {
	host := normalizeRegistrationHost(registrationHost)
	if host == "" {
		return &registrationInviteProfile{Apply: false}, nil
	}
	root := rootDomainFromHost(host)
	if root == "" {
		return &registrationInviteProfile{Apply: false}, nil
	}

	cfg, err := s.vipDomains.FindByDomain(ctx, root)
	if err != nil {
		if isMissingSchemaErr(err) {
			return &registrationInviteProfile{Apply: false}, nil
		}
		return nil, err
	}
	if cfg == nil {
		return &registrationInviteProfile{Apply: false}, nil
	}

	var parentUserID uint
	if cfg.IsOfficial {
		parentUserID = 0
	} else {
		parent, err := s.users.FindByVipDomain(ctx, host)
		if err != nil {
			if isMissingSchemaErr(err) {
				return &registrationInviteProfile{Apply: false}, nil
			}
			return nil, err
		}
		if parent == nil {
			return s.buildDefaultRegistrationInviteProfile(ctx, 0)
		}
		parentUserID = parent.ID
	}

	return s.buildDefaultRegistrationInviteProfile(ctx, parentUserID)
}

func (s *UserService) buildDefaultRegistrationInviteProfile(ctx context.Context, parentUserID uint) (*registrationInviteProfile, error) {
	vipID, err := s.defaultVipConfigID(ctx)
	if err != nil {
		return nil, err
	}
	assigned, err := s.generateAssignedVipDomain(ctx)
	if err != nil {
		return nil, err
	}
	return &registrationInviteProfile{
		ParentUserID: parentUserID,
		VipConfigID:  vipID,
		VipDomain:    assigned,
		Apply:        true,
	}, nil
}

func (s *UserService) defaultVipConfigID(ctx context.Context) (uint, error) {
	row, err := s.vipConfigs.FindDefault(ctx)
	if err != nil {
		if isMissingSchemaErr(err) {
			return 1, nil
		}
		return 0, err
	}
	if row != nil {
		return row.ID, nil
	}
	return 1, nil
}

func (s *UserService) generateAssignedVipDomain(ctx context.Context) (string, error) {
	pool, err := s.vipDomains.ListNonOfficial(ctx)
	if err != nil {
		if isMissingSchemaErr(err) {
			return "", errRegistrationNoVipDomainPool
		}
		return "", err
	}
	if len(pool) == 0 {
		return "", errRegistrationNoVipDomainPool
	}

	tryWithPrefix := func(prefix string) (string, bool, error) {
		base, err := pickRandomVipDomainBase(pool)
		if err != nil {
			return "", false, err
		}
		candidate := buildVipSubdomain(prefix, base.Domain)
		taken, err := s.isVipDomainTaken(ctx, candidate)
		if err != nil {
			return "", false, err
		}
		if !taken {
			return candidate, true, nil
		}
		return "", false, nil
	}

	for try := 0; try < vipDomainPinyinMaxTries; try++ {
		prefix, err := randomPinyinPrefix()
		if err != nil {
			return "", err
		}
		if candidate, ok, err := tryWithPrefix(prefix); err != nil {
			return "", err
		} else if ok {
			return candidate, nil
		}
	}

	for try := 0; try < vipDomainLetterMaxTries; try++ {
		prefix, err := randomAlphaLower(vipDomainRandomLen)
		if err != nil {
			return "", err
		}
		if candidate, ok, err := tryWithPrefix(prefix); err != nil {
			return "", err
		} else if ok {
			return candidate, nil
		}
	}

	return "", errRegistrationVipDomainBusy
}

func buildVipSubdomain(prefix, baseDomain string) string {
	baseDomain = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(baseDomain)), ".")
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	return prefix + "." + baseDomain
}

func (s *UserService) isVipDomainTaken(ctx context.Context, vipDomain string) (bool, error) {
	existing, err := s.users.FindByVipDomain(ctx, vipDomain)
	if err != nil {
		if isMissingSchemaErr(err) {
			return false, nil
		}
		return false, err
	}
	return existing != nil, nil
}

func pickRandomVipDomainBase(pool []entity.VipDomainConfig) (*entity.VipDomainConfig, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(pool))))
	if err != nil {
		return nil, err
	}
	row := pool[n.Int64()]
	return &row, nil
}

func randomAlphaLower(n int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b[i] = letters[idx.Int64()]
	}
	return string(b), nil
}

func applyRegistrationInviteProfile(u *entity.Users, p *registrationInviteProfile) {
	if p == nil || !p.Apply {
		return
	}
	u.ParentUserID = p.ParentUserID
	u.VipConfigID = p.VipConfigID
	if p.VipDomain != "" {
		d := p.VipDomain
		u.VipDomain = &d
	}
}
