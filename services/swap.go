package services

import (
	"context"

	"github.com/2HgO/quidax-go/types/requests"
	"github.com/2HgO/quidax-go/types/responses"
)

type InstantSwapService interface {
	CreateInstantSwap(ctx context.Context, req *requests.CreateInstantSwapRequest) (*responses.Response[*responses.InstantSwapQuotationResponseData], error)
	ConfirmInstantSwap(ctx context.Context, req *requests.ConfirmInstanSwapRequest) (*responses.Response[*responses.InstantSwapResponseData], error)
	RefreshInstantSwap(ctx context.Context, req *requests.RefreshInstantSwapRequest) (*responses.Response[*responses.InstantSwapQuotationResponseData], error)
	FetchInstantSwapTransactions(ctx context.Context, req *requests.FetchInstantSwapTransactionRequest) (*responses.Response[*responses.InstantSwapResponseData], error)
	GetInstantSwapTransactions(ctx context.Context, req *requests.GetInstantSwapTransactionsRequest) (*responses.Response[[]*responses.InstantSwapResponseData], error)
	QuoteInstantSwap(ctx context.Context, req *requests.CreateInstantSwapRequest) (*responses.Response[*responses.QuoteInstantSwapResponseData], error)
}

type instantSwapService struct {
	service
}

func (i *instantSwapService) CreateInstantSwap(ctx context.Context, req *requests.CreateInstantSwapRequest) (*responses.Response[*responses.InstantSwapQuotationResponseData], error) {
	panic("not implemented")
}
