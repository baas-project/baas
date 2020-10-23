package ipmi

import (
	"context"
	"github.com/gebn/bmc"
	"github.com/gebn/bmc/pkg/ipmi"
)

type Connection struct {
	session *bmc.V2Session
}

func NewConnection(ctx context.Context, addr, username, password string) (*Connection, error) {
	dial, err := bmc.DialV2(addr)
	if err != nil {
		return nil, err
	}

	session, err := dial.NewV2Session(ctx, &bmc.V2SessionOpts{
		SessionOpts: bmc.SessionOpts{
			Username:          username,
			Password:          []byte(password),
			MaxPrivilegeLevel: ipmi.PrivilegeLevelOperator,
		},
	})
	if err != nil {
		return nil, err
	}

	return &Connection{
		session,
	}, nil
}

func (c *Connection) ChassisStatus(ctx context.Context) (*ipmi.GetChassisStatusRsp, error) {
	res, err := c.session.GetChassisStatus(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Connection) Reboot(ctx context.Context) error {
	return c.session.ChassisControl(ctx, ipmi.ChassisControlPowerCycle)
}

func (c *Connection) GetBootDev(ctx context.Context) (*GetBootDevRsp, error) {
	cmd := &GetBootDevCmd{
		Req: GetBootDevReq{
			ParameterSelector: 5,
		},
	}

	if err := bmc.ValidateResponse(c.session.SendCommand(ctx, cmd)); err != nil {
		return nil, err
	}

	return &cmd.Rsp, nil
}
