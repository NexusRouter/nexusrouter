package keystore

import (
	"strings"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/repository"
)

func toAPIKeyModels(recs []Record) []repository.APIKeyModel {
	out := make([]repository.APIKeyModel, 0, len(recs))
	now := time.Now().UTC()
	for _, r := range recs {
		id := strings.TrimSpace(r.ID)
		if id == "" {
			id = newRandomID()
		}
		ca := now
		if r.CreatedAt != nil {
			ca = r.CreatedAt.UTC()
		}
		out = append(out, repository.APIKeyModel{
			KeyID:     id,
			Secret:    strings.TrimSpace(r.Secret),
			Disabled:  r.Disabled,
			ExpiresAt: r.ExpiresAt,
			CreatedAt: ca,
			UpdatedAt: now,
		})
	}
	return out
}

func recordsFromModels(rows []repository.APIKeyModel) []Record {
	out := make([]Record, 0, len(rows))
	for _, m := range rows {
		r := Record{
			ID:        m.KeyID,
			Secret:    m.Secret,
			Disabled:  m.Disabled,
			ExpiresAt: m.ExpiresAt,
		}
		if !m.CreatedAt.IsZero() {
			t := m.CreatedAt.UTC()
			r.CreatedAt = &t
		}
		out = append(out, r)
	}
	return out
}
