package com.astraios.auth.controller;


import com.astraios.auth.domain.dto.LoginRequest;
import com.astraios.auth.domain.vo.LoginResult;
import com.astraios.auth.domain.dto.RegisterRequest;
import com.astraios.auth.service.AuthService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@RestController
@RequestMapping("/api")
@RequiredArgsConstructor
@Tag(name = "用户认证", description = "提供用户认证相关接口")
@Slf4j
public class AuthController {
    private final AuthService authService;

    @Operation(summary = "用户登录")
    @PostMapping("/login")
    public LoginResult login(@RequestBody LoginRequest loginRequest){
        return authService.login(loginRequest);
    }


    @Operation(summary = "用户注册")
    @PostMapping("/register")
    public ResponseEntity<?> register(@RequestBody RegisterRequest request){
        return authService.register(request);
    }

    @Operation(summary = "刷新令牌")
    @PostMapping("/refresh/token")
    public ResponseEntity<?> refresh(String refreshToken){
        return authService.refreshToken(refreshToken);
    }
}
