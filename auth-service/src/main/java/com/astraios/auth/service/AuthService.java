package com.astraios.auth.service;

import com.astraios.auth.domain.dto.LoginRequest;
import com.astraios.auth.domain.dto.LoginResult;
import com.astraios.auth.domain.dto.RegisterRequest;
import com.astraios.auth.domain.dto.RegisterResult;
import org.springframework.http.ResponseEntity;


public interface AuthService {
    LoginResult login(LoginRequest loginRequest);

    ResponseEntity<RegisterResult> register(RegisterRequest request);
}
