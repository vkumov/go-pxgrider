package connection

import (
	"context"
	"fmt"

	gopxgrid "github.com/vkumov/go-pxgrid"
)

func (c *Connection) RefreshAccountState(ctx context.Context) (gopxgrid.AccountActivateResponse, error) {
	px, err := c.PX()
	if err != nil {
		return gopxgrid.AccountActivateResponse{}, fmt.Errorf("failed to get pxGrid service: %w", err)
	}

	if c.credentials.Type == CredentialsTypePassword && c.credentials.Password == "" {
		resp, err := px.AccountCreate(ctx)
		if err != nil {
			return gopxgrid.AccountActivateResponse{}, fmt.Errorf("failed to create account: %w", err)
		}

		c.SetPasswordBasedAuth(resp.Password)
		if err = c.storeUnsaved(ctx); err != nil {
			return gopxgrid.AccountActivateResponse{}, fmt.Errorf("failed to store connection: %w", err)
		}
	}

	c.log.Debug().Msg("Activating account")
	act, err := px.AccountActivate(ctx)
	if err != nil {
		return gopxgrid.AccountActivateResponse{}, fmt.Errorf("failed to activate account: %w", err)
	}

	c.log.Info().Str("state", string(act.AccountState)).Msg("Account state updated")
	c.SetState(act.AccountState)
	if err = c.storeUnsaved(ctx); err != nil {
		return gopxgrid.AccountActivateResponse{}, fmt.Errorf("failed to store connection: %w", err)
	}

	return act, nil
}
