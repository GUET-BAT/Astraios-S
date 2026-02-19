package com.astraios.auth.utils;

import com.astraios.auth.config.JwtKeyProperties;
import com.astraios.auth.config.RemoteConfigProperties;
import com.astraios.grpc.common.CommonServiceGrpc;
import com.astraios.grpc.common.LoadConfigRequest;
import com.astraios.grpc.common.LoadConfigResponse;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.grpc.StatusRuntimeException;
import io.jsonwebtoken.Claims;
import io.jsonwebtoken.ExpiredJwtException;
import io.jsonwebtoken.Jwts;
import jakarta.annotation.PostConstruct;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.client.inject.GrpcClient;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;
import org.yaml.snakeyaml.Yaml;

import java.security.KeyFactory;
import java.security.KeyPair;
import java.security.interfaces.RSAPrivateKey;
import java.security.interfaces.RSAPublicKey;
import java.security.spec.PKCS8EncodedKeySpec;
import java.security.spec.X509EncodedKeySpec;
import java.util.Base64;
import java.util.Date;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.TimeUnit;

@Component
@Slf4j
@RequiredArgsConstructor
public class JwtTokenProvider {

    private static final String ISSUER = "astraios";
    private static final String CLAIM_KEY_TYPE = "token_type";
    private static final String TYPE_ACCESS = "access";
    private static final String TYPE_REFRESH = "refresh";

    public static final long ACCESS_TOKEN_EXPIRATION = 30 * 60 * 1000;
    public static final long REFRESH_TOKEN_EXPIRATION = 7 * 24 * 60 * 60 * 1000;

    private static volatile KeyPair keyPair;
    private static String KEY_ID = UUID.randomUUID().toString();

    private final JwtKeyProperties keyProperties;
    private final RemoteConfigProperties remoteConfigProperties;
    private final ObjectMapper objectMapper;

    @GrpcClient("common-service")
    private CommonServiceGrpc.CommonServiceBlockingStub commonServiceStub;

    @PostConstruct
    public void init() {
        loadKeys();
    }

    private synchronized void loadKeys() {
        if (!remoteConfigProperties.isEnabled()) {
            loadKeysFromLocalConfig();
            return;
        }

        try {
            JwtMaterial jwtMaterial = loadJwtMaterialFromCommonService();
            keyPair = buildKeyPair(jwtMaterial.publicKey(), jwtMaterial.privateKey());
            KEY_ID = StringUtils.hasText(jwtMaterial.keyId()) ? jwtMaterial.keyId() : keyProperties.getKeyIdOrDefault();
            log.info("Loaded JWT key pair via common-service, dataId={}, kid={}", remoteConfigProperties.getNacosDataId(), KEY_ID);
        } catch (Exception e) {
            if (remoteConfigProperties.isFailFast()) {
                throw new IllegalStateException("Failed to load JWT keys via common-service", e);
            }
            log.warn("Failed to load JWT keys via common-service, fallback to local config: {}", e.getMessage());
            loadKeysFromLocalConfig();
        }
    }

    private void loadKeysFromLocalConfig() {
        try {
            keyPair = buildKeyPair(keyProperties.getPublicKey(), keyProperties.getPrivateKey());
            KEY_ID = keyProperties.getKeyIdOrDefault();
            log.info("Loaded JWT key pair from local jwt config, kid={}", KEY_ID);
        } catch (Exception e) {
            throw new IllegalStateException("Failed to load JWT keys from local jwt config", e);
        }
    }

    private JwtMaterial loadJwtMaterialFromCommonService() throws Exception {
        if (commonServiceStub == null) {
            throw new IllegalStateException("common-service grpc client is not initialized");
        }
        if (!StringUtils.hasText(remoteConfigProperties.getNacosDataId())) {
            throw new IllegalArgumentException("auth.remote-config.nacos-data-id is empty");
        }

        LoadConfigRequest request = LoadConfigRequest.newBuilder()
                .setNacosDataId(remoteConfigProperties.getNacosDataId())
                .build();
        CommonServiceGrpc.CommonServiceBlockingStub stub = commonServiceStub.withDeadlineAfter(
                remoteConfigProperties.getTimeoutMs(),
                TimeUnit.MILLISECONDS
        );

        LoadConfigResponse response;
        try {
            response = stub.loadConfig(request);
        } catch (StatusRuntimeException e) {
            throw new IllegalStateException("failed to call common-service LoadConfig", e);
        }

        if (response.getCode() != 0) {
            throw new IllegalStateException("common-service LoadConfig failed: code="
                    + response.getCode() + ", message=" + response.getMessage());
        }
        if (!StringUtils.hasText(response.getConfig())) {
            throw new IllegalStateException("common-service returned empty config");
        }
        return parseJwtMaterial(response.getConfig());
    }

    private JwtMaterial parseJwtMaterial(String configText) throws Exception {
        JsonNode root = parseConfigTree(configText);
        JsonNode jwtNode = root.path("jwt");

        String publicKey = firstText(jwtNode, "public-key", "public_key", "publicKey");
        String privateKey = firstText(jwtNode, "private-key", "private_key", "privateKey");
        String keyId = firstText(jwtNode, "key-id", "key_id", "keyId");

        if (!StringUtils.hasText(publicKey)) {
            publicKey = firstText(root, "jwt.public-key", "jwt.public_key", "jwt.publicKey");
        }
        if (!StringUtils.hasText(privateKey)) {
            privateKey = firstText(root, "jwt.private-key", "jwt.private_key", "jwt.privateKey");
        }
        if (!StringUtils.hasText(keyId)) {
            keyId = firstText(root, "jwt.key-id", "jwt.key_id", "jwt.keyId");
        }

        if (!StringUtils.hasText(publicKey) || !StringUtils.hasText(privateKey)) {
            throw new IllegalStateException("missing jwt public/private key in common-service config response");
        }
        return new JwtMaterial(publicKey, privateKey, keyId);
    }

    private JsonNode parseConfigTree(String configText) throws Exception {
        try {
            return objectMapper.readTree(configText);
        } catch (JsonProcessingException ignored) {
            Object yamlObj = new Yaml().load(configText);
            if (yamlObj == null) {
                throw new IllegalStateException("common-service returned empty yaml config");
            }
            return objectMapper.valueToTree(yamlObj);
        }
    }

    private String firstText(JsonNode node, String... names) {
        if (node == null || node.isMissingNode()) {
            return null;
        }
        for (String name : names) {
            JsonNode child = node.get(name);
            if (child != null && !child.isNull()) {
                String text = child.asText();
                if (StringUtils.hasText(text)) {
                    return text;
                }
            }
        }
        return null;
    }

    private KeyPair buildKeyPair(String publicPem, String privatePem) throws Exception {
        if (!StringUtils.hasText(publicPem) || !StringUtils.hasText(privatePem)) {
            throw new IllegalArgumentException("jwt.public-key or jwt.private-key is empty");
        }

        KeyFactory keyFactory = KeyFactory.getInstance("RSA");
        RSAPublicKey publicKey = (RSAPublicKey) keyFactory.generatePublic(new X509EncodedKeySpec(decodePem(publicPem)));
        RSAPrivateKey privateKey = (RSAPrivateKey) keyFactory.generatePrivate(new PKCS8EncodedKeySpec(decodePem(privatePem)));
        return new KeyPair(publicKey, privateKey);
    }

    private byte[] decodePem(String pem) {
        if (pem.contains("ssh-rsa ") || pem.contains("ssh-ed25519 ") || pem.contains("ecdsa-sha2-")) {
            throw new IllegalArgumentException("OpenSSH public key format is not supported. Use PEM public key.");
        }
        String sanitized = pem
                .replaceAll("-----BEGIN([\\s\\w]*)KEY-----", "")
                .replaceAll("-----END([\\s\\w]*)KEY-----", "")
                .replaceAll("\\s", "");
        return Base64.getDecoder().decode(sanitized);
    }

    public static long getRefreshTokenTtl() {
        return REFRESH_TOKEN_EXPIRATION;
    }

    public static String generateToken(String uid, long expireTime) {
        Date now = new Date();
        Date expiration = new Date(now.getTime() + expireTime);

        return Jwts.builder()
                .header().keyId(KEY_ID).and()
                .subject(uid)
                .issuer(ISSUER)
                .issuedAt(now)
                .expiration(expiration)
                .signWith(keyPair.getPrivate(), Jwts.SIG.RS256)
                .compact();
    }

    public String generateAccessToken(String userId, String username) {
        Map<String, Object> claims = new HashMap<>();
        claims.put("username", username);
        claims.put(CLAIM_KEY_TYPE, TYPE_ACCESS);

        return Jwts.builder()
                .header().keyId(KEY_ID).and()
                .claims(claims)
                .subject(userId)
                .issuer(ISSUER)
                .issuedAt(new Date())
                .expiration(new Date(System.currentTimeMillis() + ACCESS_TOKEN_EXPIRATION))
                .signWith(keyPair.getPrivate(), Jwts.SIG.RS256)
                .compact();
    }

    public String generateRefreshToken(String userId) {
        Map<String, Object> claims = new HashMap<>();
        claims.put(CLAIM_KEY_TYPE, TYPE_REFRESH);

        return Jwts.builder()
                .header().keyId(KEY_ID).and()
                .claims(claims)
                .id(UUID.randomUUID().toString())
                .subject(userId)
                .issuer(ISSUER)
                .issuedAt(new Date())
                .expiration(new Date(System.currentTimeMillis() + REFRESH_TOKEN_EXPIRATION))
                .signWith(keyPair.getPrivate(), Jwts.SIG.RS256)
                .compact();
    }

    public Claims parseToken(String token) {
        return Jwts.parser()
                .verifyWith(keyPair.getPublic())
                .build()
                .parseSignedClaims(token)
                .getPayload();
    }

    public boolean isTokenExpired(String token) {
        try {
            return parseToken(token).getExpiration().before(new Date());
        } catch (ExpiredJwtException e) {
            return true;
        }
    }

    public boolean validateAccessToken(String token) {
        try {
            Claims claims = parseToken(token);
            String type = claims.get(CLAIM_KEY_TYPE, String.class);
            return TYPE_ACCESS.equals(type);
        } catch (Exception e) {
            return false;
        }
    }

    public boolean validateRefreshToken(String token) {
        try {
            Claims claims = parseToken(token);
            String type = claims.get(CLAIM_KEY_TYPE, String.class);
            return TYPE_REFRESH.equals(type);
        } catch (Exception e) {
            return false;
        }
    }

    public Map<String, Object> getJwkSet() {
        RSAPublicKey publicKey = (RSAPublicKey) keyPair.getPublic();

        Map<String, Object> key = new HashMap<>();
        key.put("kty", "RSA");
        key.put("use", "sig");
        key.put("alg", "RS256");
        key.put("kid", KEY_ID);
        key.put("n", toBase64UrlUnsigned(publicKey.getModulus().toByteArray()));
        key.put("e", toBase64UrlUnsigned(publicKey.getPublicExponent().toByteArray()));

        return Map.of("keys", List.of(key));
    }

    private String toBase64UrlUnsigned(byte[] bytes) {
        int start = 0;
        while (start < bytes.length - 1 && bytes[start] == 0) {
            start++;
        }
        return Base64.getUrlEncoder().withoutPadding().encodeToString(
                start == 0 ? bytes : java.util.Arrays.copyOfRange(bytes, start, bytes.length)
        );
    }

    private record JwtMaterial(String publicKey, String privateKey, String keyId) {
    }
}
