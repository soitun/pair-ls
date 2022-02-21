package lsp_handler

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *LspHandler) handleTextDocumentHover(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if h.clientSendsCursor {
		return
	}
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params struct {
		lsp.TextDocumentPositionParams
	}
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}
	filename, err := h.filenameFromURI(params.TextDocument.URI)
	if err != nil {
		return nil, nil
	}
	h.state.CursorMove(filename, params.Position, nil)
	return nil, nil
}
