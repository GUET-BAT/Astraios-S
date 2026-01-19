package com.astraios.auth.service;

import com.astraios.auth.domain.dto.LoginRequest;
import com.astraios.auth.domain.vo.LoginResult;
import com.astraios.auth.domain.dto.RegisterRequest;
import com.astraios.auth.domain.vo.RefreshResult;
import com.astraios.auth.domain.vo.RegisterResult;
import org.springframework.http.ResponseEntity;

import java.util.Map;


public interface AuthService {
    LoginResult login(LoginRequest loginRequest);

    ResponseEntity<RegisterResult> register(RegisterRequest request);

    ResponseEntity<?> refreshToken(String refreshToken);
}
