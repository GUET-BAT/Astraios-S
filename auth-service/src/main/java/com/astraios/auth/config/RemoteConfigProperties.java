package com.astraios.auth.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Component
@ConfigurationProperties(prefix = "auth.remote-config")
@Data
public class RemoteConfigProperties {

    private boolean enabled = true;

    private boolean failFast = true;

    private String nacosDataId = "auth-service-key";

    private long timeoutMs = 3000L;
}
