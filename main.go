package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	pb "null-mcp/internal/gen/null/v1"
	"null-mcp/internal/gen/null/v1/nullv1connect"

	date "google.golang.org/genproto/googleapis/type/date"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

var (
	userID    string
	accounts  nullv1connect.AccountServiceClient
	txns      nullv1connect.TransactionServiceClient
	cats      nullv1connect.CategoryServiceClient
	dashboard nullv1connect.DashboardServiceClient
)

func main() {
	apiURL := os.Getenv("NULL_API_URL")
	apiKey := os.Getenv("NULL_API_KEY")
	userID = os.Getenv("NULL_USER_ID")

	if apiURL == "" || apiKey == "" || userID == "" {
		log.Fatal("required env vars: NULL_API_URL, NULL_API_KEY, NULL_USER_ID")
	}

	apiURL = strings.TrimRight(apiURL, "/")
	httpClient := &http.Client{Timeout: 30 * time.Second}
	opts := connect.WithInterceptors(authInterceptor(apiKey))

	accounts = nullv1connect.NewAccountServiceClient(httpClient, apiURL, opts)
	txns = nullv1connect.NewTransactionServiceClient(httpClient, apiURL, opts)
	cats = nullv1connect.NewCategoryServiceClient(httpClient, apiURL, opts)
	dashboard = nullv1connect.NewDashboardServiceClient(httpClient, apiURL, opts)

	s := server.NewMCPServer("null-mcp", "1.0.0")
	registerTools(s)

	if err := server.ServeStdio(s); err != nil {
		log.Fatal(err)
	}
}

func authInterceptor(apiKey string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("X-Internal-Key", apiKey)
			return next(ctx, req)
		}
	}
}

// result serializes a proto message as JSON for the MCP response.
func result(msg proto.Message) (*mcp.CallToolResult, error) {
	data, err := protojson.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(data)), nil
}

func errResult(err error) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(err.Error()), nil
}

// ----- arg helpers ------------------------------------------------

func argStr(args map[string]any, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

func argInt64(args map[string]any, key string) *int64 {
	if v, ok := args[key].(float64); ok {
		i := int64(v)
		return &i
	}
	return nil
}

func argInt32(args map[string]any, key string) *int32 {
	if v, ok := args[key].(float64); ok {
		i := int32(v)
		return &i
	}
	return nil
}

func argBool(args map[string]any, key string) *bool {
	if v, ok := args[key].(bool); ok {
		return &v
	}
	return nil
}

func argDate(args map[string]any, key string) *date.Date {
	s := argStr(args, key)
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &date.Date{Year: int32(t.Year()), Month: int32(t.Month()), Day: int32(t.Day())}
}

func argTimestamp(args map[string]any, key string) *timestamppb.Timestamp {
	s := argStr(args, key)
	if s == "" {
		return nil
	}
	if len(s) == 10 {
		s += "T00:00:00Z"
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return timestamppb.New(t)
}

func registerTools(s *server.MCPServer) {
	// Accounts

	s.AddTool(mcp.NewTool("list_accounts",
		mcp.WithDescription("List all financial accounts (chequing, savings, credit cards, investments) with their current balances. Returns account names, banks, types, currencies, and calculated balances."),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		resp, err := accounts.ListAccounts(ctx, connect.NewRequest(&pb.ListAccountsRequest{UserId: userID}))
		if err != nil {
			return errResult(err)
		}
		return result(resp.Msg)
	})

	// Categories

	s.AddTool(mcp.NewTool("list_categories",
		mcp.WithDescription("List all transaction categories. Categories use dot-separated slugs (e.g. 'food.groceries', 'transport.fuel'). Returns slug and color for each."),
		mcp.WithNumber("limit", mcp.Description("Max categories to return (1-100).")),
		mcp.WithNumber("offset", mcp.Description("Number of categories to skip for pagination.")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		resp, err := cats.ListCategories(ctx, connect.NewRequest(&pb.ListCategoriesRequest{
			UserId: userID,
			Limit:  argInt32(args, "limit"),
			Offset: argInt32(args, "offset"),
		}))
		if err != nil {
			return errResult(err)
		}
		return result(resp.Msg)
	})

	// Transactions

	s.AddTool(mcp.NewTool("list_transactions",
		mcp.WithDescription("Search and filter transactions. Supports filtering by date range, amount range, direction, account, category, merchant name, and description text. Returns transaction details including amount, merchant, category, date, and account."),
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
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		r := &pb.ListTransactionsRequest{
			UserId:    userID,
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
		resp, err := txns.ListTransactions(ctx, connect.NewRequest(r))
		if err != nil {
			return errResult(err)
		}
		return result(resp.Msg)
	})

	s.AddTool(mcp.NewTool("get_transaction",
		mcp.WithDescription("Get full details of a single transaction by its ID. Returns amount, merchant, category, date, account, notes, exchange rate, and foreign amount if applicable."),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("The transaction ID.")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		id := int64(0)
		if v := argInt64(args, "id"); v != nil {
			id = *v
		}
		resp, err := txns.GetTransaction(ctx, connect.NewRequest(&pb.GetTransactionRequest{
			UserId: userID,
			Id:     id,
		}))
		if err != nil {
			return errResult(err)
		}
		return result(resp.Msg)
	})

	// Dashboard

	s.AddTool(mcp.NewTool("get_dashboard_summary",
		mcp.WithDescription("Get a high-level financial overview: total accounts, total transactions, income vs expenses, transactions in last 30 days, and count of uncategorized transactions. Optionally filter by date range."),
		mcp.WithString("start_date", mcp.Description("Start of date range, format: YYYY-MM-DD.")),
		mcp.WithString("end_date", mcp.Description("End of date range, format: YYYY-MM-DD.")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		resp, err := dashboard.GetDashboardSummary(ctx, connect.NewRequest(&pb.GetDashboardSummaryRequest{
			UserId:    userID,
			StartDate: argDate(args, "start_date"),
			EndDate:   argDate(args, "end_date"),
		}))
		if err != nil {
			return errResult(err)
		}
		return result(resp.Msg)
	})

	s.AddTool(mcp.NewTool("get_financial_summary",
		mcp.WithDescription("Get total balance across all accounts, total debt (credit cards), and net balance. Quick snapshot of overall financial position."),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		resp, err := dashboard.GetFinancialSummary(ctx, connect.NewRequest(&pb.GetFinancialSummaryRequest{UserId: userID}))
		if err != nil {
			return errResult(err)
		}
		return result(resp.Msg)
	})

	s.AddTool(mcp.NewTool("get_monthly_comparison",
		mcp.WithDescription("Compare income, expenses, and net amount month-over-month. Returns data for the specified number of past months. Useful for spotting trends."),
		mcp.WithNumber("months_back", mcp.Required(), mcp.Description("Number of months to look back (e.g. 6 for last 6 months).")),
		mcp.WithNumber("account_id", mcp.Description("Optional account ID to filter to a specific account.")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		r := &pb.GetMonthlyComparisonRequest{
			UserId:    userID,
			AccountId: argInt64(args, "account_id"),
		}
		if v := argInt32(args, "months_back"); v != nil {
			r.MonthsBack = *v
		}
		resp, err := dashboard.GetMonthlyComparison(ctx, connect.NewRequest(r))
		if err != nil {
			return errResult(err)
		}
		return result(resp.Msg)
	})

	s.AddTool(mcp.NewTool("get_top_categories",
		mcp.WithDescription("Get top spending categories ranked by total amount. Shows category slug, label, color, transaction count, and total. Useful for understanding where money goes."),
		mcp.WithString("start_date", mcp.Description("Start of date range, format: YYYY-MM-DD.")),
		mcp.WithString("end_date", mcp.Description("End of date range, format: YYYY-MM-DD.")),
		mcp.WithNumber("limit", mcp.Description("Max number of categories to return.")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		resp, err := dashboard.GetTopCategories(ctx, connect.NewRequest(&pb.GetTopCategoriesRequest{
			UserId:    userID,
			StartDate: argDate(args, "start_date"),
			EndDate:   argDate(args, "end_date"),
			Limit:     argInt32(args, "limit"),
		}))
		if err != nil {
			return errResult(err)
		}
		return result(resp.Msg)
	})

	s.AddTool(mcp.NewTool("get_top_merchants",
		mcp.WithDescription("Get top merchants ranked by total spending. Shows merchant name, transaction count, total amount, and average transaction amount."),
		mcp.WithString("start_date", mcp.Description("Start of date range, format: YYYY-MM-DD.")),
		mcp.WithString("end_date", mcp.Description("End of date range, format: YYYY-MM-DD.")),
		mcp.WithNumber("limit", mcp.Description("Max number of merchants to return.")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		resp, err := dashboard.GetTopMerchants(ctx, connect.NewRequest(&pb.GetTopMerchantsRequest{
			UserId:    userID,
			StartDate: argDate(args, "start_date"),
			EndDate:   argDate(args, "end_date"),
			Limit:     argInt32(args, "limit"),
		}))
		if err != nil {
			return errResult(err)
		}
		return result(resp.Msg)
	})

	s.AddTool(mcp.NewTool("get_spending_trends",
		mcp.WithDescription("Get daily income and expense totals over a date range. Returns a time series of data points. Both start and end dates are required."),
		mcp.WithString("start_date", mcp.Required(), mcp.Description("Start of date range, format: YYYY-MM-DD.")),
		mcp.WithString("end_date", mcp.Required(), mcp.Description("End of date range, format: YYYY-MM-DD.")),
		mcp.WithNumber("category_id", mcp.Description("Optional category ID to filter to a single category.")),
		mcp.WithNumber("account_id", mcp.Description("Optional account ID to filter to a single account.")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		resp, err := dashboard.GetSpendingTrends(ctx, connect.NewRequest(&pb.GetSpendingTrendsRequest{
			UserId:     userID,
			StartDate:  argDate(args, "start_date"),
			EndDate:    argDate(args, "end_date"),
			CategoryId: argInt64(args, "category_id"),
			AccountId:  argInt64(args, "account_id"),
		}))
		if err != nil {
			return errResult(err)
		}
		return result(resp.Msg)
	})

	s.AddTool(mcp.NewTool("get_category_spending_comparison",
		mcp.WithDescription("Compare spending by category between the current period and the previous equivalent period. Shows how each category's spending changed period-over-period."),
		mcp.WithString("period_type", mcp.Required(), mcp.Description("Time period to compare. Values: 'PERIOD_TYPE_7_DAYS', 'PERIOD_TYPE_30_DAYS', 'PERIOD_TYPE_90_DAYS', 'PERIOD_TYPE_3_MONTHS', 'PERIOD_TYPE_6_MONTHS', 'PERIOD_TYPE_1_YEAR', 'PERIOD_TYPE_ALL_TIME', or 'PERIOD_TYPE_CUSTOM' (requires custom dates).")),
		mcp.WithString("custom_start_date", mcp.Description("Start date for custom period, format: YYYY-MM-DD. Only used with PERIOD_TYPE_CUSTOM.")),
		mcp.WithString("custom_end_date", mcp.Description("End date for custom period, format: YYYY-MM-DD. Only used with PERIOD_TYPE_CUSTOM.")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		r := &pb.GetCategorySpendingComparisonRequest{
			UserId:          userID,
			CustomStartDate: argDate(args, "custom_start_date"),
			CustomEndDate:   argDate(args, "custom_end_date"),
		}
		if pt := argStr(args, "period_type"); pt != "" {
			if v, ok := pb.PeriodType_value[pt]; ok {
				r.PeriodType = pb.PeriodType(v)
			}
		}
		resp, err := dashboard.GetCategorySpendingComparison(ctx, connect.NewRequest(r))
		if err != nil {
			return errResult(err)
		}
		return result(resp.Msg)
	})

	s.AddTool(mcp.NewTool("get_net_worth_history",
		mcp.WithDescription("Get historical net worth over a date range at a chosen granularity. Returns time series data points. Useful for tracking financial progress over time."),
		mcp.WithString("start_date", mcp.Required(), mcp.Description("Start of date range, format: YYYY-MM-DD.")),
		mcp.WithString("end_date", mcp.Required(), mcp.Description("End of date range, format: YYYY-MM-DD.")),
		mcp.WithString("granularity", mcp.Required(), mcp.Description("Data point frequency: 'GRANULARITY_DAY', 'GRANULARITY_WEEK', or 'GRANULARITY_MONTH'.")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		r := &pb.GetNetWorthHistoryRequest{
			UserId:    userID,
			StartDate: argDate(args, "start_date"),
			EndDate:   argDate(args, "end_date"),
		}
		if g := argStr(args, "granularity"); g != "" {
			if v, ok := pb.Granularity_value[g]; ok {
				r.Granularity = pb.Granularity(v)
			}
		}
		resp, err := dashboard.GetNetWorthHistory(ctx, connect.NewRequest(r))
		if err != nil {
			return errResult(err)
		}
		return result(resp.Msg)
	})
}
