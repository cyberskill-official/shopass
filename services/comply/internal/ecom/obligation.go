package ecom

import "context"

func (s *Service) Obligations(ctx context.Context) ([]EcommerceObligation, error) {
	return s.repo.Obligations(ctx)
}

func (s *Service) MarkObligation(ctx context.Context, key string, status string) error {
	return s.repo.MarkObligation(ctx, key, status)
}

func (s *Service) Outstanding(ctx context.Context) ([]EcommerceObligation, error) {
	obs, err := s.repo.Obligations(ctx)
	if err != nil {
		return nil, err
	}
	var out []EcommerceObligation
	for _, o := range obs {
		if o.Status != "done" && o.Status != "approved" && o.Status != "n_a" {
			out = append(out, o)
		}
	}
	return out, nil
}
