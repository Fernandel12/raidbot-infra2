package rbapi

import (
	"context"
)

func (svc *service) ToolStatus(context.Context, *ToolStatus_Input) (*ToolStatus_Output, error) {
	return &ToolStatus_Output{
		EverythingIsOk: true,
	}, nil
}
