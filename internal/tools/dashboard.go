package tools

import (
	"context"

	"connectrpc.com/connect"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	pb "github.com/xhos/null-mcp/internal/gen/null/v1"
)

func (h *Handler) registerDashboard(s *server.MCPServer) {
	s.AddTool(mcp.NewTool("get_dashboard_summary",
		mcp.WithDescription("Get a high-level financial overview: total accounts, total transactions, income vs expenses, transactions in last 30 days, and count of uncategorized transactions. Optionally filter by date range."),
		mcp.WithString("start_date", mcp.Description("Start of date range, format: YYYY-MM-DD.")),
		mcp.WithString("end_date", mcp.Description("End of date range, format: YYYY-MM-DD.")),
	), h.getDashboardSummary)

	s.AddTool(mcp.NewTool("get_financial_summary",
		mcp.WithDescription("Get total balance across all accounts, total debt (credit cards), and net balance. Quick snapshot of overall financial position."),
	), h.getFinancialSummary)

	s.AddTool(mcp.NewTool("get_monthly_comparison",
		mcp.WithDescription("Compare income, expenses, and net amount month-over-month. Returns data for the specified number of past months. Useful for spotting trends."),
		mcp.WithNumber("months_back", mcp.Required(), mcp.Description("Number of months to look back (e.g. 6 for last 6 months).")),
		mcp.WithNumber("account_id", mcp.Description("Optional account ID to filter to a specific account.")),
	), h.getMonthlyComparison)

	s.AddTool(mcp.NewTool("get_top_categories",
		mcp.WithDescription("Get top spending categories ranked by total amount. Shows category slug, label, color, transaction count, and total. Useful for understanding where money goes."),
		mcp.WithString("start_date", mcp.Description("Start of date range, format: YYYY-MM-DD.")),
		mcp.WithString("end_date", mcp.Description("End of date range, format: YYYY-MM-DD.")),
		mcp.WithNumber("limit", mcp.Description("Max number of categories to return.")),
	), h.getTopCategories)

	s.AddTool(mcp.NewTool("get_top_merchants",
		mcp.WithDescription("Get top merchants ranked by total spending. Shows merchant name, transaction count, total amount, and average transaction amount."),
		mcp.WithString("start_date", mcp.Description("Start of date range, format: YYYY-MM-DD.")),
		mcp.WithString("end_date", mcp.Description("End of date range, format: YYYY-MM-DD.")),
		mcp.WithNumber("limit", mcp.Description("Max number of merchants to return.")),
	), h.getTopMerchants)

	s.AddTool(mcp.NewTool("get_spending_trends",
		mcp.WithDescription("Get daily income and expense totals over a date range. Returns a time series of data points. Both start and end dates are required."),
		mcp.WithString("start_date", mcp.Required(), mcp.Description("Start of date range, format: YYYY-MM-DD.")),
		mcp.WithString("end_date", mcp.Required(), mcp.Description("End of date range, format: YYYY-MM-DD.")),
		mcp.WithNumber("category_id", mcp.Description("Optional category ID to filter to a single category.")),
		mcp.WithNumber("account_id", mcp.Description("Optional account ID to filter to a single account.")),
	), h.getSpendingTrends)

	s.AddTool(mcp.NewTool("get_category_spending_comparison",
		mcp.WithDescription("Compare spending by category between the current period and the previous equivalent period. Shows how each category's spending changed period-over-period."),
		mcp.WithString("period_type", mcp.Required(), mcp.Description("Time period to compare. Values: 'PERIOD_TYPE_7_DAYS', 'PERIOD_TYPE_30_DAYS', 'PERIOD_TYPE_90_DAYS', 'PERIOD_TYPE_3_MONTHS', 'PERIOD_TYPE_6_MONTHS', 'PERIOD_TYPE_1_YEAR', 'PERIOD_TYPE_ALL_TIME', or 'PERIOD_TYPE_CUSTOM' (requires custom dates).")),
		mcp.WithString("custom_start_date", mcp.Description("Start date for custom period, format: YYYY-MM-DD. Only used with PERIOD_TYPE_CUSTOM.")),
		mcp.WithString("custom_end_date", mcp.Description("End date for custom period, format: YYYY-MM-DD. Only used with PERIOD_TYPE_CUSTOM.")),
	), h.getCategorySpendingComparison)

	s.AddTool(mcp.NewTool("get_net_worth_history",
		mcp.WithDescription("Get historical net worth over a date range at a chosen granularity. Returns time series data points. Useful for tracking financial progress over time."),
		mcp.WithString("start_date", mcp.Required(), mcp.Description("Start of date range, format: YYYY-MM-DD.")),
		mcp.WithString("end_date", mcp.Required(), mcp.Description("End of date range, format: YYYY-MM-DD.")),
		mcp.WithString("granularity", mcp.Required(), mcp.Description("Data point frequency: 'GRANULARITY_DAY', 'GRANULARITY_WEEK', or 'GRANULARITY_MONTH'.")),
	), h.getNetWorthHistory)
}

func (h *Handler) getDashboardSummary(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	resp, err := h.dashboard.GetDashboardSummary(ctx, connect.NewRequest(&pb.GetDashboardSummaryRequest{
		UserId:    h.userID,
		StartDate: argDate(args, "start_date"),
		EndDate:   argDate(args, "end_date"),
	}))
	if err != nil {
		return h.errResult(err)
	}
	return h.result(resp.Msg)
}

func (h *Handler) getFinancialSummary(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resp, err := h.dashboard.GetFinancialSummary(ctx, connect.NewRequest(&pb.GetFinancialSummaryRequest{UserId: h.userID}))
	if err != nil {
		return h.errResult(err)
	}
	return h.result(resp.Msg)
}

func (h *Handler) getMonthlyComparison(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	r := &pb.GetMonthlyComparisonRequest{
		UserId:    h.userID,
		AccountId: argInt64(args, "account_id"),
	}
	if v := argInt32(args, "months_back"); v != nil {
		r.MonthsBack = *v
	}
	resp, err := h.dashboard.GetMonthlyComparison(ctx, connect.NewRequest(r))
	if err != nil {
		return h.errResult(err)
	}
	return h.result(resp.Msg)
}

func (h *Handler) getTopCategories(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	resp, err := h.dashboard.GetTopCategories(ctx, connect.NewRequest(&pb.GetTopCategoriesRequest{
		UserId:    h.userID,
		StartDate: argDate(args, "start_date"),
		EndDate:   argDate(args, "end_date"),
		Limit:     argInt32(args, "limit"),
	}))
	if err != nil {
		return h.errResult(err)
	}
	return h.result(resp.Msg)
}

func (h *Handler) getTopMerchants(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	resp, err := h.dashboard.GetTopMerchants(ctx, connect.NewRequest(&pb.GetTopMerchantsRequest{
		UserId:    h.userID,
		StartDate: argDate(args, "start_date"),
		EndDate:   argDate(args, "end_date"),
		Limit:     argInt32(args, "limit"),
	}))
	if err != nil {
		return h.errResult(err)
	}
	return h.result(resp.Msg)
}

func (h *Handler) getSpendingTrends(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	resp, err := h.dashboard.GetSpendingTrends(ctx, connect.NewRequest(&pb.GetSpendingTrendsRequest{
		UserId:     h.userID,
		StartDate:  argDate(args, "start_date"),
		EndDate:    argDate(args, "end_date"),
		CategoryId: argInt64(args, "category_id"),
		AccountId:  argInt64(args, "account_id"),
	}))
	if err != nil {
		return h.errResult(err)
	}
	return h.result(resp.Msg)
}

func (h *Handler) getCategorySpendingComparison(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	r := &pb.GetCategorySpendingComparisonRequest{
		UserId:          h.userID,
		CustomStartDate: argDate(args, "custom_start_date"),
		CustomEndDate:   argDate(args, "custom_end_date"),
	}
	if pt := argStr(args, "period_type"); pt != "" {
		if v, ok := pb.PeriodType_value[pt]; ok {
			r.PeriodType = pb.PeriodType(v)
		}
	}
	resp, err := h.dashboard.GetCategorySpendingComparison(ctx, connect.NewRequest(r))
	if err != nil {
		return h.errResult(err)
	}
	return h.result(resp.Msg)
}

func (h *Handler) getNetWorthHistory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	r := &pb.GetNetWorthHistoryRequest{
		UserId:    h.userID,
		StartDate: argDate(args, "start_date"),
		EndDate:   argDate(args, "end_date"),
	}
	if g := argStr(args, "granularity"); g != "" {
		if v, ok := pb.Granularity_value[g]; ok {
			r.Granularity = pb.Granularity(v)
		}
	}
	resp, err := h.dashboard.GetNetWorthHistory(ctx, connect.NewRequest(r))
	if err != nil {
		return h.errResult(err)
	}
	return h.result(resp.Msg)
}
