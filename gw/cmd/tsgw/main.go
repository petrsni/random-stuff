package main

import (
	"github.com/alecthomas/kong"
	"petr.local/inflgw/internal/app/tsgw"
)

func main() {

	//ctx, cancel := context.WithCancel(context.Background())

	cli := &CLI{}
	kctx := kong.Parse(cli)

	err := kctx.Run()
	if err != nil {
		kctx.FatalIfErrorf(err)
	}

}

type CLI struct {
	Influx InfluxCmd `cmd:"" help:"run InfluxDB gateway"`
}

type InfluxCmd struct {
	Url   string `name:"url" default:"http://localhost:8086" help:"InfluxDB URL"`
	Port  int    `name:"port" default:"888" help:"Port to listen on"`
	Token string `name:"token" default:"" help:"InfluxDB token"`
	User  string `name:"user" default:"" help:"API user"`
	Pass  string `name:"pass" default:"" help:"API pass"`
}

func (c *InfluxCmd) Run() error {
	gw, err := tsgw.NewInfluxGw(tsgw.AppParams{
		InfluxUrl:   c.Url,
		InfluxToken: c.Token,
		AppPort:     c.Port,
		AppUser:     c.User,
		AppPass:     c.Pass,
	})
	if err != nil {
		return err
	}
	return gw.Run()
}
