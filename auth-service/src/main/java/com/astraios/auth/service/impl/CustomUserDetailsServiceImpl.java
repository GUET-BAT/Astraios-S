package com.astraios.auth.service.impl;


import lombok.RequiredArgsConstructor;
import org.springframework.security.core.GrantedAuthority;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.security.core.userdetails.UserDetailsService;
import org.springframework.stereotype.Service;

import java.util.Collection;
import java.util.List;

@Service
@RequiredArgsConstructor
public class CustomUserDetailsServiceImpl implements UserDetailsService {
    @Override
    public UserDetails loadUserByUsername(String userid){
        // TODO gPRC调用user服务获取用户信息
        String mockDbPassword = "$2a$10$t8q3WS8gMJ.N8wj9F5NlGuxzBCu8bAtSsSPJ6RGu.AIFenaVt67j."; //123
        return new UserDetails() {
            @Override
            public Collection<? extends GrantedAuthority> getAuthorities() {
                return null;
            }

            @Override
            public String getPassword() {
                return mockDbPassword;
            }

            @Override
            public String getUsername() {
                return userid;
            }
        };
    }
}