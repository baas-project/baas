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
	a := GetBootDevCmd{
		Req: GetBootDevReq{
			ParameterSelector: 5,
		},
	}

	if _, err := c.session.SendCommand(ctx, &a); err != nil {
		return &a.Rsp, err
	}

	return &a.Rsp, nil
}
