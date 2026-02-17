package com.astraios.auth.config;

import com.astraios.grpc.common.CommonServiceGrpc;
import com.astraios.grpc.common.LoadConfigRequest;
import com.astraios.grpc.common.LoadConfigResponse;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.grpc.StatusRuntimeException;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import org.yaml.snakeyaml.Yaml;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.TimeUnit;

@Component
@Slf4j
@RequiredArgsConstructor
public class RemoteConfigLoader {

    private final RemoteConfigProperties remoteConfigProperties;
    private final ObjectMapper objectMapper;

    private CommonServiceGrpc.CommonServiceBlockingStub commonServiceStub;

    public void setCommonServiceStub(CommonServiceGrpc.CommonServiceBlockingStub stub) {
        this.commonServiceStub = stub;
    }

    public Map<String, Object> loadRedisConfig() {
        if (!remoteConfigProperties.isEnabled()) {
            log.info("Remote config is disabled, skipping Redis config load");
            return Map.of();
        }

        String dataId = remoteConfigProperties.getRedisDataId();
        if (!StringUtils.hasText(dataId)) {
            log.warn("Redis dataId is not configured");
            return Map.of();
        }

        try {
            String configText = loadConfigFromCommonService(dataId);
            Map<String, Object> redisConfig = parseRedisConfig(configText);
            log.info("Loaded Redis config from common-service, dataId={}", dataId);
            return redisConfig;
        } catch (Exception e) {
            if (remoteConfigProperties.isFailFast()) {
                throw new IllegalStateException("Failed to load Redis config from common-service", e);
            }
            log.warn("Failed to load Redis config from common-service: {}", e.getMessage());
            return Map.of();
        }
    }

    public String loadConfigFromCommonService(String dataId) {
        if (commonServiceStub == null) {
            throw new IllegalStateException("common-service gRPC client is not initialized");
        }

        LoadConfigRequest request = LoadConfigRequest.newBuilder()
                .setNacosDataId(dataId)
                .build();

        CommonServiceGrpc.CommonServiceBlockingStub stub = commonServiceStub.withDeadlineAfter(
                remoteConfigProperties.getTimeoutMs(),
                TimeUnit.MILLISECONDS
        );

        LoadConfigResponse response;
        try {
            response = stub.loadConfig(request);
        } catch (StatusRuntimeException e) {
            throw new IllegalStateException("Failed to call common-service LoadConfig", e);
        }

        if (response.getCode() != 0) {
            throw new IllegalStateException("common-service LoadConfig failed: code="
                    + response.getCode() + ", message=" + response.getMessage());
        }

        if (!StringUtils.hasText(response.getConfig())) {
            throw new IllegalStateException("common-service returned empty config");
        }

        return response.getConfig();
    }

    @SuppressWarnings("unchecked")
    private Map<String, Object> parseRedisConfig(String configText) throws Exception {
        JsonNode root = parseConfigTree(configText);
        JsonNode springNode = root.path("spring");
        JsonNode dataNode = springNode.path("data");
        JsonNode redisNode = dataNode.path("redis");

        if (redisNode.isMissingNode()) {
            log.warn("Redis config not found in remote config");
            return Map.of();
        }

        Map<String, Object> redisConfig = new HashMap<>();

        if (redisNode.has("host")) {
            redisConfig.put("host", redisNode.get("host").asText());
        }
        if (redisNode.has("port")) {
            redisConfig.put("port", redisNode.get("port").asInt());
        }
        if (redisNode.has("password")) {
            redisConfig.put("password", redisNode.get("password").asText());
        }
        if (redisNode.has("database")) {
            redisConfig.put("database", redisNode.get("database").asInt());
        }

        return redisConfig;
    }

    private JsonNode parseConfigTree(String configText) throws Exception {
        try {
            return objectMapper.readTree(configText);
        } catch (JsonProcessingException ignored) {
            Object yamlObj = new Yaml().load(configText);
            if (yamlObj == null) {
                throw new IllegalStateException("Empty YAML config");
            }
            return objectMapper.valueToTree(yamlObj);
        }
    }
}
