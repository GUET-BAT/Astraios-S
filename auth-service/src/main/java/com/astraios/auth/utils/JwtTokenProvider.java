package com.astraios.auth.utils;


import io.jsonwebtoken.Claims;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.SignatureAlgorithm;
import io.jsonwebtoken.security.Keys;
import io.jsonwebtoken.security.MacAlgorithm;
import io.jsonwebtoken.security.WeakKeyException;
import jakarta.annotation.PostConstruct;
import lombok.extern.slf4j.Slf4j;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.stereotype.Component;

import javax.crypto.SecretKey;
import javax.crypto.spec.SecretKeySpec;
import java.nio.charset.StandardCharsets;
import java.util.Base64;
import java.util.Date;

@Component
@Slf4j
public class JwtTokenProvider {
    private static final String SECRET_KEY = "Your_Secret_Key";
    private static final long EXPIRATION_TIME = 1000L * 60 * 60 * 24;  // 1天
    private static final String ISSUER = "astraios";

    private static SecretKey secretKey;
    private static final MacAlgorithm ALGORITHM = Jwts.SIG.HS512; // 或 HS256

    @PostConstruct
    public void init() {
        secretKey = buildKey(SECRET_KEY);
    }

    private static SecretKey buildKey(String secret) {
        try {

            byte[] keyBytes = secret.getBytes(StandardCharsets.UTF_8);
            // 如果长度不够，自动扩展
            if (keyBytes.length < ALGORITHM.key().build().getEncoded().length) {
                byte[] newKey = new byte[64];
                System.arraycopy(keyBytes, 0, newKey, 0, Math.min(keyBytes.length, 64));
                keyBytes = newKey;
            }
            return Keys.hmacShaKeyFor(keyBytes);
        } catch (WeakKeyException e) {
            throw new IllegalArgumentException("JWT secret key too weak. Must be at least 32 bytes for HS256 / 64 bytes for HS512", e);
        }
    }

    public static String generateToken(String uid) {
        Date now = new Date();
        Date expiration = new Date(now.getTime() + EXPIRATION_TIME);

        return Jwts.builder()
                .subject(uid)
                .issuer(ISSUER)
                .issuedAt(now)
                .expiration(expiration)
                .signWith(secretKey, ALGORITHM)
                .compact();
    }

    public static String parseTokenForUserId(String token) {
        return Jwts.parser()
                .verifyWith(secretKey)
                .build()
                .parseSignedClaims(token)
                .getPayload()
                .getSubject();
    }
}
