package handler

import (
	"context"
	"html/template"
	"net/http"
	"os"

	"github.com/arussellsaw/youneedaspreadsheet/pkg/stripe"

	"github.com/arussellsaw/youneedaspreadsheet/pkg/sheets"

	"github.com/arussellsaw/youneedaspreadsheet/pkg/truelayer"

	"github.com/arussellsaw/youneedaspreadsheet/domain"

	"github.com/monzo/slog"
)

type indexData struct {
	User                 *domain.User
	HasTruelayer         bool
	HasSheets            bool
	HasStripe            bool
	StripePublishableKey string
	StripePriceID        string
	Accounts             []truelayer.Account
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	t := template.New("index.html")
	t, err := t.ParseFiles("tmpl/index.html")
	if err != nil {
		slog.Error(ctx, "Error parsing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
	u := domain.UserFromContext(ctx)
	hasTL, accs := hasTruelayer(ctx, u)
	err = t.Execute(w, indexData{
		User:                 u,
		HasTruelayer:         hasTL,
		HasSheets:            hasSheets(ctx, u),
		HasStripe:            hasStripe(ctx, u),
		Accounts:             accs,
		StripePublishableKey: os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		StripePriceID:        os.Getenv("STRIPE_PRICE_ID"),
	})
	if err != nil {
		slog.Error(ctx, "Index: %s", err)
	}
}

func hasTruelayer(ctx context.Context, user *domain.User) (bool, []truelayer.Account) {
	if user == nil {
		return false, nil
	}
	tl, err := truelayer.NewClient(ctx, user.ID)
	if err != nil {
		slog.Error(ctx, "error getting truelayer client: %s", err)
		return false, nil
	}
	accs, err := tl.Accounts(ctx)
	if err != nil {
		slog.Error(ctx, "error getting authorised accounts: %s", err)
		return false, nil
	}
	return len(accs) != 0, accs
}

func hasSheets(ctx context.Context, user *domain.User) bool {
	if user == nil {
		return false
	}
	s, err := sheets.NewClient(ctx, user.ID)
	if err != nil {
		slog.Error(ctx, "error getting sheets client: %s", err)
		return false
	}
	if s == nil {
		return false
	}
	if user.SheetID != "" {
		_, err = s.Get(ctx, user.SheetID)
		if err != nil {
			return false
		}
	}
	return true
}

func hasStripe(ctx context.Context, user *domain.User) bool {
	if user == nil {
		return false
	}
	ok, err := stripe.HasSubscription(ctx, user)
	if err != nil {
		slog.Error(ctx, "error getting stripe subscription: %s", err)
		return false
	}
	return ok
}
