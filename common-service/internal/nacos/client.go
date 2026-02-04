package nacos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	envNacosServerAddr  = "NACOS_SERVER_ADDR"
	envNacosUsername    = "NACOS_USERNAME"
	envNacosPassword    = "NACOS_PASSWORD"
	envNacosNamespace   = "NACOS_NAMESPACE"
	envNacosGroup       = "NACOS_GROUP"
	envNacosDataIdSuffix = "NACOS_DATA_ID_SUFFIX"

	defaultGroup          = "DEFAULT_GROUP"
	defaultDataIdSuffix   = ".yaml"
	defaultRequestTimeout = 5 * time.Second
)

// Client 封装了访问 Nacos 配置中心所需的参数与调用逻辑。
// 所有连接信息通过环境变量注入（通常由 K8s Secret 提供）。
type Client struct {
	baseURL      string
	username     string
	password     string
	namespace    string
	group        string
	dataIdSuffix string
	httpClient   *http.Client
}

func NewClientFromEnv() (*Client, error) {
	addr := strings.TrimSpace(os.Getenv(envNacosServerAddr))
	if addr == "" {
		return nil, fmt.Errorf("%s is required", envNacosServerAddr)
	}
	baseURL := addr
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}

	group := strings.TrimSpace(os.Getenv(envNacosGroup))
	if group == "" {
		group = defaultGroup
	}

	dataIdSuffix := strings.TrimSpace(os.Getenv(envNacosDataIdSuffix))
	if dataIdSuffix == "" {
		dataIdSuffix = defaultDataIdSuffix
	}

	return &Client{
		baseURL:      strings.TrimRight(baseURL, "/"),
		username:     strings.TrimSpace(os.Getenv(envNacosUsername)),
		password:     strings.TrimSpace(os.Getenv(envNacosPassword)),
		namespace:    strings.TrimSpace(os.Getenv(envNacosNamespace)),
		group:        group,
		dataIdSuffix: dataIdSuffix,
		httpClient: &http.Client{
			Timeout: defaultRequestTimeout,
		},
	}, nil
}

// LoadConfig 根据 nacosDataId 读取配置，若无后缀则追加默认后缀。
func (c *Client) LoadConfig(ctx context.Context, nacosDataId string) (string, error) {
	if strings.TrimSpace(nacosDataId) == "" {
		return "", errors.New("nacosDataId is empty")
	}

	dataId := nacosDataId
	if !strings.HasSuffix(dataId, c.dataIdSuffix) {
		dataId += c.dataIdSuffix
	}

	token, err := c.login(ctx)
	if err != nil {
		return "", err
	}

	query := url.Values{}
	query.Set("dataId", dataId)
	query.Set("group", c.group)
	if c.namespace != "" {
		query.Set("tenant", c.namespace)
	}
	if token != "" {
		query.Set("accessToken", token)
	}

	endpoint := fmt.Sprintf("%s/nacos/v1/cs/configs?%s", c.baseURL, query.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("nacos config request failed: %s", strings.TrimSpace(string(body)))
	}

	return string(body), nil
}

// login 使用用户名/密码获取 accessToken。
// 如果未配置用户名或密码，则认为 Nacos 认证关闭，直接跳过。
func (c *Client) login(ctx context.Context) (string, error) {
	if c.username == "" && c.password == "" {
		return "", nil
	}
	if c.username == "" || c.password == "" {
		return "", errors.New("nacos username/password must both be set")
	}

	form := url.Values{}
	form.Set("username", c.username)
	form.Set("password", c.password)

	endpoint := fmt.Sprintf("%s/nacos/v1/auth/login", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("nacos login failed: %s", strings.TrimSpace(string(body)))
	}

	var payload struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("parse nacos login response: %w", err)
	}
	if payload.AccessToken == "" {
		return "", errors.New("nacos login response missing accessToken")
	}
	return payload.AccessToken, nil
}
