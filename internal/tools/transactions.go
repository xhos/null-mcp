package tools

import (
	"context"

	"connectrpc.com/connect"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	pb "github.com/xhos/null-mcp/internal/gen/null/v1"
)

func (h *Handler) registerTransactions(s *server.MCPServer) {
	s.AddTool(mcp.NewTool("list_transactions",
		mcp.WithDescription("Search and filter transactions. Supports filtering by date range, direction, account, merchant name, and description text. Returns transaction details including amount, merchant, category, date, and account."),
		mcp.WithNumber("account_id", mcp.Description("Filter to a specific account ID.")),
		mcp.WithString("start_date", mcp.Description("Start of date range, format: YYYY-MM-DD.")),
		mcp.WithString("end_date", mcp.Description("End of date range, format: YYYY-MM-DD.")),
		mcp.WithNumber("limit", mcp.Description("Max transactions to return (1-1000). Default is 50.")),
		mcp.WithNumber("offset", mcp.Description("Number of transactions to skip for pagination.")),
		mcp.WithString("direction", mcp.Description("Filter by direction: 'DIRECTION_INCOMING' for income/deposits, 'DIRECTION_OUTGOING' for expenses/payments.")),
		mcp.WithString("merchant_query", mcp.Description("Search by merchant name (partial match, case-insensitive).")),
		mcp.WithString("description_query", mcp.Description("Search by transaction description (partial match, case-insensitive).")),
		mcp.WithString("currency", mcp.Description("Filter by 3-letter currency code (e.g. 'USD', 'EUR').")),
		mcp.WithBoolean("uncategorized", mcp.Description("Set to true to only return transactions without a category.")),
	), h.listTransactions)

	s.AddTool(mcp.NewTool("get_transaction",
		mcp.WithDescription("Get full details of a single transaction by its ID. Returns amount, merchant, category, date, account, notes, exchange rate, and foreign amount if applicable."),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("The transaction ID.")),
	), h.getTransaction)
}

func (h *Handler) listTransactions(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	r := &pb.ListTransactionsRequest{
		UserId:    h.userID,
		AccountId: argInt64(args, "account_id"),
		Limit:     argInt32(args, "limit"),
		Offset:    argInt32(args, "offset"),
		StartDate: argTimestamp(args, "start_date"),
		EndDate:   argTimestamp(args, "end_date"),
	}
	if s := argStr(args, "merchant_query"); s != "" {
		r.MerchantQuery = &s
	}
	if s := argStr(args, "description_query"); s != "" {
		r.DescriptionQuery = &s
	}
	if s := argStr(args, "currency"); s != "" {
		r.Currency = &s
	}
	r.Uncategorized = argBool(args, "uncategorized")
	if d := argStr(args, "direction"); d != "" {
		if v, ok := pb.TransactionDirection_value[d]; ok {
			dir := pb.TransactionDirection(v)
			r.Direction = &dir
		}
	}
	resp, err := h.txns.ListTransactions(ctx, connect.NewRequest(r))
	if err != nil {
		return h.errResult(err)
	}
	return h.result(resp.Msg)
}

func (h *Handler) getTransaction(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	id := int64(0)
	if v := argInt64(args, "id"); v != nil {
		id = *v
	}
	resp, err := h.txns.GetTransaction(ctx, connect.NewRequest(&pb.GetTransactionRequest{
		UserId: h.userID,
		Id:     id,
	}))
	if err != nil {
		return h.errResult(err)
	}
	return h.result(resp.Msg)
}
