package com.astraios.auth.utils;


import com.nimbusds.jose.JWSAlgorithm;
import com.nimbusds.jose.jwk.JWKSet;
import com.nimbusds.jose.jwk.KeyUse;
import com.nimbusds.jose.jwk.RSAKey;
import io.jsonwebtoken.Claims;
import io.jsonwebtoken.ExpiredJwtException;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.MalformedJwtException;
import jakarta.annotation.PostConstruct;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.security.KeyPair;
import java.security.interfaces.RSAPrivateKey;
import java.security.interfaces.RSAPublicKey;
import java.util.Date;
import java.util.Map;
import java.util.UUID;
import java.util.HashMap;

@Component
@Slf4j
public class JwtTokenProvider {

    private static final String ISSUER = "astraios";

    private static KeyPair keyPair;

    private static final String KEY_ID = UUID.randomUUID().toString();

    public static final long ACCESS_TOKEN_EXPIRATION = 15 * 60 * 1000;
    public static final long REFRESH_TOKEN_EXPIRATION = 7 * 24 * 60 * 60 * 1000;

    private static final String CLAIM_KEY_TYPE = "token_type";
    private static final String TYPE_ACCESS = "access";
    private static final String TYPE_REFRESH = "refresh";

    @PostConstruct
    public void init() {
        keyPair = Jwts.SIG.RS256.keyPair().build();
    }

    public static long getRefreshTokenTtl(){
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
    // 生成 Access Token
    public String generateAccessToken(String userId, String username) {
        Map<String, Object> claims = new HashMap<>();
        claims.put("username", username);
        claims.put(CLAIM_KEY_TYPE, TYPE_ACCESS); // 标记为 Access Token
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

    // 生成 Refresh Token
    public String generateRefreshToken(String userId) {
        Map<String, Object> claims = new HashMap<>();
        claims.put(CLAIM_KEY_TYPE, TYPE_REFRESH); // 标记为 Access Token
        return Jwts.builder()
                .header().keyId(KEY_ID).and()
                .claims(claims)
                .id(UUID.randomUUID().toString())
                .subject(userId)
                .issuer(ISSUER)
                .issuedAt(new Date())
                .expiration(new Date(System.currentTimeMillis() + ACCESS_TOKEN_EXPIRATION))
                .signWith(keyPair.getPrivate(), Jwts.SIG.RS256)
                .compact();
    }

    // 解析 Token，同时校验签名和过期时间
    public Claims parseToken(String token) {
        return Jwts.parser()
                .verifyWith(keyPair.getPublic())
                .build()
                .parseSignedClaims(token)// 正式校验签名和过期时间
                .getPayload();
    }

    // 校验 Token 是否过期
    public boolean isTokenExpired(String token) {
        try {
            return parseToken(token).getExpiration().before(new Date());
        } catch (ExpiredJwtException e) {
            return true;
        }
    }

    /**
     * 校验 Access Token 是否有效
     */
    public boolean validateAccessToken(String token) {
        try {
            Claims claims = parseToken(token);
            String type = claims.get(CLAIM_KEY_TYPE, String.class);
            return TYPE_ACCESS.equals(type);
        } catch (Exception e) {
            return false;
        }
    }

    /**
     * 校验 Refresh Token 是否基本有效
     */
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

        RSAKey key = new RSAKey.Builder(publicKey)
                .keyUse(KeyUse.SIGNATURE)
                .algorithm(JWSAlgorithm.RS256)
                .keyID(KEY_ID)
                .build();

        return new JWKSet(key).toJSONObject();
    }
}
