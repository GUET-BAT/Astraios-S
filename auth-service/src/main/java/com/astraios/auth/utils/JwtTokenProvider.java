package com.astraios.auth.utils;


import com.nimbusds.jose.JWSAlgorithm;
import com.nimbusds.jose.jwk.JWKSet;
import com.nimbusds.jose.jwk.KeyUse;
import com.nimbusds.jose.jwk.RSAKey;
import io.jsonwebtoken.Jwts;
import jakarta.annotation.PostConstruct;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.security.KeyPair;
import java.security.interfaces.RSAPrivateKey;
import java.security.interfaces.RSAPublicKey;
import java.util.Date;
import java.util.Map;
import java.util.UUID;

@Component
@Slf4j
public class JwtTokenProvider {

    private static final long EXPIRATION_TIME = 1000L * 60 * 60 * 24; // 1天
    private static final String ISSUER = "astraios";

    private static KeyPair keyPair;

    private static final String KEY_ID = UUID.randomUUID().toString();

    @PostConstruct
    public void init() {

        keyPair = Jwts.SIG.RS256.keyPair().build();
    }


    public static String generateToken(String uid) {
        Date now = new Date();
        Date expiration = new Date(now.getTime() + EXPIRATION_TIME);

        return Jwts.builder()
                .header().keyId(KEY_ID).and()
                .subject(uid)
                .issuer(ISSUER)
                .issuedAt(now)
                .expiration(expiration)
                .signWith(keyPair.getPrivate(), Jwts.SIG.RS256)
                .compact();
    }

    public String parseTokenForUserId(String token) {
        return Jwts.parser()
                .verifyWith(keyPair.getPublic())
                .build()
                .parseSignedClaims(token)
                .getPayload()
                .getSubject();
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
