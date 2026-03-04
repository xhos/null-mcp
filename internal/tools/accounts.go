package tools

import (
	"context"

	"connectrpc.com/connect"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	pb "github.com/xhos/null-mcp/internal/gen/null/v1"
)

func (h *Handler) registerAccounts(s *server.MCPServer) {
	s.AddTool(mcp.NewTool("list_accounts",
		mcp.WithDescription("List all financial accounts (chequing, savings, credit cards, investments) with their current balances. Returns account names, banks, types, currencies, and calculated balances."),
	), h.listAccounts)
}

func (h *Handler) registerCategories(s *server.MCPServer) {
	s.AddTool(mcp.NewTool("list_categories",
		mcp.WithDescription("List all transaction categories. Categories use dot-separated slugs (e.g. 'food.groceries', 'transport.fuel'). Returns slug and color for each."),
		mcp.WithNumber("limit", mcp.Description("Max categories to return (1-100).")),
		mcp.WithNumber("offset", mcp.Description("Number of categories to skip for pagination.")),
	), h.listCategories)
}

func (h *Handler) listAccounts(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resp, err := h.accounts.ListAccounts(ctx, connect.NewRequest(&pb.ListAccountsRequest{UserId: h.userID}))
	if err != nil {
		return h.errResult(err)
	}
	return h.result(resp.Msg)
}

func (h *Handler) listCategories(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	resp, err := h.cats.ListCategories(ctx, connect.NewRequest(&pb.ListCategoriesRequest{
		UserId: h.userID,
		Limit:  argInt32(args, "limit"),
		Offset: argInt32(args, "offset"),
	}))
	if err != nil {
		return h.errResult(err)
	}
	return h.result(resp.Msg)
}
