package conf

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	commonpb "github.com/GUET-BAT/Astraios-S/common-service/pb/commonpb"
	"github.com/GUET-BAT/Astraios-S/user-service/internal/config"

	zconf "github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"gopkg.in/yaml.v3"
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

	return mergeRemoteConfig(c, resp.Config)
}

func mergeRemoteConfig(c *config.Config, remoteYaml string) error {
	if strings.TrimSpace(remoteYaml) == "" {
		return nil
	}

	localMap, err := structToMap(c)
	if err != nil {
		return err
	}

	remoteMap, err := yamlToMap(remoteYaml)
	if err != nil {
		return err
	}

	merged := mergeMaps(localMap, remoteMap)
	mergedJSON, err := json.Marshal(merged)
	if err != nil {
		return err
	}

	return zconf.LoadFromJsonBytes(mergedJSON, c)
}

func structToMap(v any) (map[string]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return canonicalizeMapKeys(m), nil
}

func yamlToMap(content string) (map[string]any, error) {
	var m map[string]any
	if err := yaml.Unmarshal([]byte(content), &m); err != nil {
		return nil, err
	}

	normalized := normalizeValue(m)
	nm, ok := normalized.(map[string]any)
	if !ok {
		return map[string]any{}, nil
	}

	return canonicalizeMapKeys(nm), nil
}

func normalizeValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		for k, vv := range val {
			val[k] = normalizeValue(vv)
		}
		return val
	case map[any]any:
		out := make(map[string]any, len(val))
		for k, vv := range val {
			out[fmt.Sprint(k)] = normalizeValue(vv)
		}
		return out
	case []any:
		for i, vv := range val {
			val[i] = normalizeValue(vv)
		}
		return val
	default:
		return v
	}
}

// canonicalizeMapKeys lowercases all keys to align with go-zero config canonicalization.
// This avoids case-only duplicates (e.g. ConfigDataId vs configDataId) causing nondeterminism.
func canonicalizeMapKeys(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		lk := strings.ToLower(k)
		out[lk] = canonicalizeValue(v)
	}
	return out
}

func canonicalizeValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		return canonicalizeMapKeys(val)
	case map[any]any:
		out := make(map[string]any, len(val))
		for k, vv := range val {
			out[strings.ToLower(fmt.Sprint(k))] = canonicalizeValue(vv)
		}
		return out
	case []any:
		for i, vv := range val {
			val[i] = canonicalizeValue(vv)
		}
		return val
	default:
		return v
	}
}

func mergeMaps(dst, src map[string]any) map[string]any {
	if dst == nil {
		dst = map[string]any{}
	}

	for k, v := range src {
		dv, ok := dst[k]
		if ok {
			dm, okDst := dv.(map[string]any)
			sm, okSrc := v.(map[string]any)
			if okDst && okSrc {
				dst[k] = mergeMaps(dm, sm)
				continue
			}
		}
		dst[k] = v
	}

	return dst
}
