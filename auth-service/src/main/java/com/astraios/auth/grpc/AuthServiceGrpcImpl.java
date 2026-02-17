package com.astraios.auth.grpc;

import com.astraios.auth.domain.dto.LoginRequest;
import com.astraios.auth.domain.dto.RefreshRequest;
import com.astraios.auth.domain.dto.RegisterRequest;
import com.astraios.auth.domain.vo.LoginResult;
import com.astraios.auth.domain.vo.RefreshResult;
import com.astraios.auth.service.AuthService;
import com.astraios.auth.utils.JwtTokenProvider;
import com.astraios.auth.exception.GrpcStatusException;
import com.astraios.grpc.auth.AuthServiceGrpc;
import com.astraios.grpc.auth.Jwk;
import com.astraios.grpc.auth.JwksResponse;
import com.astraios.grpc.auth.LoginResponse;
import com.astraios.grpc.auth.RegisterResponse;
import com.astraios.grpc.auth.RefreshTokenRequest;
import com.astraios.grpc.auth.RefreshTokenResponse;
import com.astraios.grpc.auth.Empty;
import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.server.service.GrpcService;

import java.util.List;
import java.util.Map;

/**
 * gRPC 认证服务实现
 * 提供登录、注册、刷新令牌等认证功能
 */
@Slf4j
@GrpcService
@RequiredArgsConstructor
public class AuthServiceGrpcImpl extends AuthServiceGrpc.AuthServiceImplBase {

    private final AuthService authService;
    private final JwtTokenProvider jwtTokenProvider;

    @Override
    public void login(com.astraios.grpc.auth.LoginRequest request, StreamObserver<LoginResponse> responseObserver) {
        try {
            log.info("收到登录请求: username={}", request.getUsername());

            if (request.getUsername().isBlank() || request.getPassword().isBlank()) {
                respondError(responseObserver, Status.INVALID_ARGUMENT, "username or password is blank", null);
                return;
            }
            
            // 转换gRPC请求为内部DTO
            LoginRequest loginRequest = new LoginRequest();
            loginRequest.setUsername(request.getUsername());
            loginRequest.setPassword(request.getPassword());
            loginRequest.setType(request.getType());
            
            // 调用业务服务
            LoginResult result = authService.login(loginRequest);
            
            // 构建gRPC响应
            LoginResponse response = LoginResponse.newBuilder()
                    .setAccessToken(result.getAccessToken() != null ? result.getAccessToken() : "")
                    .setRefreshToken(result.getRefreshToken() != null ? result.getRefreshToken() : "")
                    .build();
            
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("登录失败", e);
            handleException(responseObserver, e);
        }
    }

    @Override
    public void register(com.astraios.grpc.auth.RegisterRequest request, StreamObserver<RegisterResponse> responseObserver) {
        try {
            log.info("收到注册请求: username={}", request.getUsername());

            if (request.getUsername().isBlank() || request.getPassword().isBlank()) {
                respondError(responseObserver, Status.INVALID_ARGUMENT, "username or password is blank", null);
                return;
            }
            
            // 转换gRPC请求为内部DTO
            RegisterRequest registerRequest = new RegisterRequest();
            registerRequest.setUsername(request.getUsername());
            registerRequest.setPassword(request.getPassword());
            registerRequest.setType(request.getType());
            
            // 调用业务服务
            authService.register(registerRequest);
            
            // 构建gRPC响应
            RegisterResponse response = RegisterResponse.newBuilder()
                    .build();
            
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("注册失败", e);
            handleException(responseObserver, e);
        }
    }

    @Override
    public void refreshToken(RefreshTokenRequest request, StreamObserver<RefreshTokenResponse> responseObserver) {
        try {
            log.info("收到刷新令牌请求");

            if (request.getRefreshToken().isBlank()) {
                respondError(responseObserver, Status.INVALID_ARGUMENT, "refresh_token is blank", null);
                return;
            }
            
            // 转换gRPC请求为内部DTO
            RefreshRequest refreshRequest = new RefreshRequest();
            refreshRequest.setRefreshToken(request.getRefreshToken());
            
            // 调用业务服务
            RefreshResult result = authService.refreshToken(refreshRequest);
            
            // 构建gRPC响应
            RefreshTokenResponse response = RefreshTokenResponse.newBuilder()
                    .setAccessToken(result.getAccessToken() != null ? result.getAccessToken() : "")
                    .setRefreshToken(result.getRefreshToken() != null ? result.getRefreshToken() : "")
                    .build();
            
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("刷新令牌失败", e);
            handleException(responseObserver, e);
        }
    }

    @Override
    public void getJwks(Empty request, StreamObserver<JwksResponse> responseObserver) {
        try {
            Map<String, Object> jwkSet = jwtTokenProvider.getJwkSet();
            Object keysObject = jwkSet.get("keys");

            JwksResponse.Builder responseBuilder = JwksResponse.newBuilder();
            if (keysObject instanceof List<?> keys) {
                for (Object keyObject : keys) {
                    if (keyObject instanceof Map<?, ?> keyMap) {
                        Jwk.Builder jwkBuilder = Jwk.newBuilder();
                        setIfPresent(jwkBuilder::setKty, keyMap.get("kty"));
                        setIfPresent(jwkBuilder::setUse, keyMap.get("use"));
                        setIfPresent(jwkBuilder::setKid, keyMap.get("kid"));
                        setIfPresent(jwkBuilder::setAlg, keyMap.get("alg"));
                        setIfPresent(jwkBuilder::setN, keyMap.get("n"));
                        setIfPresent(jwkBuilder::setE, keyMap.get("e"));
                        responseBuilder.addKeys(jwkBuilder.build());
                    }
                }
            }

            responseObserver.onNext(responseBuilder.build());
            responseObserver.onCompleted();
        } catch (Exception e) {
            log.error("获取 JWKS 失败", e);
            responseObserver.onError(
                    Status.INTERNAL.withDescription("Failed to get JWKS").withCause(e).asRuntimeException()
            );
        }
    }

    private void setIfPresent(java.util.function.Consumer<String> setter, Object value) {
        if (value != null) {
            setter.accept(value.toString());
        }
    }

    private void handleException(StreamObserver<?> responseObserver, Exception e) {
        if (e instanceof GrpcStatusException grpcException) {
            respondError(responseObserver, grpcException.getStatus(), grpcException.getMessage(), grpcException);
            return;
        }
        if (e instanceof StatusRuntimeException statusRuntimeException) {
            responseObserver.onError(statusRuntimeException);
            return;
        }
        respondError(responseObserver, Status.INTERNAL, "request failed", e);
    }

    private void respondError(StreamObserver<?> responseObserver, Status status, String message, Throwable cause) {
        Status resultStatus = status.withDescription(message);
        if (cause != null) {
            resultStatus = resultStatus.withCause(cause);
        }
        responseObserver.onError(resultStatus.asRuntimeException());
    }
}
