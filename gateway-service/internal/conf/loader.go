package conf

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	commonpb "github.com/GUET-BAT/Astraios-S/common-service/pb/commonpb"
	"github.com/GUET-BAT/Astraios-S/gateway-service/internal/config"

	zconf "github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/mapping"
	"github.com/zeromicro/go-zero/zrpc"
	"gopkg.in/yaml.v3"
)

const remoteTimeout = 5 * time.Second

// MustLoad loads local config first, then fetches runtime config from common-service.
func MustLoad(path string, c *config.Config) {
	zconf.MustLoad(path, c)
	logx.Must(loadRemoteConfig(path, c))
}

func loadRemoteConfig(path string, c *config.Config) error {
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

	return mergeRemoteConfig(path, c, resp.Config)
}

func mergeRemoteConfig(path string, c *config.Config, remoteYaml string) error {
	if strings.TrimSpace(remoteYaml) == "" {
		return nil
	}

	localMap, err := loadLocalConfigMap(path)
	if err != nil {
		localMap, err = structToMap(c)
		if err != nil {
			return err
		}
	}

	remoteMap, err := yamlToMap(remoteYaml)
	if err != nil {
		return err
	}

	info, err := getConfigFieldInfo()
	if err != nil {
		return err
	}

	localMap = canonicalizeWithInfo(localMap, info)
	remoteMap = canonicalizeWithInfo(remoteMap, info)
	remoteMap = pruneEmptyStringsMap(remoteMap)

	merged := mergeMaps(localMap, remoteMap)
	mergedJSON, err := json.Marshal(merged)
	if err != nil {
		return err
	}

	return zconf.LoadFromJsonBytes(mergedJSON, c)
}

func loadLocalConfigMap(path string) (map[string]any, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(string(content)) == "" {
		return map[string]any{}, nil
	}

	return yamlToMap(string(content))
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

	return m, nil
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

	return nm, nil
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

// pruneEmptyStringsMap removes keys whose value is an empty string (or becomes empty after pruning).
// This prevents remote empty values from overriding local defaults.
func pruneEmptyStringsMap(m map[string]any) map[string]any {
	if m == nil {
		return m
	}

	for k, v := range m {
		nv, keep := pruneEmptyStringsValue(v)
		if !keep {
			delete(m, k)
			continue
		}
		m[k] = nv
	}

	return m
}

func pruneEmptyStringsValue(v any) (any, bool) {
	switch vv := v.(type) {
	case string:
		if strings.TrimSpace(vv) == "" {
			return nil, false
		}
		return vv, true
	case map[string]any:
		pruned := pruneEmptyStringsMap(vv)
		return pruned, true
	case map[any]any:
		converted := make(map[string]any, len(vv))
		for k, vvv := range vv {
			converted[fmt.Sprint(k)] = vvv
		}
		pruned := pruneEmptyStringsMap(converted)
		return pruned, true
	case []any:
		out := make([]any, 0, len(vv))
		for _, item := range vv {
			nv, keep := pruneEmptyStringsValue(item)
			if keep {
				out = append(out, nv)
			}
		}
		if len(out) == 0 {
			return nil, false
		}
		return out, true
	default:
		return v, true
	}
}

type fieldInfo struct {
	children map[string]*fieldInfo
	mapField *fieldInfo
}

var (
	configFieldInfo     *fieldInfo
	configFieldInfoOnce sync.Once
	configFieldInfoErr  error
)

func getConfigFieldInfo() (*fieldInfo, error) {
	configFieldInfoOnce.Do(func() {
		configFieldInfo, configFieldInfoErr = buildFieldsInfo(reflect.TypeOf(config.Config{}), "")
	})
	return configFieldInfo, configFieldInfoErr
}

func canonicalizeWithInfo(m map[string]any, info *fieldInfo) map[string]any {
	if info == nil {
		return m
	}

	res := make(map[string]any, len(m))
	for k, v := range m {
		if ti, ok := info.children[k]; ok {
			res[k] = canonicalizeValueWithInfo(v, ti)
			continue
		}

		lk := strings.ToLower(k)
		if ti, ok := info.children[lk]; ok {
			res[lk] = canonicalizeValueWithInfo(v, ti)
		} else if info.mapField != nil {
			res[k] = canonicalizeValueWithInfo(v, info.mapField)
		} else if vv, ok := v.(map[string]any); ok {
			res[k] = canonicalizeWithInfo(vv, info)
		} else {
			res[k] = v
		}
	}

	return res
}

func canonicalizeValueWithInfo(v any, info *fieldInfo) any {
	switch vv := v.(type) {
	case map[string]any:
		return canonicalizeWithInfo(vv, info)
	case map[any]any:
		m := make(map[string]any, len(vv))
		for k, vvv := range vv {
			m[fmt.Sprint(k)] = vvv
		}
		return canonicalizeWithInfo(m, info)
	case []any:
		arr := make([]any, 0, len(vv))
		for _, vvv := range vv {
			arr = append(arr, canonicalizeValueWithInfo(vvv, info))
		}
		return arr
	default:
		return v
	}
}

func buildFieldsInfo(tp reflect.Type, fullName string) (*fieldInfo, error) {
	tp = mapping.Deref(tp)

	switch tp.Kind() {
	case reflect.Struct:
		return buildStructFieldsInfo(tp, fullName)
	case reflect.Array, reflect.Slice, reflect.Map:
		return buildFieldsInfo(mapping.Deref(tp.Elem()), fullName)
	default:
		return &fieldInfo{
			children: make(map[string]*fieldInfo),
		}, nil
	}
}

func buildStructFieldsInfo(tp reflect.Type, fullName string) (*fieldInfo, error) {
	info := &fieldInfo{
		children: make(map[string]*fieldInfo),
	}

	for i := 0; i < tp.NumField(); i++ {
		field := tp.Field(i)
		name := field.Name
		tag := strings.TrimSpace(field.Tag.Get("json"))
		if tag == "-" {
			continue
		}
		if tag != "" {
			parts := strings.Split(tag, ",")
			if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
				name = parts[0]
			}
		}
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		child, err := buildFieldsInfo(field.Type, fullName+"."+name)
		if err != nil {
			return nil, err
		}
		info.children[name] = child
		info.children[strings.ToLower(name)] = child

		if field.Type.Kind() == reflect.Map {
			info.mapField = child
		}
	}

	return info, nil
}
