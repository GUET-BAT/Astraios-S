package com.astraios.auth.security;


import org.springframework.security.authentication.*;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.AuthenticationException;
import org.springframework.stereotype.Component;

@Component
public class UserAuthenticationProvider implements AuthenticationProvider {

//    @GrpcClient("userService")
//    private UserServiceGrpc.UserServiceBlockingStub userServiceStub;

    @Override
    public Authentication authenticate(Authentication authentication) throws AuthenticationException {
        String username = authentication.getName();
        String rawPassword = authentication.getCredentials().toString();

        // TODO RPC调用user服务验证密码
        String userId = username;
        return new UsernamePasswordAuthenticationToken(userId, null, null);
    }

    @Override
    public boolean supports(Class<?> authentication) {
        return UsernamePasswordAuthenticationToken.class.isAssignableFrom(authentication);
    }
}