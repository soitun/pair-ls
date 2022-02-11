package lsp_handler

import (
	"context"
	"encoding/json"
	"pair-ls/util"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type CursorMoveRequest struct {
	TextDocument lsp.TextDocumentIdentifier `json:"textDocument"`
	Position     lsp.Position               `json:"position"`
	Range        *lsp.Range                 `json:"range"`
}

func (h *lspHandler) handleCursorMove(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params CursorMoveRequest
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}
	filename, err := util.FromURI(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	h.state.CursorMove(filename, params.Position, params.Range)
	return nil, nil
}