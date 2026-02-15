package com.astraios.auth.config;

import com.astraios.grpc.common.CommonServiceGrpc;
import io.lettuce.core.ClientOptions;
import io.lettuce.core.SocketOptions;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.client.inject.GrpcClient;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.boot.autoconfigure.data.redis.RedisProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;
import org.springframework.data.redis.connection.RedisConnectionFactory;
import org.springframework.data.redis.connection.RedisStandaloneConfiguration;
import org.springframework.data.redis.connection.lettuce.LettuceClientConfiguration;
import org.springframework.data.redis.connection.lettuce.LettuceConnectionFactory;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.data.redis.serializer.StringRedisSerializer;

import java.time.Duration;
import java.util.Map;

@Configuration
@Slf4j
@RequiredArgsConstructor
public class RedisConfig {

    private final RemoteConfigLoader remoteConfigLoader;
    private final RemoteConfigProperties remoteConfigProperties;
    private final RedisProperties redisProperties;

    @GrpcClient("common-service")
    private CommonServiceGrpc.CommonServiceBlockingStub commonServiceStub;

    @Bean
    @Primary
    public RedisConnectionFactory redisConnectionFactory() {
        remoteConfigLoader.setCommonServiceStub(commonServiceStub);

        String host = redisProperties.getHost();
        int port = redisProperties.getPort();
        String password = redisProperties.getPassword();
        int database = redisProperties.getDatabase();

        if (remoteConfigProperties.isEnabled()) {
            try {
                Map<String, Object> remoteRedisConfig = remoteConfigLoader.loadRedisConfig();
                if (!remoteRedisConfig.isEmpty()) {
                    if (remoteRedisConfig.containsKey("host")) {
                        host = (String) remoteRedisConfig.get("host");
                    }
                    if (remoteRedisConfig.containsKey("port")) {
                        port = ((Number) remoteRedisConfig.get("port")).intValue();
                    }
                    if (remoteRedisConfig.containsKey("password")) {
                        password = (String) remoteRedisConfig.get("password");
                    }
                    if (remoteRedisConfig.containsKey("database")) {
                        database = ((Number) remoteRedisConfig.get("database")).intValue();
                    }
                    log.info("Using Redis config from remote: host={}, port={}, database={}", host, port, database);
                }
            } catch (Exception e) {
                log.warn("Failed to load remote Redis config, using local config: {}", e.getMessage());
            }
        }

        RedisStandaloneConfiguration config = new RedisStandaloneConfiguration();
        config.setHostName(host);
        config.setPort(port);
        config.setDatabase(database);
        if (password != null && !password.isEmpty()) {
            config.setPassword(password);
        }

        SocketOptions socketOptions = SocketOptions.builder()
                .connectTimeout(Duration.ofSeconds(5))
                .build();

        ClientOptions clientOptions = ClientOptions.builder()
                .socketOptions(socketOptions)
                .autoReconnect(true)
                .build();

        LettuceClientConfiguration clientConfig = LettuceClientConfiguration.builder()
                .commandTimeout(Duration.ofSeconds(5))
                .clientOptions(clientOptions)
                .build();

        return new LettuceConnectionFactory(config, clientConfig);
    }

    @Bean
    @ConditionalOnMissingBean
    public RedisTemplate<String, Object> redisTemplate(RedisConnectionFactory connectionFactory) {
        RedisTemplate<String, Object> template = new RedisTemplate<>();
        template.setConnectionFactory(connectionFactory);
        template.setKeySerializer(new StringRedisSerializer());
        template.setValueSerializer(new StringRedisSerializer());
        template.setHashKeySerializer(new StringRedisSerializer());
        template.setHashValueSerializer(new StringRedisSerializer());
        template.afterPropertiesSet();
        return template;
    }
}
