package conf

import (
	"context"
	"fmt"
	"time"

	commonpb "github.com/GUET-BAT/Astraios-S/common-service/pb/commonpb"
	"github.com/GUET-BAT/Astraios-S/user-service/internal/config"

	zconf "github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

const remoteTimeout = 5 * time.Second

// MustLoad loads local config first, then fetches runtime config from common-service.
func MustLoad(path string, c *config.Config) {
	zconf.MustLoad(path, c)
	logx.Must(loadRemoteConfig(c))
}

func loadRemoteConfig(c *config.Config) error {
	if c.ConfigDataId == "" {
		return fmt.Errorf("ConfigDataId is required")
	}

	client, err := zrpc.NewClient(c.CommonService)
	if err != nil {
		return err
	}
	defer client.Conn().Close()

	commonClient := commonpb.NewCommonServiceClient(client.Conn())

	ctx, cancel := context.WithTimeout(context.Background(), remoteTimeout)
	defer cancel()

	resp, err := commonClient.LoadConfig(ctx, &commonpb.LoadConfigRequest{
		NacosDataId: c.ConfigDataId,
	})
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("load config failed: %s", resp.Message)
	}

	if err := zconf.LoadFromYamlBytes([]byte(resp.Config), c); err != nil {
		return err
	}

	return nil
}
