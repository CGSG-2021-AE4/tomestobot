package log

import (
	"context"
	"fmt"
	"log/slog"
)

// Here are some small outputs

// Deffered output
// Like a proxy that allows you to set output receiver later
// Output set/get is by =
type DeferredOutput struct {
	Output Output
}

// Output interface implementation
func (o *DeferredOutput) Handle(ctx context.Context, groups []string, record slog.Record) error {
	if o.Output == nil {
		return fmt.Errorf("deffered output is nil")
	}
	return o.Output.Handle(ctx, groups, record)
}

// Default creation function
func NewDefferedOutput() *DeferredOutput {
	return &DeferredOutput{}
}

// Multi output - allow you to combine outputs
type multiOutput struct {
	outputs []Output
}

func (out *multiOutput) Handle(ctx context.Context, groups []string, record slog.Record) error {
	for _, o := range out.outputs { // Suppose that it is not null
		if err := o.Handle(ctx, groups, record); err != nil {
			return err
		}
	}
	return nil
}

func NewMultiOutput(outputs ...Output) Output {
	return &multiOutput{
		outputs: outputs,
	}
}
