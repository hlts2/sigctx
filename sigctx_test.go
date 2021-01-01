package sigctx

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/goleak"
)

func TestWithCancel(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx, cancel := WithCancel(context.Background())
	defer cancel()

	fmt.Println(<-ctx.Done())
}
