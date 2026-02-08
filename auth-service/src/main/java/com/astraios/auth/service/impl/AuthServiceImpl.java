package com.astraios.auth.service.impl;

import com.astraios.auth.contants.AuthConstants;
import com.astraios.auth.domain.dto.LoginRequest;
import com.astraios.auth.domain.dto.RefreshRequest;
import com.astraios.auth.domain.dto.RegisterRequest;
import com.astraios.auth.domain.vo.LoginResult;
import com.astraios.auth.domain.vo.RefreshResult;
import com.astraios.auth.domain.vo.RegisterResult;
import com.astraios.auth.exception.GrpcStatusException;
import com.astraios.auth.service.AuthService;
import com.astraios.auth.utils.JwtTokenProvider;
import com.astraios.grpc.user.UserDataRequest;
import com.astraios.grpc.user.UserDataResponse;
import com.astraios.grpc.user.UserServiceGrpc;
import com.astraios.grpc.user.VerifyPasswordRequest;
import com.astraios.grpc.user.VerifyPasswordResponse;
import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import io.jsonwebtoken.Claims;
import lombok.RequiredArgsConstructor;
import net.devh.boot.grpc.client.inject.GrpcClient;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import java.util.concurrent.TimeUnit;


@Service
@RequiredArgsConstructor
public class AuthServiceImpl implements AuthService {

    private final JwtTokenProvider jwtTokenProvider;

    private final StringRedisTemplate redisTemplate;

    @GrpcClient("user-service")
    private UserServiceGrpc.UserServiceBlockingStub userServiceStub;

    @Override
    public LoginResult login(LoginRequest request) {
        // 1.校验请求
        if (!StringUtils.hasText(request.getUsername()) || !StringUtils.hasText(request.getPassword())){
            throw new GrpcStatusException(Status.INVALID_ARGUMENT, "Invalid username or password");
        }

        // 2. 调用 user-service 校验账号密码
        VerifyPasswordRequest rpcRequest = VerifyPasswordRequest.newBuilder()
                .setUsername(request.getUsername())
                .setPassword(request.getPassword())
                .build();

        VerifyPasswordResponse rpcResponse;
        try {
            rpcResponse = userServiceStub.verifyPassword(rpcRequest);
        } catch (StatusRuntimeException e) {
            throw new GrpcStatusException(Status.UNAVAILABLE, "用户服务不可用", e);
        } catch (Exception e) {
            throw new GrpcStatusException(Status.UNAVAILABLE, "用户服务不可用", e);
        }

        if (rpcResponse.getCode() == 0) {
            throw new GrpcStatusException(Status.UNAUTHENTICATED, "Invalid username or password");
        }

        String userId = rpcResponse.getUserId();

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
        LoginResult loginResult = new LoginResult();
        loginResult.setRefreshToken(refreshToken);
        loginResult.setAccessToken(accessToken);
        return loginResult;
    }

    @Override
    public RegisterResult register(RegisterRequest request) {
        // 1.校验请求
        if (!StringUtils.hasText(request.getUsername()) || !StringUtils.hasText(request.getPassword())){
            throw new GrpcStatusException(Status.INVALID_ARGUMENT, "Invalid username or password");
        }
        
        // 2. RPC调用user服务注册
        com.astraios.grpc.user.RegisterRequest rpcRequest = com.astraios.grpc.user.RegisterRequest.newBuilder()
                .setUsername(request.getUsername())
                .setPassword(request.getPassword())
                .build();
        com.astraios.grpc.user.RegisterResponse rpcResponse;
        try {
            rpcResponse = userServiceStub.register(rpcRequest);
        } catch (StatusRuntimeException e) {
            throw new GrpcStatusException(Status.UNAVAILABLE, "用户服务不可用", e);
        } catch (Exception e) {
            throw new GrpcStatusException(Status.UNAVAILABLE, "用户服务不可用", e);
        }
        
        RegisterResult registerResult = new RegisterResult();
        if(rpcResponse.getCode() == 0){
            return registerResult;
        }
        throw new GrpcStatusException(Status.INTERNAL, "register failed");
    }

    @Override
    public RefreshResult refreshToken(RefreshRequest request) {
        // 1. 校验格式基本合法性 (签名校验)
        Claims claims;
        try {
            claims = jwtTokenProvider.parseToken(request.getRefreshToken());
        } catch (Exception e) {
            throw new GrpcStatusException(Status.UNAUTHENTICATED, "Invalid Refresh Token", e);
        }

        String userId = claims.getSubject();
        String redisKey = AuthConstants.REDIS_REFRESH_TOKEN_PREFIX + userId;

        // 2. 校验Redis 中是否存在该 Token
        String storedRefreshToken = redisTemplate.opsForValue().get(redisKey);

        if (storedRefreshToken == null || !storedRefreshToken.equals(request.getRefreshToken())) {
            throw new GrpcStatusException(Status.UNAUTHENTICATED, "Refresh Token expired or invalid");
        }

        // 3. 生成新的 Access Token
        UserDataRequest rpcRequest = UserDataRequest.newBuilder().setUserId(userId).build();
        UserDataResponse  rpcResponse;
        try {
            rpcResponse = userServiceStub.getUserData(rpcRequest);
        } catch (StatusRuntimeException e) {
            throw new GrpcStatusException(Status.UNAVAILABLE, "用户服务不可用", e);
        } catch (Exception e) {
            throw new GrpcStatusException(Status.UNAVAILABLE, "用户服务不可用", e);
        }
        String newAccessToken = jwtTokenProvider.generateAccessToken(userId, rpcResponse.getNickname());

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
        return result;
    }
}
