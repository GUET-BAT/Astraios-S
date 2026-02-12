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

    private static final String USER_SERVICE_UNAVAILABLE = "user-service is unavailable";

    private final JwtTokenProvider jwtTokenProvider;
    private final StringRedisTemplate redisTemplate;

    @GrpcClient("user-service")
    private UserServiceGrpc.UserServiceBlockingStub userServiceStub;

    @Override
    public LoginResult login(LoginRequest request) {
        validateCredentials(request.getUsername(), request.getPassword());

        VerifyPasswordRequest rpcRequest = VerifyPasswordRequest.newBuilder()
                .setUsername(request.getUsername())
                .setPassword(request.getPassword())
                .build();

        VerifyPasswordResponse rpcResponse = verifyPassword(rpcRequest);
        if (rpcResponse.getCode() == 0) {
            throw new GrpcStatusException(Status.UNAUTHENTICATED, "Invalid username or password");
        }

        String userId = rpcResponse.getUserId();
        String accessToken = jwtTokenProvider.generateAccessToken(userId, request.getUsername());
        String refreshToken = jwtTokenProvider.generateRefreshToken(userId);

        String redisKey = AuthConstants.REDIS_REFRESH_TOKEN_PREFIX + userId;
        redisTemplate.opsForValue().set(
                redisKey,
                refreshToken,
                JwtTokenProvider.REFRESH_TOKEN_EXPIRATION,
                TimeUnit.MILLISECONDS
        );

        LoginResult loginResult = new LoginResult();
        loginResult.setAccessToken(accessToken);
        loginResult.setRefreshToken(refreshToken);
        return loginResult;
    }

    @Override
    public RegisterResult register(RegisterRequest request) {
        validateCredentials(request.getUsername(), request.getPassword());

        com.astraios.grpc.user.RegisterRequest rpcRequest = com.astraios.grpc.user.RegisterRequest.newBuilder()
                .setUsername(request.getUsername())
                .setPassword(request.getPassword())
                .build();

        com.astraios.grpc.user.RegisterResponse rpcResponse = registerUser(rpcRequest);
        RegisterResult registerResult = new RegisterResult();

        if (rpcResponse.getCode() == 0) {
            return registerResult;
        }
        throw new GrpcStatusException(Status.INTERNAL, "register failed");
    }

    @Override
    public RefreshResult refreshToken(RefreshRequest request) {
        Claims claims;
        try {
            claims = jwtTokenProvider.parseToken(request.getRefreshToken());
        } catch (Exception e) {
            throw new GrpcStatusException(Status.UNAUTHENTICATED, "Invalid refresh token", e);
        }

        String userId = claims.getSubject();
        String redisKey = AuthConstants.REDIS_REFRESH_TOKEN_PREFIX + userId;
        String storedRefreshToken = redisTemplate.opsForValue().get(redisKey);

        if (!StringUtils.hasText(storedRefreshToken) || !storedRefreshToken.equals(request.getRefreshToken())) {
            throw new GrpcStatusException(Status.UNAUTHENTICATED, "Refresh token expired or invalid");
        }

        UserDataResponse userData = getUserData(UserDataRequest.newBuilder().setUserId(userId).build());
        String newAccessToken = jwtTokenProvider.generateAccessToken(userId, userData.getNickname());
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

    private void validateCredentials(String username, String password) {
        if (!StringUtils.hasText(username) || !StringUtils.hasText(password)) {
            throw new GrpcStatusException(Status.INVALID_ARGUMENT, "Invalid username or password");
        }
    }

    private VerifyPasswordResponse verifyPassword(VerifyPasswordRequest request) {
        try {
            return userServiceStub.verifyPassword(request);
        } catch (StatusRuntimeException e) {
            throw new GrpcStatusException(Status.UNAVAILABLE, USER_SERVICE_UNAVAILABLE, e);
        } catch (Exception e) {
            throw new GrpcStatusException(Status.UNAVAILABLE, USER_SERVICE_UNAVAILABLE, e);
        }
    }

    private com.astraios.grpc.user.RegisterResponse registerUser(com.astraios.grpc.user.RegisterRequest request) {
        try {
            return userServiceStub.register(request);
        } catch (StatusRuntimeException e) {
            throw new GrpcStatusException(Status.UNAVAILABLE, USER_SERVICE_UNAVAILABLE, e);
        } catch (Exception e) {
            throw new GrpcStatusException(Status.UNAVAILABLE, USER_SERVICE_UNAVAILABLE, e);
        }
    }

    private UserDataResponse getUserData(UserDataRequest request) {
        try {
            return userServiceStub.getUserData(request);
        } catch (StatusRuntimeException e) {
            throw new GrpcStatusException(Status.UNAVAILABLE, USER_SERVICE_UNAVAILABLE, e);
        } catch (Exception e) {
            throw new GrpcStatusException(Status.UNAVAILABLE, USER_SERVICE_UNAVAILABLE, e);
        }
    }
}
