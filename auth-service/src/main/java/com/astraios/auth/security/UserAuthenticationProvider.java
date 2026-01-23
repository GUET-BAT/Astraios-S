package com.astraios.auth.security;


import com.astraios.grpc.user.VerifyPasswordRequest;
import com.astraios.grpc.user.VerifyPasswordResponse;
import net.devh.boot.grpc.client.inject.GrpcClient;
import org.springframework.security.authentication.*;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.AuthenticationException;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.stereotype.Component;
import com.astraios.grpc.user.UserServiceGrpc;

import java.util.List;
import java.util.stream.Collectors;


@Component
public class UserAuthenticationProvider implements AuthenticationProvider {

    @GrpcClient("user-service")
    private UserServiceGrpc.UserServiceBlockingStub userServiceStub;

    @Override
    public Authentication authenticate(Authentication authentication) throws AuthenticationException {
        String username = authentication.getName();
        String password = authentication.getCredentials().toString();
        // 2. 构建请求
        VerifyPasswordRequest request = VerifyPasswordRequest.newBuilder()
                .setUsername(username)
                .setPassword(password)
                .build();

        try {
            // 3. 发起RPC调用
            VerifyPasswordResponse response = userServiceStub.verifyPassword(request);
            String userId = response.getUserId();
            if (response.getSuccess()) {
                // 登录成功
                List<SimpleGrantedAuthority> authorities = response.getRolesList().stream()
                        .map(SimpleGrantedAuthority::new)
                        .collect(Collectors.toList());
                return new UsernamePasswordAuthenticationToken(userId, null, authorities);
            } else {
                throw new BadCredentialsException("验证失败");
            }
        } catch (Exception e) {
            throw new AuthenticationServiceException("用户服务不可用");
        }
    }

    @Override
    public boolean supports(Class<?> authentication) {
        return UsernamePasswordAuthenticationToken.class.isAssignableFrom(authentication);
    }
}