package affil

import "context"

// TemplateFor returns the NetworkTemplate for the given platform ID.
// This is a stub for FR-AFFIL-003. In a real scenario, this would query the DB.
func (r *Repo) TemplateFor(ctx context.Context, platformID int16) (string, NetworkTemplate, bool) {
	// Query the affiliate_network table
	var network string
	var tmpl NetworkTemplate
	err := r.pool.QueryRow(ctx,
		`SELECT code, base_url, target_param, sub_id_param
		 FROM affiliate_network
		 WHERE platform_id = $1 AND active = true
		 ORDER BY id ASC LIMIT 1`, platformID).Scan(&network, &tmpl.BaseURL, &tmpl.TargetParam, &tmpl.SubIDParam)
	if err != nil {
		return "", NetworkTemplate{}, false
	}
	return network, tmpl, true
}
