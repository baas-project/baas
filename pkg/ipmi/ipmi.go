package ipmi

import (
	"context"
	"github.com/baas-project/bmc"
	"github.com/baas-project/bmc/pkg/ipmi"
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
			MaxPrivilegeLevel: ipmi.PrivilegeLevelAdministrator,
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
	return c.session.GetChassisStatus(ctx)
}

func (c *Connection) Reboot(ctx context.Context) error {
	return c.session.ChassisControl(ctx, ipmi.ChassisControlPowerCycle)
}

func (c *Connection) GetBootDev(ctx context.Context, req *ipmi.GetSystemBootOptionsReq) (*ipmi.GetSystemBootOptionsRsp, error) {
	return c.session.GetSystemBootOptions(ctx, req)
}

func (c *Connection) SetBootDev(ctx context.Context, req *ipmi.SetSystemBootOptionsReq) error {
	return c.session.SetSystemBootOptions(ctx, req)
}
