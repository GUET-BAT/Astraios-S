package com.astraios.auth;

import com.astraios.grpc.auth.AuthServiceGrpc;
import com.astraios.grpc.auth.JwksResponse;
import com.astraios.grpc.auth.LoginRequest;
import com.astraios.grpc.auth.LoginResponse;
import com.astraios.grpc.auth.RefreshTokenRequest;
import com.astraios.grpc.auth.RefreshTokenResponse;
import com.astraios.grpc.auth.RegisterRequest;
import com.astraios.grpc.auth.RegisterResponse;
import com.astraios.grpc.user.UserDataRequest;
import com.astraios.grpc.user.UserDataResponse;
import com.astraios.grpc.user.UserServiceGrpc;
import com.astraios.grpc.user.VerifyPasswordRequest;
import com.astraios.grpc.user.VerifyPasswordResponse;
import com.google.protobuf.Empty;
import io.grpc.stub.StreamObserver;
import net.devh.boot.grpc.client.inject.GrpcClient;
import net.devh.boot.grpc.server.service.GrpcService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.context.annotation.Import;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.data.redis.core.ValueOperations;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.mockito.ArgumentMatchers.anyLong;
import static org.mockito.ArgumentMatchers.anyString;

@SpringBootTest(
        classes = AuthApplication.class,
        properties = {
                "grpc.server.inProcessName=auth-test",
                "grpc.server.port=-1",
                "grpc.client.auth.address=in-process:auth-test",
                "grpc.client.user-service.address=in-process:auth-test"
        }
)
@Import(AuthGrpcIntegrationTest.FakeUserService.class)
class AuthGrpcIntegrationTest {

    @GrpcClient("auth")
    private AuthServiceGrpc.AuthServiceBlockingStub authStub;

    @MockBean
    private StringRedisTemplate redisTemplate;

    private final Map<String, String> redisStore = new ConcurrentHashMap<>();
    private ValueOperations<String, String> valueOperations;

    @BeforeEach
    void setUp() {
        redisStore.clear();
        valueOperations = Mockito.mock(ValueOperations.class);
        Mockito.when(redisTemplate.opsForValue()).thenReturn(valueOperations);
        Mockito.doAnswer(invocation -> {
            String key = invocation.getArgument(0);
            String value = invocation.getArgument(1);
            redisStore.put(key, value);
            return null;
        }).when(valueOperations).set(anyString(), anyString(), anyLong(), Mockito.eq(TimeUnit.MILLISECONDS));
        Mockito.when(valueOperations.get(anyString()))
                .thenAnswer(invocation -> redisStore.get(invocation.getArgument(0)));
    }

    @Test
    void login_returns_tokens() {
        LoginResponse response = authStub.login(LoginRequest.newBuilder()
                .setUsername("demo")
                .setPassword("pass")
                .build());

        assertFalse(response.getAccessToken().isEmpty());
        assertFalse(response.getRefreshToken().isEmpty());
    }

    @Test
    void register_returns_success() {
        RegisterResponse response = authStub.register(RegisterRequest.newBuilder()
                .setUsername("demo")
                .setPassword("pass")
                .build());

        assertNotNull(response);
    }

    @Test
    void refresh_returns_new_tokens() {
        LoginResponse login = authStub.login(LoginRequest.newBuilder()
                .setUsername("demo")
                .setPassword("pass")
                .build());

        RefreshTokenResponse response = authStub.refreshToken(RefreshTokenRequest.newBuilder()
                .setRefreshToken(login.getRefreshToken())
                .build());

        assertNotNull(response.getAccessToken());
        assertNotNull(response.getRefreshToken());
        assertFalse(response.getAccessToken().isEmpty());
        assertFalse(response.getRefreshToken().isEmpty());
    }

    @Test
    void get_jwks_returns_keys() {
        JwksResponse response = authStub.getJwks(Empty.getDefaultInstance());
        assertFalse(response.getKeysList().isEmpty());
    }

    @GrpcService
    static class FakeUserService extends UserServiceGrpc.UserServiceImplBase {
        @Override
        public void verifyPassword(VerifyPasswordRequest request,
                                   StreamObserver<VerifyPasswordResponse> responseObserver) {
            boolean success = "demo".equals(request.getUsername()) && "pass".equals(request.getPassword());
            VerifyPasswordResponse response = VerifyPasswordResponse.newBuilder()
                    .setSuccess(success)
                    .setUserId(success ? "user-1" : "")
                    .addRoles("USER")
                    .build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }

        @Override
        public void register(com.astraios.grpc.user.RegisterRequest request,
                             StreamObserver<com.astraios.grpc.user.RegisterResponse> responseObserver) {
            com.astraios.grpc.user.RegisterResponse response = com.astraios.grpc.user.RegisterResponse.newBuilder()
                    .setCode(0)
                    .build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }

        @Override
        public void getUserId(UserDataRequest request, StreamObserver<UserDataResponse> responseObserver) {
            UserDataResponse response = UserDataResponse.newBuilder()
                    .setUserid("user-1")
                    .setUsername("demo")
                    .build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }
    }
}
