package packformat

import (
	"context"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"io"
)

func ValidateNoticeStream(ctx context.Context, source io.Reader) error {
	return formatError(envelope.ValidateNoticeStream(ctx, source))
}
