package routes_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"golang.org/x/exp/slog"
)

// failOnPanic to prevent Testify assertion mismatch panics from bubbling out of a subtest, ensuring correct reporting
// of failures to the right test case.
func failOnPanic(t *testing.T) {
	if r := recover(); r != nil {
		t.Fatal(r)
	}
}

// NewTestLogger provides a logger for testing, which writes logs within the context of the test being executed.
func NewTestLogger(tb testing.TB) *slog.Logger {
	return slog.New(slog.NewTextHandler(&testLogWriter{tb: tb}, &slog.HandlerOptions{}))
}

// testLogWriter writes logs against the test, which means they are:
// * Only exposed on test failure, or in verbose mode
// * Scoped and associated to the test, which makes output clear
type testLogWriter struct {
	tb testing.TB
}

func (w *testLogWriter) Write(p []byte) (n int, err error) {
	w.tb.Log(string(p))

	return len(p), nil
}

type testCtxKey string

const testCtx testCtxKey = "test"

// contextFromTest allows propagation of contexts used in the application to be verified in mock calls, ensuring that
// the context passed along has been derived from the calling context, not generated independently.
// This is done because mock args are matched by equality (which doesn't work for derived contexts), or need to use a
// matcher function, which provides poor reporting on why there is a mismatch in expected vs actual args.
type contextFromTest struct {
	tb testing.TB

	// unique function allows the instance to never be considered equal, used for cases where there is no test context
	unique func()
}

func (c contextFromTest) String() string {
	if c.tb == nil {
		return "** Context is not derived from test **"
	}

	return c.tb.Name()
}

func (c contextFromTest) GoString() string {
	if c.tb == nil {
		return "contextFromTest{/* Context is not derived from test */}"
	}

	return fmt.Sprintf("contextFromTest{tb: Test(%q)}", c.tb.Name())
}

func withTestContext(ctx context.Context, tb testing.TB) context.Context {
	return context.WithValue(ctx, testCtx, contextFromTest{tb: tb})
}

func testContext(ctx context.Context) contextFromTest {
	v := ctx.Value(testCtx)
	if v == nil {
		// We force the value to be unique here, so it can never compare as equal (non-nil funcs are never equal)
		return contextFromTest{unique: func() {}}
	}

	return v.(contextFromTest)
}

// call2 value type safety for mocking calls which return two values.
type call2[T any, U any] struct {
	m *mock.Call
}

func (c *call2[T, U]) Return(t T, u U) *call2[T, U] {
	c.m.Return(t, u)

	return c
}
