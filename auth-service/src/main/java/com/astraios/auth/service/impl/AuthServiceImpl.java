package com.astraios.auth.service.impl;

import com.astraios.auth.contants.AuthConstants;
import com.astraios.auth.domain.dto.LoginRequest;
import com.astraios.auth.domain.dto.RefreshRequest;
import com.astraios.auth.domain.vo.LoginResult;
import com.astraios.auth.domain.dto.RegisterRequest;
import com.astraios.auth.domain.vo.RefreshResult;
import com.astraios.auth.domain.vo.RegisterResult;
import com.astraios.auth.service.AuthService;
import com.astraios.auth.utils.JwtTokenProvider;
import com.nimbusds.oauth2.sdk.token.RefreshToken;
import io.jsonwebtoken.Claims;
import lombok.RequiredArgsConstructor;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.TimeUnit;


@Service
@RequiredArgsConstructor
public class AuthServiceImpl implements AuthService {

    private final AuthenticationManager authenticationManager;
    private final PasswordEncoder passwordEncoder;

    private static final long TOKEN_EXPIRATION_TIME = 1000L * 60 * 60; // 1小时
    private static final long TOKEN_REFRESH_TIME = 1000L * 60 * 60 * 24 * 7; // 7天
    private final JwtTokenProvider jwtTokenProvider;

    private final StringRedisTemplate redisTemplate;

    @Override
    public ResponseEntity<?> login(LoginRequest request) {
        try {
            LoginResult loginResult = new LoginResult();
            // 1.校验请求
            if (!StringUtils.hasText(request.getUsername()) || !StringUtils.hasText(request.getPassword())){
                loginResult.setMsg("Invalid username or password");
                loginResult.setStatus(HttpStatus.UNAUTHORIZED.value());
                return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(loginResult);
            }

            // 2. 封装认证请求
            Authentication requestAuth = new UsernamePasswordAuthenticationToken(request.getUsername(), request.getPassword());

            // 3.执行认证
            Authentication auth = authenticationManager.authenticate(requestAuth);

            // 4.获取认证结果
            String userId = (String) auth.getPrincipal();  //需获取的是userId

            // 5.生成token
            String accessToken =  jwtTokenProvider.generateAccessToken(userId, request.getUsername());
            String refreshToken = jwtTokenProvider.generateRefreshToken(userId);

            // 6. 将 Refresh Token 存入 Redis，设置过期时间
            String redisKey = AuthConstants.REDIS_REFRESH_TOKEN_PREFIX  + userId;
            redisTemplate.opsForValue().set(
                    redisKey,
                    refreshToken,
                    JwtTokenProvider.REFRESH_TOKEN_EXPIRATION,
                    TimeUnit.MILLISECONDS
            );

            //7.返回结果
            loginResult.setRefreshToken(refreshToken);
            loginResult.setAccessToken(accessToken);
            loginResult.setStatus(HttpStatus.OK.value());
            loginResult.setMsg("Successfully logged in");
            return ResponseEntity.ok(loginResult);
        }
        catch (Exception e) {
            LoginResult loginResult = new LoginResult();
            loginResult.setMsg(e.getMessage());
            loginResult.setStatus(HttpStatus.UNAUTHORIZED.value());
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(loginResult);
        }
    }

    @Override
    public ResponseEntity<RegisterResult> register(RegisterRequest request) {
        RegisterResult registerResult = new RegisterResult();
        registerResult.setMsg("Invalid username or password");
        registerResult.setStatus(HttpStatus.UNAUTHORIZED.value());
        try {
            // 1.校验请求
            if (!StringUtils.hasText(request.getUsername()) || !StringUtils.hasText(request.getPassword())){

                return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(registerResult);
            }


            // 2. TODO RPC调用user服务注册
            registerResult.setStatus(HttpStatus.OK.value());
            registerResult.setMsg("success");
            return ResponseEntity.ok(registerResult);

        }
        catch (Exception e) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(registerResult);
        }
    }

    public ResponseEntity<?> refreshToken(RefreshRequest request) {
        // 1. 校验格式基本合法性 (签名校验)
        Claims claims;
        try {
            claims = jwtTokenProvider.parseToken(request.getRefreshToken());
        } catch (Exception e) {
            throw new RuntimeException("Invalid Refresh Token");
        }

        String userId = claims.getSubject();
        String redisKey = AuthConstants.REDIS_REFRESH_TOKEN_PREFIX + userId;

        // 2. 校验Redis 中是否存在该 Token
        String storedRefreshToken = redisTemplate.opsForValue().get(redisKey);

        if (storedRefreshToken == null || !storedRefreshToken.equals(request.getRefreshToken())) {
            throw new RuntimeException("Refresh Token expired or invalid");
        }

        // 3. 生成新的 Access Token
        // TODO gRPC调用获取用户名
        String username = "123";
        String newAccessToken = jwtTokenProvider.generateAccessToken(userId, username);

        // 4. refreshToken轮换，再重新生成refreshToken
        String newRefreshToken = jwtTokenProvider.generateRefreshToken(userId);
        redisTemplate.opsForValue().set(
                redisKey,
                newRefreshToken,
                JwtTokenProvider.REFRESH_TOKEN_EXPIRATION,
                TimeUnit.MILLISECONDS
        );

        RefreshResult result = new RefreshResult();
        result.setAccessToken(newAccessToken);
        result.setRefreshToken(newRefreshToken);
        result.setStatus(HttpStatus.OK.value());
        return ResponseEntity.ok(result);
    }
}
